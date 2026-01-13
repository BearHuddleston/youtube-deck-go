package handlers

import (
	"log/slog"

	"youtube-deck-go/internal/auth"
	"youtube-deck-go/internal/db"
	"youtube-deck-go/internal/youtube"
)

type Handlers struct {
	queries *db.Queries
	yt      *youtube.Client
	auth    *auth.Manager
	log     *slog.Logger
}

func New(queries *db.Queries, yt *youtube.Client, authMgr *auth.Manager, log *slog.Logger) *Handlers {
	return &Handlers{
		queries: queries,
		yt:      yt,
		auth:    authMgr,
		log:     log,
	}
}
