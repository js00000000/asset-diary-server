-- Add composite index for faster refresh token validation
CREATE INDEX idx_refresh_token_validation ON refresh_tokens(token_hash, revoked, expires_at);
