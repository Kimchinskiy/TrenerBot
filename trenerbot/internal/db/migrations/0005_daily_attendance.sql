CREATE TABLE IF NOT EXISTS daily_attendance (
    date TEXT NOT NULL,
    client_id INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    present INTEGER NOT NULL DEFAULT 0,
    marked_by INTEGER REFERENCES users(id),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (date, client_id)
);
CREATE INDEX IF NOT EXISTS idx_daily_attendance_client ON daily_attendance(client_id);
