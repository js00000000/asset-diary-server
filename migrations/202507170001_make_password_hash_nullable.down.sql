-- +migrate Down
-- Set a default empty password for any null values before altering the column
UPDATE users SET password_hash = '' WHERE password_hash IS NULL;

-- Make the column NOT NULL again
ALTER TABLE users 
ALTER COLUMN password_hash SET NOT NULL;
