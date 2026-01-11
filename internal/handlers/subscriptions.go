package handlers

import (
	"database/sql"
	"encoding/json"
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	maxPosRaw, _ := h.queries.GetMaxPosition(r.Context())
	var maxPos int64
	if v, ok := maxPosRaw.(int64); ok {
		maxPos = v
	}

	sub, err := h.queries.CreateSubscription(r.Context(), db.CreateSubscriptionParams{
		YoutubeID:    req.YoutubeID,
		Name:         req.Name,
		Type:         req.Type,
		ThumbnailUrl: sql.NullString{String: req.ThumbnailURL, Valid: req.ThumbnailURL != ""},
		Position:     sql.NullInt64{Int64: maxPos + 1, Valid: true},
		Active:       sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.SidebarItem(templates.SubscriptionWithCount{
		Subscription:   sub,
		UnwatchedCount: 0,
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sub.Type == "channel" {
		vids, err := h.yt.FetchChannelVideos(r.Context(), sub.YoutubeID, 20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = h.saveVideos(r, sub.ID, vids)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		vids, err := h.yt.FetchPlaylistVideos(r.Context(), sub.YoutubeID, 20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = h.saveVideos(r, sub.ID, vids)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.queries.UpdateSubscriptionChecked(r.Context(), id)

	unwatchedCount, _ := h.queries.CountUnwatchedBySubscription(r.Context(), id)

	swc := templates.SubscriptionWithCount{
		Subscription:   sub,
		UnwatchedCount: unwatchedCount,
	}

	if sub.Active.Valid && sub.Active.Int64 == 1 {
		templates.Column(swc).Render(r.Context(), w)
	} else {
		templates.SubscriptionCard(swc).Render(r.Context(), w)
	}
}

func (h *Handlers) saveVideos(r *http.Request, subID int64, vids []youtube.VideoInfo) ([]db.Video, error) {
	var saved []db.Video
	for _, v := range vids {
		exists, _ := h.queries.VideoExistsByYoutubeID(r.Context(), v.ID)
		if exists == 1 {
			continue
		}
		video, err := h.queries.CreateVideo(r.Context(), db.CreateVideoParams{
			SubscriptionID: subID,
			YoutubeID:      v.ID,
			Title:          v.Title,
			ThumbnailUrl:   sql.NullString{String: v.ThumbnailURL, Valid: v.ThumbnailURL != ""},
			Duration:       sql.NullString{String: v.Duration, Valid: v.Duration != ""},
			PublishedAt:    sql.NullTime{Time: v.PublishedAt, Valid: !v.PublishedAt.IsZero()},
		})
		if err != nil {
			return nil, err
		}
		saved = append(saved, video)
	}
	return saved, nil
}
