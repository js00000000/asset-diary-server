package handlers

import (
	"asset-diary/db"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RedisHandler struct {
	redisClient db.RedisClient
}

func NewRedisHandler() *RedisHandler {
	return &RedisHandler{
		redisClient: db.GetRedis(),
	}
}

// GetKey gets a value by key from Redis
// @Summary Get a value by key from Redis
// @Description Get a value by key from Redis
// @Tags redis
// @Produce json
// @Param key path string true "Redis key"
// @Success 200 {object} map[string]interface{}
// @Router /api/redis/get/{key} [get]
func (h *RedisHandler) GetKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	ctx := context.Background()

	// Get value
	val, err := h.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get TTL
	ttl, err := h.redisClient.TTL(ctx, key).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":   key,
		"value": val,
		"ttl":   ttl.Seconds(),
	})
}

// ListKeys lists all keys matching a pattern
// @Summary List all keys matching a pattern
// @Description List all keys matching a pattern (default: '*')
// @Tags redis
// @Produce json
// @Param pattern query string false "Key pattern (default: '*')"
// @Success 200 {object} map[string]interface{}
// @Router /api/redis/keys [get]
func (h *RedisHandler) ListKeys(c *gin.Context) {
	pattern := c.DefaultQuery("pattern", "*")

	ctx := context.Background()

	keys, err := h.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get values for all keys
	result := make(map[string]string)
	for _, key := range keys {
		val, err := h.redisClient.Get(ctx, key).Result()
		if err == nil {
			result[key] = val
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(keys),
		"keys":  result,
	})
}

// FlushKeys deletes all keys in the current database
// @Summary Delete all keys in the current database
// @Description Flush all keys from the current Redis database
// @Tags redis
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/redis/flush [post]
func (h *RedisHandler) FlushKeys(c *gin.Context) {
	ctx := context.Background()

	// Flush all keys in the current database
	status, err := h.redisClient.FlushDB(ctx).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  status,
		"message": "All keys have been deleted from the current database",
	})
}
