-- New concept: the website is the primary product. Users are a single entity that
-- may authenticate via phone+password, Telegram, or MAX. telegram_id стоп быть
-- первичным идентификатором — основным становится users.id.

-- Extend the users table with account fields (was: id, telegram_id, role, jwt_refresh, created_at).
ALTER TABLE users ADD COLUMN phone         TEXT;
ALTER TABLE users ADD COLUMN password_hash TEXT;
ALTER TABLE users ADD COLUMN max_id        TEXT;
ALTER TABLE users ADD COLUMN first_name    TEXT;
ALTER TABLE users ADD COLUMN last_name     TEXT;
ALTER TABLE users ADD COLUMN avatar_url    TEXT;
ALTER TABLE users ADD COLUMN updated_at    TEXT;

-- Backfill name from the linked client profile.
UPDATE users
SET first_name = (
    SELECT c.full_name FROM clients c
    WHERE c.user_id = users.id AND c.full_name IS NOT NULL AND c.full_name <> ''
    LIMIT 1
)
WHERE first_name IS NULL;

-- Backfill phone from the linked client profile, applying the SAME canonicalization
-- used by the Go normalizePhone helper (strip spaces/parens/dashes; leading 8 -> +7;
-- otherwise prefix +). Only assign when the canonical phone is NOT already taken by
-- another user, so the partial unique index created below cannot fail on shared numbers
-- (e.g. a parent phone reused across children). IMPORTANT: this runs BEFORE the unique
-- index so a collision can never abort the migration.
WITH normalized AS (
    SELECT
        c.user_id AS uid,
        CASE
            WHEN length(replace(replace(replace(replace(replace(c.phone,' ',''),'(',''),')',''),'-',''),'+','')) = 11
                 AND substr(replace(replace(replace(replace(replace(c.phone,' ',''),'(',''),')',''),'-',''),'+',''),1,1) = '8'
            THEN '+7' || substr(replace(replace(replace(replace(replace(c.phone,' ',''),'(',''),')',''),'-',''),'+',''),2)
            ELSE '+' || replace(replace(replace(replace(replace(c.phone,' ',''),'(',''),')',''),'-',''),'+','')
        END AS canon
    FROM clients c
    WHERE c.phone IS NOT NULL AND c.phone <> ''
)
UPDATE users
SET phone = (
    SELECT n.canon FROM normalized n
    WHERE n.uid = users.id
      AND NOT EXISTS (SELECT 1 FROM users u2 WHERE u2.phone = n.canon)
      AND NOT EXISTS (SELECT 1 FROM normalized n2 WHERE n2.canon = n.canon AND n2.uid < n.uid)
    LIMIT 1
)
WHERE phone IS NULL;

-- Uniqueness for the new login identifiers, created AFTER the backfill (partial: ignore
-- NULLs so multiple users without phone/max are allowed). telegram_id already had a
-- UNIQUE index from 0001, so it is not redefined here.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone  ON users(phone)  WHERE phone  IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_max_id ON users(max_id) WHERE max_id IS NOT NULL;

-- Refresh tokens: one row per issued refresh token, enables rotation + revocation.
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,     -- sha256 of the opaque refresh token
    expires_at  TEXT NOT NULL,
    revoked     INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_refresh_user ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_hash ON refresh_tokens(token_hash);
