package handlers

import (
	"database/sql"
	"net/http"

	"youtube-deck-go/internal/auth"
	"youtube-deck-go/internal/db"
	"youtube-deck-go/internal/templates"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type AuthHandlers struct {
	auth    *auth.Manager
	queries *db.Queries
	state   string
}

func NewAuthHandlers(auth *auth.Manager, queries *db.Queries) *AuthHandlers {
	return &AuthHandlers{auth: auth, queries: queries}
}

func (h *AuthHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	url, state := h.auth.AuthURL()
	h.state = state

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandlers) HandleCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		return
	}

	if err := h.auth.Exchange(r.Context(), code); err != nil {
		http.Error(w, "Failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.auth.SaveToken("token.json")

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *AuthHandlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	h.auth.Logout()
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *AuthHandlers) HandleImportSubscriptions(w http.ResponseWriter, r *http.Request) {
	if !h.auth.IsAuthenticated() {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	token := h.auth.Token()
	client := h.auth.Config().Client(ctx, token)

	svc, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		templates.SearchError("Failed to create YouTube service: " + err.Error()).Render(ctx, w)
		return
	}

	var imported int
	pageToken := ""

	for {
		call := svc.Subscriptions.List([]string{"snippet"}).Mine(true).MaxResults(50)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			templates.SearchError("Failed to fetch subscriptions: " + err.Error()).Render(ctx, w)
			return
		}

		for _, item := range resp.Items {
			channelID := item.Snippet.ResourceId.ChannelId
			title := item.Snippet.Title
			thumbnail := getBestThumbnail(item.Snippet.Thumbnails)

			_, err := h.queries.CreateSubscription(ctx, db.CreateSubscriptionParams{
				Name:         title,
				YoutubeID:    channelID,
				Type:         "channel",
				ThumbnailUrl: sql.NullString{String: thumbnail, Valid: thumbnail != ""},
			})
			if err == nil {
				imported++
			}
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	subs, _ := h.queries.ListSubscriptionsWithUnwatchedCount(ctx)
	subsWithCount := make([]templates.SubscriptionWithCount, len(subs))
	for i, s := range subs {
		subsWithCount[i] = templates.SubscriptionWithCount{
			Subscription: db.Subscription{
				ID:           s.ID,
				Name:         s.Name,
				YoutubeID:    s.YoutubeID,
				Type:         s.Type,
				ThumbnailUrl: s.ThumbnailUrl,
				LastChecked:  s.LastChecked,
				CreatedAt:    s.CreatedAt,
			},
			UnwatchedCount: s.UnwatchedCount,
		}
	}

	w.Header().Set("HX-Trigger", `{"showToast": "Imported `+itoa(imported)+` subscriptions"}`)
	templates.SubscriptionGrid(subsWithCount).Render(ctx, w)
}

func (h *AuthHandlers) IsAuthenticated() bool {
	return h.auth.IsAuthenticated()
}

func getBestThumbnail(t *youtube.ThumbnailDetails) string {
	if t == nil {
		return ""
	}
	if t.Medium != nil {
		return t.Medium.Url
	}
	if t.High != nil {
		return t.High.Url
	}
	if t.Default != nil {
		return t.Default.Url
	}
	return ""
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
