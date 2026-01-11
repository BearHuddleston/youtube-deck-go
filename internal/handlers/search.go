package handlers

import (
	"net/http"

	"youtube-deck-go/internal/templates"
)

func (h *Handlers) HandleSearch(w http.ResponseWriter, r *http.Request) {
	_ = templates.SearchModal().Render(r.Context(), w)
}

func (h *Handlers) HandleSearchResults(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		_, _ = w.Write([]byte(""))
		return
	}

	searchType := r.URL.Query().Get("type")
	if searchType == "" {
		searchType = "channel"
	}

	if searchType == "channel" {
		results, err := h.yt.SearchChannels(r.Context(), query, 10)
		if err != nil {
			_ = templates.SearchError(err.Error()).Render(r.Context(), w)
			return
		}
		_ = templates.SearchResults(results).Render(r.Context(), w)
	} else {
		results, err := h.yt.SearchPlaylists(r.Context(), query, 10)
		if err != nil {
			_ = templates.SearchError(err.Error()).Render(r.Context(), w)
			return
		}
		_ = templates.SearchResults(results).Render(r.Context(), w)
	}
}

func (h *Handlers) HandleSearchClose(w http.ResponseWriter, r *http.Request) {
	_ = templates.SearchClose().Render(r.Context(), w)
}
