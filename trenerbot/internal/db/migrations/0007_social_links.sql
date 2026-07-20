CREATE TABLE IF NOT EXISTS coach_social_links (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    coach_id   INTEGER NOT NULL REFERENCES coaches(id) ON DELETE CASCADE,
    platform   TEXT NOT NULL,
    url        TEXT,
    enabled    INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT,
    UNIQUE(coach_id, platform)
);
