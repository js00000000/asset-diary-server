package repositories

import (
	"asset-diary/db"
	"asset-diary/models"
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type PriceCacheRepositoryInterface interface {
	Get(assetType string, symbol string) (*models.PriceCache, error)
	Set(cache *models.PriceCache) error
}

type PriceCacheRepository struct {
	redisClient db.RedisClient
}

func NewPriceCacheRepository(redisClient db.RedisClient) *PriceCacheRepository {
	return &PriceCacheRepository{redisClient: redisClient}
}

func (r *PriceCacheRepository) Get(assetType string, symbol string) (*models.PriceCache, error) {
	ctx := context.Background()

	// Get the cache data from Redis
	data, err := r.redisClient.Get(ctx, assetType+"_"+symbol).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss, return nil without error
		}
		return nil, err
	}

	// Unmarshal the JSON data into PriceCache struct
	var cache models.PriceCache
	if err := json.Unmarshal([]byte(data), &cache); err != nil {
		return nil, err
	}

	// Check if the cache has expired
	if time.Now().After(cache.ExpiresAt) {
		// Delete expired cache and return nil
		r.redisClient.Del(ctx, cache.GetRedisKey())
		return nil, nil
	}

	return &cache, nil
}

func (r *PriceCacheRepository) Set(cache *models.PriceCache) error {
	ctx := context.Background()

	// Marshal the cache to JSON
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	// Calculate TTL in seconds
	ttl := time.Until(cache.ExpiresAt)
	if ttl <= 0 {
		// If already expired, don't set the cache
		return nil
	}

	// Set the cache with TTL
	return r.redisClient.Set(ctx, cache.GetRedisKey(), data, ttl).Err()
}
