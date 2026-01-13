package handlers

import (
	"net/http"

	"youtube-deck-go/internal/db"
	"youtube-deck-go/internal/templates"
)

func (h *Handlers) HandleHome(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.ListSubscriptionsWithUnwatchedCount(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	subs := make([]templates.SubscriptionWithCount, len(rows))
	for i, row := range rows {
		subs[i] = templates.SubscriptionWithCount{
			Subscription: db.Subscription{
				ID:           row.ID,
				Name:         row.Name,
				YoutubeID:    row.YoutubeID,
				Type:         row.Type,
				ThumbnailUrl: row.ThumbnailUrl,
				LastChecked:  row.LastChecked,
				CreatedAt:    row.CreatedAt,
			},
			UnwatchedCount: row.UnwatchedCount,
		}
	}

	isAuth := h.auth != nil && h.auth.IsAuthenticated()
	_ = templates.Home(subs, isAuth).Render(r.Context(), w)
}
