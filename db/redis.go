package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	TTL(ctx context.Context, key string) *redis.DurationCmd
	Keys(ctx context.Context, pattern string) *redis.StringSliceCmd
	FlushDB(ctx context.Context) *redis.StatusCmd
	Close() error
}

var redisClient *redis.Client

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

	opt.MaxRetries = 3
	opt.MinRetryBackoff = 8 * time.Millisecond
	opt.MaxRetryBackoff = 512 * time.Millisecond

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	redisClient = client
	return client, nil
}

func GetRedis() RedisClient {
	return redisClient
}
