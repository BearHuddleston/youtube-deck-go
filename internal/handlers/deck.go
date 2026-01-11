package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"youtube-deck-go/internal/db"
	"youtube-deck-go/internal/templates"
)

func (h *Handlers) HandleDeck(w http.ResponseWriter, r *http.Request) {
	allRows, err := h.queries.ListAllSubscriptionsOrdered(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	activeRows, err := h.queries.ListActiveSubscriptions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	allSubs := make([]templates.SubscriptionWithCount, len(allRows))
	for i, row := range allRows {
		allSubs[i] = templates.SubscriptionWithCount{
			Subscription: db.Subscription{
				ID:           row.ID,
				Name:         row.Name,
				YoutubeID:    row.YoutubeID,
				Type:         row.Type,
				ThumbnailUrl: row.ThumbnailUrl,
				LastChecked:  row.LastChecked,
				CreatedAt:    row.CreatedAt,
				Position:     row.Position,
				Active:       row.Active,
			},
			UnwatchedCount: row.UnwatchedCount,
		}
	}

	activeSubs := make([]templates.SubscriptionWithCount, len(activeRows))
	for i, row := range activeRows {
		activeSubs[i] = templates.SubscriptionWithCount{
			Subscription: db.Subscription{
				ID:           row.ID,
				Name:         row.Name,
				YoutubeID:    row.YoutubeID,
				Type:         row.Type,
				ThumbnailUrl: row.ThumbnailUrl,
				LastChecked:  row.LastChecked,
				CreatedAt:    row.CreatedAt,
				Position:     row.Position,
				Active:       row.Active,
			},
			UnwatchedCount: row.UnwatchedCount,
		}
	}

	isAuth := h.auth != nil && h.auth.IsAuthenticated()
	templates.Deck(allSubs, activeSubs, isAuth).Render(r.Context(), w)
}

func (h *Handlers) HandleColumnVideos(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	videos, err := h.queries.ListUnwatchedVideos(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.ColumnVideos(videos).Render(r.Context(), w)
}

func (h *Handlers) HandleToggleActive(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	activeParam := r.URL.Query().Get("active")
	active, _ := strconv.ParseInt(activeParam, 10, 64)

	err = h.queries.UpdateSubscriptionActive(r.Context(), db.UpdateSubscriptionActiveParams{
		Active: sql.NullInt64{Int64: active, Valid: true},
		ID:     id,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if active == 1 {
		sub, err := h.queries.GetSubscription(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		count, _ := h.queries.CountUnwatchedBySubscription(r.Context(), id)
		templates.Column(templates.SubscriptionWithCount{
			Subscription:   sub,
			UnwatchedCount: count,
		}).Render(r.Context(), w)
	} else {
		rows, err := h.queries.ListAllSubscriptionsOrdered(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, row := range rows {
			if row.ID == id {
				templates.SidebarItem(templates.SubscriptionWithCount{
					Subscription: db.Subscription{
						ID:           row.ID,
						Name:         row.Name,
						YoutubeID:    row.YoutubeID,
						Type:         row.Type,
						ThumbnailUrl: row.ThumbnailUrl,
						LastChecked:  row.LastChecked,
						CreatedAt:    row.CreatedAt,
						Position:     row.Position,
						Active:       row.Active,
					},
					UnwatchedCount: row.UnwatchedCount,
				}).Render(r.Context(), w)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handlers) HandleReorder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs     []string `json:"ids"`
		Context string   `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i, idStr := range req.IDs {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}
		h.queries.UpdateSubscriptionPosition(r.Context(), db.UpdateSubscriptionPositionParams{
			Position: sql.NullInt64{Int64: int64(i), Valid: true},
			ID:       id,
		})
	}

	w.WriteHeader(http.StatusOK)
}
