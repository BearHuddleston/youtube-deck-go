package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"youtube-deck-go/internal/db"
	"youtube-deck-go/internal/templates"
	"youtube-deck-go/internal/youtube"
)

func (h *Handlers) HandleAddSubscription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		YoutubeID    string `json:"youtube_id"`
		Name         string `json:"name"`
		Type         string `json:"type"`
		ThumbnailURL string `json:"thumbnail_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	maxPos, _ := h.queries.GetMaxPosition(r.Context())

	sub, err := h.queries.CreateSubscription(r.Context(), db.CreateSubscriptionParams{
		YoutubeID:    req.YoutubeID,
		Name:         req.Name,
		Type:         req.Type,
		ThumbnailUrl: sql.NullString{String: req.ThumbnailURL, Valid: req.ThumbnailURL != ""},
		Position:     sql.NullInt64{Int64: maxPos + 1, Valid: true},
		Active:       sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var unwatchedCount int64
	if req.Type == "channel" {
		vids, err := h.yt.FetchChannelVideos(r.Context(), req.YoutubeID, 20)
		if err == nil {
			if err := h.saveVideos(r, sub.ID, vids); err != nil {
				log.Printf("save videos error: %v", err)
			}
			unwatchedCount = int64(len(vids))
		}
	} else {
		vids, err := h.yt.FetchPlaylistVideos(r.Context(), req.YoutubeID, 20)
		if err == nil {
			if err := h.saveVideos(r, sub.ID, vids); err != nil {
				log.Printf("save videos error: %v", err)
			}
			unwatchedCount = int64(len(vids))
		}
	}
	if err := h.queries.UpdateSubscriptionChecked(r.Context(), sub.ID); err != nil {
		log.Printf("update subscription checked error: %v", err)
	}

	_ = templates.SidebarItem(templates.SubscriptionWithCount{
		Subscription:   sub,
		UnwatchedCount: unwatchedCount,
	}).Render(r.Context(), w)
}

func (h *Handlers) HandleDeleteSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.queries.DeleteSubscription(r.Context(), id); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) HandleRefreshSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	sub, err := h.queries.GetSubscription(r.Context(), id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if sub.Type == "channel" {
		vids, err := h.yt.FetchChannelVideos(r.Context(), sub.YoutubeID, 20)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err = h.saveVideos(r, sub.ID, vids); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	} else {
		vids, err := h.yt.FetchPlaylistVideos(r.Context(), sub.YoutubeID, 20)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err = h.saveVideos(r, sub.ID, vids); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	if err := h.queries.UpdateSubscriptionChecked(r.Context(), id); err != nil {
		log.Printf("update subscription checked error: %v", err)
	}

	hideShorts := int64(0)
	if sub.HideShorts.Valid {
		hideShorts = sub.HideShorts.Int64
	}

	unwatchedCount, err := h.queries.CountUnwatchedBySubscriptionFiltered(r.Context(), db.CountUnwatchedBySubscriptionFilteredParams{
		SubscriptionID: id,
		Column2:        hideShorts,
	})
	if err != nil {
		log.Printf("count unwatched error: %v", err)
	}

	swc := templates.SubscriptionWithCount{
		Subscription:   sub,
		UnwatchedCount: unwatchedCount,
	}

	if sub.Active.Valid && sub.Active.Int64 == 1 {
		videos, err := h.queries.ListUnwatchedVideosPaginatedFiltered(r.Context(), db.ListUnwatchedVideosPaginatedFilteredParams{
			SubscriptionID: id,
			Column2:        hideShorts,
			Limit:          11,
			Offset:         0,
		})
		if err != nil {
			log.Printf("list videos error: %v", err)
		}

		hasMoreDB := len(videos) > 10
		if hasMoreDB {
			videos = videos[:10]
		}

		canFetchMore := !sub.PageToken.Valid || sub.PageToken.String != ""
		_ = templates.ColumnWithVideos(swc, videos, hasMoreDB, canFetchMore, int64(len(videos))).Render(r.Context(), w)
	} else {
		_ = templates.SubscriptionCard(swc).Render(r.Context(), w)
	}
}

func (h *Handlers) saveVideos(r *http.Request, subID int64, vids []youtube.VideoInfo) error {
	// Filter to only new videos
	var newVids []youtube.VideoInfo
	for _, v := range vids {
		exists, _ := h.queries.VideoExistsByYoutubeID(r.Context(), v.ID)
		if exists == 0 {
			newVids = append(newVids, v)
		}
	}

	if len(newVids) == 0 {
		return nil
	}

	// Check shorts status in parallel for new videos
	newVids = h.yt.CheckShortsParallel(r.Context(), newVids)

	for _, v := range newVids {
		isShort := int64(0)
		if v.IsShort {
			isShort = 1
		}
		_, err := h.queries.CreateVideo(r.Context(), db.CreateVideoParams{
			SubscriptionID: subID,
			YoutubeID:      v.ID,
			Title:          v.Title,
			ThumbnailUrl:   sql.NullString{String: v.ThumbnailURL, Valid: v.ThumbnailURL != ""},
			Duration:       sql.NullString{String: v.Duration, Valid: v.Duration != ""},
			PublishedAt:    sql.NullTime{Time: v.PublishedAt, Valid: !v.PublishedAt.IsZero()},
			IsShort:        sql.NullInt64{Int64: isShort, Valid: true},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
