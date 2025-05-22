-- +migrate Down
ALTER TABLE users 
ADD COLUMN name VARCHAR(255);
