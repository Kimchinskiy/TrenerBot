-- Add bot_access and subscription_ends_at columns to clients table
ALTER TABLE clients ADD COLUMN bot_access INTEGER DEFAULT 0;
ALTER TABLE clients ADD COLUMN subscription_ends_at TEXT;