-- Bot v2: Student, Relationship, Lead, TrainingTemplate, Training, NotificationPrefs

-- Students: athletes who attend training (separate from User account)
CREATE TABLE IF NOT EXISTS students (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    full_name       TEXT NOT NULL,
    birth_date      TEXT,
    age             INTEGER,
    level           TEXT DEFAULT 'beginner',  -- beginner|intermediate|advanced
    phone           TEXT,
    additional_info TEXT,
    status          TEXT DEFAULT 'active',     -- active|paused|left
    client_id       INTEGER REFERENCES clients(id), -- link to existing client record if any
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT
);
CREATE INDEX IF NOT EXISTS idx_students_client ON students(client_id);

-- Relationships: User <-> Student (self / parent / guardian)
CREATE TABLE IF NOT EXISTS relationships (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id  INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    relation    TEXT NOT NULL DEFAULT 'self',  -- self|parent|guardian
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(user_id, student_id)
);
CREATE INDEX IF NOT EXISTS idx_rel_user ON relationships(user_id);
CREATE INDEX IF NOT EXISTS idx_rel_student ON relationships(student_id);

-- Leads: registration requests from the bot
CREATE TABLE IF NOT EXISTS leads (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    telegram_id     TEXT NOT NULL,
    full_name       TEXT NOT NULL,           -- User's full name (parent/adult)
    phone           TEXT,
    target_name     TEXT,                    -- Student's name (if child)
    target_age      INTEGER,
    target_level    TEXT DEFAULT 'beginner',
    reg_type        TEXT NOT NULL DEFAULT 'self',  -- self|child
    status          TEXT NOT NULL DEFAULT 'pending', -- pending|approved|rejected
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    reviewed_at     TEXT,
    reviewed_by     INTEGER REFERENCES users(id),
    created_user_id INTEGER,                 -- filled on approval
    created_student_id INTEGER,              -- filled on approval
    UNIQUE(telegram_id, target_name)
);
CREATE INDEX IF NOT EXISTS idx_leads_status ON leads(status);
CREATE INDEX IF NOT EXISTS idx_leads_telegram ON leads(telegram_id);

-- Training templates: recurring schedule per group
CREATE TABLE IF NOT EXISTS training_templates (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id    INTEGER REFERENCES groups(id) ON DELETE CASCADE,
    weekday     INTEGER NOT NULL,  -- 1=Mon ... 7=Sun
    time        TEXT NOT NULL,     -- HH:MM
    duration    INTEGER DEFAULT 60,
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_templates_group ON training_templates(group_id);

-- Trainings: concrete sessions (extends lesson_entries for bot display)
CREATE TABLE IF NOT EXISTS trainings (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id    INTEGER REFERENCES groups(id),
    coach_id    INTEGER REFERENCES coaches(id),
    date        TEXT NOT NULL,            -- YYYY-MM-DD
    time        TEXT NOT NULL,            -- HH:MM
    duration    INTEGER DEFAULT 60,
    status      TEXT DEFAULT 'planned',   -- planned|done|canceled
    location    TEXT,
    comment     TEXT,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_trainings_date ON trainings(date);
CREATE INDEX IF NOT EXISTS idx_trainings_group ON trainings(group_id);

-- Training attendance (per student per training)
CREATE TABLE IF NOT EXISTS training_attendance (
    training_id INTEGER NOT NULL REFERENCES trainings(id) ON DELETE CASCADE,
    student_id  INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    present     INTEGER DEFAULT 0,
    note        TEXT,
    marked_at   TEXT,
    marked_by   INTEGER REFERENCES users(id),
    PRIMARY KEY (training_id, student_id)
);

-- Training absence reasons (when student clicks "Не смогу прийти")
CREATE TABLE IF NOT EXISTS training_absences (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    training_id INTEGER NOT NULL REFERENCES trainings(id) ON DELETE CASCADE,
    student_id  INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    reason      TEXT,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(training_id, student_id)
);

-- Notification preferences (customizable notifications)
CREATE TABLE IF NOT EXISTS notification_prefs (
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id      INTEGER REFERENCES students(id) ON DELETE CASCADE,
    reminder_day    INTEGER DEFAULT 1,   -- reminder 1 day before
    reminder_hours  INTEGER DEFAULT 2,   -- reminder N hours before
    lessons_low     INTEGER DEFAULT 1,   -- notify when N lessons left
    sub_expiring    INTEGER DEFAULT 3,   -- notify N days before subscription ends
    news            INTEGER DEFAULT 1,   -- school news
    updated_at      TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (user_id, student_id)
);

-- Messages: client -> coach communication
CREATE TABLE IF NOT EXISTS client_messages (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER REFERENCES users(id),
    student_id  INTEGER REFERENCES students(id),
    text        TEXT NOT NULL,
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);
