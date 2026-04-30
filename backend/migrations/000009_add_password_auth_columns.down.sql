-- Remove password authentication columns from users table
-- Rollback to OAuth-only authentication

ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
ALTER TABLE users DROP COLUMN IF EXISTS verification_code;
ALTER TABLE users DROP COLUMN IF EXISTS verification_code_expiry;
ALTER TABLE users DROP COLUMN IF EXISTS password_reset_code;
ALTER TABLE users DROP COLUMN IF EXISTS password_reset_expiry;
