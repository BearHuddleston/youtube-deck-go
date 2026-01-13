package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"youtube-deck-go/internal/auth"
	"youtube-deck-go/internal/handlers"
	"youtube-deck-go/internal/middleware"
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

	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	database.SetConnMaxLifetime(0)

	if _, err := database.Exec(schema); err != nil {
		log.Fatalf("failed to create schema: %v", err)
	}

	for _, stmt := range migrations {
		if _, err := database.Exec(stmt); err != nil {
			if !isAlterTableDuplicate(err) {
				log.Printf("migration warning: %v", err)
			}
		}
	}

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

	h := handlers.New(database, ytClient, authMgr)

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
		authH := handlers.NewAuthHandlers(authMgr, database)
		mux.HandleFunc("GET /auth/login", authH.HandleLogin)
		mux.HandleFunc("GET /auth/callback", authH.HandleCallback)
		mux.HandleFunc("GET /auth/logout", authH.HandleLogout)
		mux.HandleFunc("POST /import", authH.HandleImportSubscriptions)
	}

	handler := middleware.CSRF(mux)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-done
	log.Println("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}

	if err := database.Close(); err != nil {
		log.Printf("database close error: %v", err)
	}

	log.Println("Server stopped")
}

// isAlterTableDuplicate checks if an ALTER TABLE error is due to a duplicate column.
// This uses string matching which is SQLite-specific. For production use with
// multiple database backends, consider using a proper migration library.
func isAlterTableDuplicate(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate column") || contains(err.Error(), "already exists"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
CREATE INDEX IF NOT EXISTS idx_videos_sub_watched_short ON videos(subscription_id, watched, is_short);
CREATE INDEX IF NOT EXISTS idx_subscriptions_active_position ON subscriptions(active, position);
`

var migrations = []string{
	"ALTER TABLE subscriptions ADD COLUMN position INTEGER DEFAULT 0",
	"ALTER TABLE subscriptions ADD COLUMN active INTEGER DEFAULT 0",
	"ALTER TABLE subscriptions ADD COLUMN page_token TEXT",
	"ALTER TABLE subscriptions ADD COLUMN hide_shorts INTEGER DEFAULT 0",
	"ALTER TABLE videos ADD COLUMN is_short INTEGER DEFAULT 0",
	"CREATE INDEX IF NOT EXISTS idx_videos_sub_watched_short ON videos(subscription_id, watched, is_short)",
}
