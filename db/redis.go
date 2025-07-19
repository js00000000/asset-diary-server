package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient is an interface that matches the Redis client methods we use
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	TTL(ctx context.Context, key string) *redis.DurationCmd
	Keys(ctx context.Context, pattern string) *redis.StringSliceCmd
	Close() error
}

var redisClient *redis.Client

// InitRedis initializes and returns a Redis client
func InitRedis() (RedisClient, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Println("Warning: REDIS_URL not set, falling back to default URL")
		redisURL = "redis://localhost:6379/0"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	// Add additional configuration
	opt.MaxRetries = 3
	opt.MinRetryBackoff = 8 * time.Millisecond
	opt.MaxRetryBackoff = 512 * time.Millisecond

	client := redis.NewClient(opt)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	// Set global Redis client
	redisClient = client
	return client, nil
}

// GetRedis returns the global Redis client
func GetRedis() RedisClient {
	return redisClient
}
