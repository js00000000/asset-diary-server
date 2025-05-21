-- Drop ticker_name column from trades table
ALTER TABLE trades 
DROP COLUMN IF EXISTS ticker_name;
