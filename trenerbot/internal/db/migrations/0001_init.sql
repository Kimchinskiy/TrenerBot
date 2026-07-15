-- Users: one row per Telegram-linked account. role drives access (ТЗ §12).
CREATE TABLE IF NOT EXISTS users (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    telegram_id  TEXT UNIQUE,
    role         TEXT NOT NULL DEFAULT 'client', -- admin|coach|client|parent
    jwt_refresh  TEXT,
    created_at   TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_users_telegram ON users(telegram_id);

-- Clients (ТЗ §3.1)
CREATE TABLE IF NOT EXISTS clients (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    full_name       TEXT,
    photo           TEXT,
    birth_date      TEXT,
    age             INTEGER,
    phone           TEXT,
    telegram        TEXT,
    whatsapp        TEXT,
    parent_id       INTEGER REFERENCES users(id),
    second_parent_id INTEGER REFERENCES users(id),
    email           TEXT,
    medical_limits  TEXT,
    note            TEXT,
    status          TEXT DEFAULT 'active', -- active|paused|left
    registered_at   TEXT DEFAULT (datetime('now')),
    source          TEXT,
    bot_access      INTEGER DEFAULT 0,
    subscription_ends_at TEXT
);

-- Coaches (ТЗ §3.3)
CREATE TABLE IF NOT EXISTS coaches (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    full_name  TEXT,
    photo      TEXT,
    contacts   TEXT,
    position   TEXT,
    sport      TEXT,
    schedule   TEXT,
    group_ids  TEXT -- foundation: JSON array, not yet enforced
);

-- Groups (foundation only, ТЗ §3.2)
CREATE TABLE IF NOT EXISTS groups (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT,
    coach_id    INTEGER REFERENCES coaches(id),
    max_members INTEGER,
    schedule    TEXT,
    price       REAL,
    location    TEXT,
    active      INTEGER DEFAULT 1,
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Lessons (ТЗ §3.4). group_id is nullable foundation; participants live in attendance.
CREATE TABLE IF NOT EXISTS lessons (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    date       TEXT NOT NULL,            -- YYYY-MM-DD
    time       TEXT NOT NULL,            -- HH:MM
    coach_id   INTEGER REFERENCES coaches(id),
    duration   INTEGER DEFAULT 60,       -- minutes
    status     TEXT DEFAULT 'planned',   -- planned|ongoing|done|canceled|moved
    location   TEXT,
    comment    TEXT,
    group_id   INTEGER REFERENCES groups(id)
);
CREATE INDEX IF NOT EXISTS idx_lessons_date_coach ON lessons(date, coach_id);

-- Attendance: explicit participant list per lesson (ТЗ §3.5)
CREATE TABLE IF NOT EXISTS attendance (
    lesson_id INTEGER REFERENCES lessons(id) ON DELETE CASCADE,
    client_id INTEGER REFERENCES clients(id) ON DELETE CASCADE,
    present   INTEGER DEFAULT 0,
    marked_at TEXT,
    marked_by INTEGER REFERENCES users(id),
    PRIMARY KEY (lesson_id, client_id)
);
CREATE INDEX IF NOT EXISTS idx_attendance_client ON attendance(client_id);

-- Subscriptions (foundation, ТЗ §3.6/§3.8)
CREATE TABLE IF NOT EXISTS subscriptions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id   INTEGER REFERENCES clients(id),
    type        TEXT,   -- count|period|unlimited
    price       REAL,
    bought_at   TEXT,
    ends_at     TEXT,
    lessons_left INTEGER,
    freeze      INTEGER DEFAULT 0,
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Payments (foundation, ТЗ §3.7)
CREATE TABLE IF NOT EXISTS payments (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id  INTEGER REFERENCES clients(id),
    method     TEXT,
    amount     REAL,
    status     TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Notifications outbox (ТЗ §9). status + atomic UPDATE gives DB-level lock.
CREATE TABLE IF NOT EXISTS notifications (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    channel          TEXT DEFAULT 'telegram',
    recipient_user_id INTEGER REFERENCES users(id),
    type             TEXT,
    payload          TEXT, -- JSON
    send_at          TEXT NOT NULL,
    status           TEXT DEFAULT 'pending', -- pending|claimed|sent|failed
    claim_token      TEXT,
    sent_at          TEXT,
    created_at       TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_notifications_due ON notifications(status, send_at);

-- Files (ТЗ §7/§3.11)
CREATE TABLE IF NOT EXISTS files (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_type TEXT, -- client|lesson|coach
    owner_id   INTEGER,
    path       TEXT,
    kind       TEXT, -- photo|video
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Comments (ТЗ §3.11)
CREATE TABLE IF NOT EXISTS comments (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id  INTEGER REFERENCES clients(id),
    coach_id   INTEGER REFERENCES coaches(id),
    lesson_id  INTEGER REFERENCES lessons(id),
    text       TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Homework (ТЗ §3.11)
CREATE TABLE IF NOT EXISTS homework (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id  INTEGER REFERENCES clients(id),
    lesson_id  INTEGER REFERENCES lessons(id),
    text       TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Waiting list (ТЗ §3.9)
CREATE TABLE IF NOT EXISTS waiting_list (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id  INTEGER REFERENCES clients(id),
    group_id   INTEGER REFERENCES groups(id),
    position   INTEGER,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Activity log: client history (ТЗ §3.11)
CREATE TABLE IF NOT EXISTS activity_log (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id  INTEGER REFERENCES clients(id),
    type       TEXT, -- visit|payment|transfer|cancel|comment
    ref_id     INTEGER,
    note       TEXT,
    created_by INTEGER REFERENCES users(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Daily analytics snapshots for reports (ТЗ §11)
CREATE TABLE IF NOT EXISTS analytics_daily (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    day       TEXT NOT NULL, -- YYYY-MM-DD
    metric    TEXT NOT NULL,
    value     REAL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_analytics_day ON analytics_daily(day);
