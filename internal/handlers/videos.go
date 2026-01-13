package handlers

import (
	"log"
	"net/http"
	"strconv"

	"youtube-deck-go/internal/templates"
)

func (h *Handlers) HandleVideos(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	sub, err := h.queries.GetSubscription(r.Context(), id)
	if err != nil {
		http.Error(w, "subscription not found", http.StatusNotFound)
		return
	}

	videos, err := h.queries.ListVideos(r.Context(), id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = templates.Videos(sub, videos).Render(r.Context(), w)
}

func (h *Handlers) HandleToggleWatched(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	video, err := h.queries.GetVideo(r.Context(), id)
	if err != nil {
		http.Error(w, "video not found", http.StatusNotFound)
		return
	}

	if err := h.queries.MarkWatched(r.Context(), id); err != nil {
		log.Printf("failed to mark video %d as watched: %v", id, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	count, err := h.queries.CountUnwatchedBySubscription(r.Context(), video.SubscriptionID)
	if err != nil {
		log.Printf("failed to count unwatched videos for subscription %d: %v", video.SubscriptionID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = templates.UnwatchedCountsOOB(video.SubscriptionID, count).Render(r.Context(), w)
}
