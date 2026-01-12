CREATE TABLE subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    youtube_id TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL CHECK(type IN ('channel', 'playlist')),
    thumbnail_url TEXT,
    last_checked DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    position INTEGER DEFAULT 0,
    active INTEGER DEFAULT 0,
    page_token TEXT
);

CREATE TABLE videos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subscription_id INTEGER NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    youtube_id TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    thumbnail_url TEXT,
    duration TEXT,
    published_at DATETIME,
    watched INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_videos_subscription ON videos(subscription_id);
CREATE INDEX idx_videos_watched ON videos(watched);
