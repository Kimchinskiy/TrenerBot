CREATE TABLE IF NOT EXISTS coach_subscriptions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    coach_id    INTEGER NOT NULL REFERENCES coaches(id) ON DELETE CASCADE,
    status      TEXT NOT NULL DEFAULT 'trial',
    trial_start TEXT NOT NULL DEFAULT (datetime('now')),
    trial_end   TEXT,
    paid_until  TEXT,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT
);

CREATE TABLE IF NOT EXISTS parent_notification_prefs (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    child_id      INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    lesson_start  INTEGER NOT NULL DEFAULT 1,
    lesson_end_15 INTEGER NOT NULL DEFAULT 1,
    lesson_missed INTEGER NOT NULL DEFAULT 1,
    created_at    TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(parent_user_id, child_id)
);
