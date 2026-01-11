package handlers

import (
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
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	videos, err := h.queries.ListVideos(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.Videos(sub, videos).Render(r.Context(), w)
}

func (h *Handlers) HandleToggleWatched(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	video, err := h.queries.ToggleWatched(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.VideoCard(video).Render(r.Context(), w)
}
