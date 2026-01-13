package handlers

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"

	"youtube-deck-go/internal/auth"
	"youtube-deck-go/internal/db"
	"youtube-deck-go/internal/youtube"
)

type Handlers struct {
	queries *db.Queries
	db      *sql.DB
	yt      *youtube.Client
	auth    *auth.Manager
	log     *slog.Logger
}

func New(database *sql.DB, yt *youtube.Client, authMgr *auth.Manager, log *slog.Logger) *Handlers {
	return &Handlers{
		queries: db.New(database),
		db:      database,
		yt:      yt,
		auth:    authMgr,
		log:     log,
	}
}

func (h *Handlers) render(w http.ResponseWriter, ctx context.Context, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := c.Render(ctx, w); err != nil {
		h.log.Error("render error", "error", err)
	}
}
