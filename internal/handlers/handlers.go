package handlers

import (
	"database/sql"

	"youtube-deck-go/internal/auth"
	"youtube-deck-go/internal/db"
	"youtube-deck-go/internal/youtube"
)

type Handlers struct {
	queries *db.Queries
	db      *sql.DB
	yt      *youtube.Client
	auth    *auth.Manager
}

func New(database *sql.DB, yt *youtube.Client, authMgr *auth.Manager) *Handlers {
	return &Handlers{
		queries: db.New(database),
		db:      database,
		yt:      yt,
		auth:    authMgr,
	}
}
