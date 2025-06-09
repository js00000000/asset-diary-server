-- Drop the trigger first
DROP TRIGGER IF EXISTS update_refresh_tokens_updated_at ON refresh_tokens;

-- Drop the function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop the table and all its dependent objects
DROP TABLE IF EXISTS refresh_tokens CASCADE;
