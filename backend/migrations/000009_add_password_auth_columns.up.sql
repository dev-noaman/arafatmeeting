-- Add password authentication columns to users table
-- Migration for SQLite-based local authentication

-- Password hash column (stores bcrypt hash)
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);

-- Email verification columns
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS verification_code VARCHAR(6);
ALTER TABLE users ADD COLUMN IF NOT EXISTS verification_code_expiry TIMESTAMP;

-- Password reset columns
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_reset_code VARCHAR(6);
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_reset_expiry TIMESTAMP;

-- Set default value for email_verified
UPDATE users SET email_verified = true WHERE email_verified IS NULL;
