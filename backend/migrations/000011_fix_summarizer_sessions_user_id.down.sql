BEGIN;

-- Revert user_id back to integer
ALTER TABLE summarizer_sessions
    ALTER COLUMN user_id TYPE integer USING user_id::integer;

COMMIT;
