-- +migrate Down
DROP INDEX IF EXISTS idx_user_daily_total_asset_values_date;
DROP INDEX IF EXISTS idx_user_daily_total_asset_values_user_id;
DROP TABLE IF EXISTS user_daily_total_asset_values;
