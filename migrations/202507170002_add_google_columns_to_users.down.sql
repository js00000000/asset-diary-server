-- +migrate Down
-- Remove the index first
DROP INDEX IF EXISTS idx_users_google_id;

-- Remove the columns
ALTER TABLE users 
DROP COLUMN IF EXISTS google_id,
DROP COLUMN IF EXISTS google_email;
