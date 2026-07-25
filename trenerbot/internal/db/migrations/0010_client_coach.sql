ALTER TABLE clients ADD COLUMN coach_id INTEGER REFERENCES coaches(id);
CREATE INDEX IF NOT EXISTS idx_clients_coach ON clients(coach_id);
