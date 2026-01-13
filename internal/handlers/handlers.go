package handlers

import (
	"database/sql"
	"log/slog"

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
