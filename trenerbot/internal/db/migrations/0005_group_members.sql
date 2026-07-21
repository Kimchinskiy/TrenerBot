-- Group members: many-to-many between clients and groups.
CREATE TABLE IF NOT EXISTS group_members (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id   INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    client_id  INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    role       TEXT NOT NULL DEFAULT 'member', -- member|assistant
    joined_at  TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(group_id, client_id)
);
CREATE INDEX IF NOT EXISTS idx_group_members_group ON group_members(group_id);
CREATE INDEX IF NOT EXISTS idx_group_members_client ON group_members(client_id);
