BEGIN;

-- Drop the existing FK constraint that references users(id) with integer type
ALTER TABLE summarizer_sessions DROP CONSTRAINT IF EXISTS summarizer_sessions_user_id_fkey;

-- Change user_id from integer to varchar(36) for InsForge UUID compatibility
ALTER TABLE summarizer_sessions
    ALTER COLUMN user_id TYPE varchar(36) USING user_id::varchar(36);

COMMIT;
