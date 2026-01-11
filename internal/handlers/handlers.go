package handlers

import (
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
