package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"youtube-deck-go/internal/auth"
	"youtube-deck-go/internal/db"
	"youtube-deck-go/internal/handlers"
	"youtube-deck-go/internal/youtube"

	_ "modernc.org/sqlite"
)

func main() {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		log.Fatal("YOUTUBE_API_KEY environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data.db"
	}

	clientSecretFile := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientSecretFile == "" {
		clientSecretFile = "client_secret.json"
	}

	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if _, err := database.Exec(schema); err != nil {
		log.Fatalf("failed to create schema: %v", err)
	}

	for _, stmt := range migrations {
		_, _ = database.Exec(stmt)
	}

	queries := db.New(database)

	ytClient, err := youtube.New(apiKey)
	if err != nil {
		log.Fatalf("failed to create youtube client: %v", err)
	}

	var authMgr *auth.Manager
	if _, err := os.Stat(clientSecretFile); err == nil {
		authMgr, err = auth.NewManager(clientSecretFile)
		if err != nil {
			log.Printf("Warning: failed to init OAuth: %v", err)
		} else {
			_ = authMgr.LoadToken("token.json")
			log.Printf("OAuth enabled")
		}
	} else {
		log.Printf("OAuth disabled (no client_secret.json found)")
	}

	h := handlers.New(queries, ytClient, authMgr)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.HandleDeck)
	mux.HandleFunc("GET /search", h.HandleSearch)
	mux.HandleFunc("GET /search/results", h.HandleSearchResults)
	mux.HandleFunc("GET /search/close", h.HandleSearchClose)
	mux.HandleFunc("POST /subscriptions", h.HandleAddSubscription)
	mux.HandleFunc("GET /subscriptions/filter", h.HandleFilterSubscriptions)
	mux.HandleFunc("POST /subscriptions/reorder", h.HandleReorder)
	mux.HandleFunc("DELETE /subscriptions/{id}", h.HandleDeleteSubscription)
	mux.HandleFunc("GET /subscriptions/{id}/videos", h.HandleVideos)
	mux.HandleFunc("GET /subscriptions/{id}/column", h.HandleColumnVideos)
	mux.HandleFunc("POST /subscriptions/{id}/refresh", h.HandleRefreshSubscription)
	mux.HandleFunc("POST /subscriptions/{id}/fetch-more", h.HandleFetchMoreVideos)
	mux.HandleFunc("PATCH /subscriptions/{id}/active", h.HandleToggleActive)
	mux.HandleFunc("PATCH /subscriptions/{id}/hide-shorts", h.HandleToggleHideShorts)
	mux.HandleFunc("POST /videos/{id}/watched", h.HandleToggleWatched)
	mux.HandleFunc("GET /proxy/image", h.HandleImageProxy)

	if authMgr != nil {
		authH := handlers.NewAuthHandlers(authMgr, queries)
		mux.HandleFunc("GET /auth/login", authH.HandleLogin)
		mux.HandleFunc("GET /auth/callback", authH.HandleCallback)
		mux.HandleFunc("GET /auth/logout", authH.HandleLogout)
		mux.HandleFunc("POST /import", authH.HandleImportSubscriptions)
	}

	log.Printf("Server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

const schema = `
CREATE TABLE IF NOT EXISTS subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    youtube_id TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL CHECK(type IN ('channel', 'playlist')),
    thumbnail_url TEXT,
    last_checked DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    position INTEGER DEFAULT 0,
    active INTEGER DEFAULT 0,
    page_token TEXT,
    hide_shorts INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS videos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subscription_id INTEGER NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    youtube_id TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    thumbnail_url TEXT,
    duration TEXT,
    published_at DATETIME,
    watched INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_short INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_videos_subscription ON videos(subscription_id);
CREATE INDEX IF NOT EXISTS idx_videos_watched ON videos(watched);
`

var migrations = []string{
	"ALTER TABLE subscriptions ADD COLUMN position INTEGER DEFAULT 0",
	"ALTER TABLE subscriptions ADD COLUMN active INTEGER DEFAULT 0",
	"ALTER TABLE subscriptions ADD COLUMN page_token TEXT",
	"ALTER TABLE subscriptions ADD COLUMN hide_shorts INTEGER DEFAULT 0",
	"ALTER TABLE videos ADD COLUMN is_short INTEGER DEFAULT 0",
}
