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

func (h *Handlers) HandleDeck(w http.ResponseWriter, r *http.Request) {
	sidebarRows, err := h.queries.ListSubscriptionsPaginated(r.Context(), db.ListSubscriptionsPaginatedParams{
		Limit:  sidebarPageSize + 1,
		Offset: 0,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hasMore := len(sidebarRows) > sidebarPageSize
	if hasMore {
		sidebarRows = sidebarRows[:sidebarPageSize]
	}

	activeRows, err := h.queries.ListActiveSubscriptions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sidebarSubs := h.rowsToSubs(sidebarRows)
	activeSubs := h.activeRowsToSubs(activeRows)
	nextOffset := int64(len(sidebarSubs))

	isAuth := h.auth != nil && h.auth.IsAuthenticated()
	_ = templates.Deck(sidebarSubs, activeSubs, hasMore, nextOffset, isAuth).Render(r.Context(), w)
}

func (h *Handlers) activeRowsToSubs(rows []db.ListActiveSubscriptionsRow) []templates.SubscriptionWithCount {
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
				Position:     row.Position,
				Active:       row.Active,
			},
			UnwatchedCount: row.UnwatchedCount,
		}
	}
	return subs
}

const columnVideoPageSize = 10

func (h *Handlers) HandleColumnVideos(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	offsetStr := r.URL.Query().Get("offset")
	offset, _ := strconv.ParseInt(offsetStr, 10, 64)

	videos, err := h.queries.ListUnwatchedVideosPaginated(r.Context(), db.ListUnwatchedVideosPaginatedParams{
		SubscriptionID: id,
		Limit:          columnVideoPageSize + 1,
		Offset:         offset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hasMoreDB := len(videos) > columnVideoPageSize
	if hasMoreDB {
		videos = videos[:columnVideoPageSize]
	}

	canFetchMore := false
	if !hasMoreDB {
		sub, err := h.queries.GetSubscription(r.Context(), id)
		if err == nil {
			// Allow fetching more if we have a next page token OR if we haven't fetched yet (NULL token)
			// Empty string "" means we've exhausted all pages
			canFetchMore = !sub.PageToken.Valid || sub.PageToken.String != ""
		}
	}

	nextOffset := offset + int64(len(videos))
	_ = templates.ColumnVideos(videos, id, hasMoreDB, canFetchMore, nextOffset).Render(r.Context(), w)
}

func (h *Handlers) HandleFetchMoreVideos(w http.ResponseWriter, r *http.Request) {
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

	// Count existing unwatched videos before fetching more
	existingCount, _ := h.queries.CountUnwatchedBySubscription(r.Context(), id)

	pageToken := ""
	if sub.PageToken.Valid {
		pageToken = sub.PageToken.String
	}

	var result *youtube.FetchResult
	if sub.Type == "channel" {
		result, err = h.yt.FetchChannelVideosWithToken(r.Context(), sub.YoutubeID, pageToken, 20)
	} else {
		result, err = h.yt.FetchPlaylistVideosWithToken(r.Context(), sub.YoutubeID, pageToken, 20)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = h.saveVideos(r, sub.ID, result.Videos)

	newToken := sql.NullString{String: result.NextPageToken, Valid: result.NextPageToken != ""}
	_ = h.queries.UpdateSubscriptionPageToken(r.Context(), db.UpdateSubscriptionPageTokenParams{
		PageToken: newToken,
		ID:        sub.ID,
	})

	// Query starting from where we left off (after existing videos)
	videos, _ := h.queries.ListUnwatchedVideosPaginated(r.Context(), db.ListUnwatchedVideosPaginatedParams{
		SubscriptionID: id,
		Limit:          columnVideoPageSize + 1,
		Offset:         existingCount,
	})

	hasMoreDB := len(videos) > columnVideoPageSize
	if hasMoreDB {
		videos = videos[:columnVideoPageSize]
	}

	canFetchMore := result.NextPageToken != ""

	nextOffset := existingCount + int64(len(videos))
	_ = templates.ColumnVideos(videos, id, hasMoreDB, canFetchMore, nextOffset).Render(r.Context(), w)
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

		// Fetch videos from YouTube when adding to deck
		var result *youtube.FetchResult
		if sub.Type == "channel" {
			result, err = h.yt.FetchChannelVideosWithToken(r.Context(), sub.YoutubeID, "", 20)
		} else {
			result, err = h.yt.FetchPlaylistVideosWithToken(r.Context(), sub.YoutubeID, "", 20)
		}
		if err == nil {
			_ = h.saveVideos(r, sub.ID, result.Videos)
			newToken := sql.NullString{String: result.NextPageToken, Valid: result.NextPageToken != ""}
			_ = h.queries.UpdateSubscriptionPageToken(r.Context(), db.UpdateSubscriptionPageTokenParams{
				PageToken: newToken,
				ID:        sub.ID,
			})
		}
		_ = h.queries.UpdateSubscriptionChecked(r.Context(), id)

		activeCount, _ := h.queries.CountActiveSubscriptions(r.Context())
		count, _ := h.queries.CountUnwatchedBySubscription(r.Context(), id)
		_ = templates.ColumnWithChip(templates.SubscriptionWithCount{
			Subscription:   sub,
			UnwatchedCount: count,
		}, activeCount == 1).Render(r.Context(), w)
	} else {
		rows, err := h.queries.ListAllSubscriptionsOrdered(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, row := range rows {
			if row.ID == id {
				_ = templates.SidebarItem(templates.SubscriptionWithCount{
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

const sidebarPageSize = 50

func (h *Handlers) HandleFilterSubscriptions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	offsetStr := r.URL.Query().Get("offset")
	offset, _ := strconv.ParseInt(offsetStr, 10, 64)

	if q != "" {
		rows, err := h.queries.FilterSubscriptions(r.Context(), sql.NullString{String: q, Valid: true})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		subs := h.filterRowsToSubs(rows)
		_ = templates.SidebarList(subs, false, 0).Render(r.Context(), w)
		return
	}

	rows, err := h.queries.ListSubscriptionsPaginated(r.Context(), db.ListSubscriptionsPaginatedParams{
		Limit:  sidebarPageSize + 1,
		Offset: offset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hasMore := len(rows) > sidebarPageSize
	if hasMore {
		rows = rows[:sidebarPageSize]
	}

	subs := h.rowsToSubs(rows)
	nextOffset := offset + int64(len(subs))
	_ = templates.SidebarList(subs, hasMore, nextOffset).Render(r.Context(), w)
}

func (h *Handlers) rowsToSubs(rows []db.ListSubscriptionsPaginatedRow) []templates.SubscriptionWithCount {
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
				Position:     row.Position,
				Active:       row.Active,
			},
			UnwatchedCount: row.UnwatchedCount,
		}
	}
	return subs
}

func (h *Handlers) filterRowsToSubs(rows []db.FilterSubscriptionsRow) []templates.SubscriptionWithCount {
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
				Position:     row.Position,
				Active:       row.Active,
			},
			UnwatchedCount: row.UnwatchedCount,
		}
	}
	return subs
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
		_ = h.queries.UpdateSubscriptionPosition(r.Context(), db.UpdateSubscriptionPositionParams{
			Position: sql.NullInt64{Int64: int64(i), Valid: true},
			ID:       id,
		})
	}

	w.WriteHeader(http.StatusOK)
}
