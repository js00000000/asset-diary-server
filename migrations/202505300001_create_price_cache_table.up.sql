CREATE TABLE price_caches (
    id SERIAL PRIMARY KEY,
    cache_key VARCHAR(255) NOT NULL UNIQUE,
    symbol VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(24, 8) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_price_cache_key ON price_caches (cache_key);
CREATE INDEX idx_price_cache_expires ON price_caches (expires_at);
