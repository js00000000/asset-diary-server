package middleware

import (
	"asset-diary/db"
	"asset-diary/models"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimit creates a middleware for rate limiting using Redis
// limit: number of requests allowed per window
// window: duration of the window
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		redisClient := db.GetRedis()
		if redisClient == nil {
			log.Println("Rate limiting skipped: Redis client not initialized")
			c.Next()
			return
		}

		// Use client IP as the key, include path to have per-endpoint limits
		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s:%s", c.FullPath(), ip)

		ctx := c.Request.Context()

		// Fixed window implementation
		// Check the current count
		countCmd := redisClient.Get(ctx, key)
		count, err := countCmd.Int64()
		if err != nil && err != redis.Nil {
			log.Printf("Rate limit lookup error: %v", err)
			c.Next() // Fallback: allow request if redis is failing
			return
		}

		// If limit exceeded, return error
		if count >= int64(limit) {
			c.JSON(http.StatusTooManyRequests, models.NewAppError(
				models.ErrCodeTooManyRequests,
				fmt.Sprintf("Too many requests. You are limited to %d requests per %s.", limit, window),
			))
			c.Abort()
			return
		}

		// Increment the counter
		newCount, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			log.Printf("Rate limit increment error: %v", err)
			c.Next()
			return
		}

		// If this is the first request in the window, set expiration
		if newCount == 1 {
			redisClient.Expire(ctx, key, window)
		}

		c.Next()
	}
}
