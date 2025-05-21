-- Add ticker_name column to trades table
ALTER TABLE trades 
ADD COLUMN ticker_name TEXT NOT NULL DEFAULT '';

-- Update existing rows to set ticker_name to ticker
UPDATE trades 
SET ticker_name = ticker;
