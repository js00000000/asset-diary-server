-- +migrate Up
CREATE TABLE IF NOT EXISTS user_daily_total_asset_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    total_value DECIMAL(20, 8) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date)
);

-- Index for faster lookups by user_id and date
CREATE INDEX IF NOT EXISTS idx_user_daily_total_asset_values_user_id ON user_daily_total_asset_values(user_id);
CREATE INDEX IF NOT EXISTS idx_user_daily_total_asset_values_date ON user_daily_total_asset_values(date);

-- Add comment to the table
COMMENT ON TABLE user_daily_total_asset_values IS 'Stores daily snapshots of user total asset values in their default currency';
