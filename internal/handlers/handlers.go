package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/a-h/templ"

	"youtube-deck-go/internal/auth"
	"youtube-deck-go/internal/db"
	"youtube-deck-go/internal/youtube"
)

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

func (h *Handlers) render(w http.ResponseWriter, ctx context.Context, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := c.Render(ctx, w); err != nil {
		log.Printf("render error: %v", err)
	}
}
