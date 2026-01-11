package handlers

import (
	"net/http"

	"youtube-deck-go/internal/templates"
)

func (h *Handlers) HandleSearch(w http.ResponseWriter, r *http.Request) {
	templates.SearchModal().Render(r.Context(), w)
}

func (h *Handlers) HandleSearchResults(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		w.Write([]byte(""))
		return
	}

	searchType := r.URL.Query().Get("type")
	if searchType == "" {
		searchType = "channel"
	}

	if searchType == "channel" {
		results, err := h.yt.SearchChannels(r.Context(), query, 10)
		if err != nil {
			templates.SearchError(err.Error()).Render(r.Context(), w)
			return
		}
		templates.SearchResults(results).Render(r.Context(), w)
	} else {
		results, err := h.yt.SearchPlaylists(r.Context(), query, 10)
		if err != nil {
			templates.SearchError(err.Error()).Render(r.Context(), w)
			return
		}
		templates.SearchResults(results).Render(r.Context(), w)
	}
}

func (h *Handlers) HandleSearchClose(w http.ResponseWriter, r *http.Request) {
	templates.SearchClose().Render(r.Context(), w)
}
