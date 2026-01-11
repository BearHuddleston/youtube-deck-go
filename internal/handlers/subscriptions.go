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

	sub, err := h.queries.CreateSubscription(r.Context(), db.CreateSubscriptionParams{
		YoutubeID:    req.YoutubeID,
		Name:         req.Name,
		Type:         req.Type,
		ThumbnailUrl: sql.NullString{String: req.ThumbnailURL, Valid: req.ThumbnailURL != ""},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.SubscriptionCard(templates.SubscriptionWithCount{
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

	var videos []db.Video
	if sub.Type == "channel" {
		vids, err := h.yt.FetchChannelVideos(r.Context(), sub.YoutubeID, 20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		videos, err = h.saveVideos(r, sub.ID, vids)
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
		videos, err = h.saveVideos(r, sub.ID, vids)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.queries.UpdateSubscriptionChecked(r.Context(), id)

	unwatchedCount := int64(0)
	for _, v := range videos {
		if !v.Watched.Valid || v.Watched.Int64 == 0 {
			unwatchedCount++
		}
	}

	templates.SubscriptionCard(templates.SubscriptionWithCount{
		Subscription:   sub,
		UnwatchedCount: unwatchedCount,
	}).Render(r.Context(), w)
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
