CREATE TABLE lesson_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    time TEXT NOT NULL,
    client_id INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    coach_id INTEGER REFERENCES coaches(id),
    group_id INTEGER REFERENCES groups(id),
    duration INTEGER NOT NULL DEFAULT 60,
    status TEXT NOT NULL DEFAULT 'planned',
    comment TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_lesson_entries_date ON lesson_entries(date);
CREATE INDEX idx_lesson_entries_client ON lesson_entries(client_id);
CREATE INDEX idx_lesson_entries_coach ON lesson_entries(coach_id);
