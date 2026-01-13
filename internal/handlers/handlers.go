package handlers

import (
	"context"
	"net/http"

	"github.com/a-h/templ"

	"youtube-deck-go/internal/auth"
	"youtube-deck-go/internal/db"
	"youtube-deck-go/internal/youtube"
)

func render(w http.ResponseWriter, ctx context.Context, c templ.Component) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return c.Render(ctx, w)
}

type Handlers struct {
	queries *db.Queries
	yt      *youtube.Client
	auth    *auth.Manager
}

func New(queries *db.Queries, yt *youtube.Client, authMgr *auth.Manager) *Handlers {
	return &Handlers{
		queries: queries,
		yt:      yt,
		auth:    authMgr,
	}
}
