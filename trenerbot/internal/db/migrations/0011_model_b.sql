-- Model B: User + Student (remove Client as separate entity)
-- Students become the central entity for ALL training individuals.
-- Clients table is kept as FK bridge (will be removed after migration stabilizes).

-- 1. Add all client fields to students
ALTER TABLE students ADD COLUMN user_id INTEGER REFERENCES users(id);
ALTER TABLE students ADD COLUMN photo TEXT;
ALTER TABLE students ADD COLUMN telegram TEXT;
ALTER TABLE students ADD COLUMN whatsapp TEXT;
ALTER TABLE students ADD COLUMN email TEXT;
ALTER TABLE students ADD COLUMN medical_limits TEXT;
ALTER TABLE students ADD COLUMN note TEXT;
ALTER TABLE students ADD COLUMN registered_at TEXT;
ALTER TABLE students ADD COLUMN source TEXT;
ALTER TABLE students ADD COLUMN bot_access INTEGER DEFAULT 0;
ALTER TABLE students ADD COLUMN coach_id INTEGER REFERENCES coaches(id);

-- 2. Migrate data from clients to students via client_id bridge
UPDATE students SET
  user_id       = (SELECT c.user_id       FROM clients c WHERE c.id = students.client_id),
  photo         = (SELECT c.photo         FROM clients c WHERE c.id = students.client_id),
  telegram      = (SELECT c.telegram      FROM clients c WHERE c.id = students.client_id),
  whatsapp      = (SELECT c.whatsapp      FROM clients c WHERE c.id = students.client_id),
  email         = (SELECT c.email         FROM clients c WHERE c.id = students.client_id),
  medical_limits= (SELECT c.medical_limits FROM clients c WHERE c.id = students.client_id),
  note          = (SELECT c.note          FROM clients c WHERE c.id = students.client_id),
  registered_at = (SELECT c.registered_at FROM clients c WHERE c.id = students.client_id),
  source        = (SELECT c.source        FROM clients c WHERE c.id = students.client_id),
  bot_access    = (SELECT c.bot_access    FROM clients c WHERE c.id = students.client_id),
  coach_id      = (SELECT c.coach_id      FROM clients c WHERE c.id = students.client_id)
WHERE client_id IS NOT NULL;

-- 3. Create students for clients that don't have one (self-registered via /start without lead)
INSERT INTO students(user_id, full_name, birth_date, age, phone, telegram, whatsapp,
  email, medical_limits, note, status, registered_at, source, bot_access, coach_id, client_id)
SELECT c.user_id, c.full_name, c.birth_date, c.age, c.phone, c.telegram, c.whatsapp,
  c.email, c.medical_limits, c.note, c.status, c.registered_at, c.source, c.bot_access, c.coach_id, c.id
FROM clients c
WHERE c.id NOT IN (SELECT s.client_id FROM students s WHERE s.client_id IS NOT NULL);

-- 4. Create 'self' relationships for students without one
INSERT OR IGNORE INTO relationships(user_id, student_id, relation)
SELECT s.user_id, s.id, 'self'
FROM students s
WHERE s.user_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM relationships r WHERE r.user_id = s.user_id AND r.student_id = s.id);

-- 5. Migrate parent relationships from clients.parent_id / second_parent_id
INSERT OR IGNORE INTO relationships(user_id, student_id, relation)
SELECT c.parent_id, s.id, 'parent'
FROM clients c
JOIN students s ON s.client_id = c.id
WHERE c.parent_id IS NOT NULL;

INSERT OR IGNORE INTO relationships(user_id, student_id, relation)
SELECT c.second_parent_id, s.id, 'parent'
FROM clients c
JOIN students s ON s.client_id = c.id
WHERE c.second_parent_id IS NOT NULL;

-- 6. Index for performance
CREATE INDEX IF NOT EXISTS idx_students_user ON students(user_id);
CREATE INDEX IF NOT EXISTS idx_students_coach ON students(coach_id);
CREATE INDEX IF NOT EXISTS idx_students_status ON students(status);
