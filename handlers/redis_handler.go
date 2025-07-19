package handlers

import (
	"asset-diary/db"
	"context"
	"net/http"
	"time"

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

type SetKeyRequest struct {
	Key   string        `json:"key" binding:"required"`
	Value string        `json:"value" binding:"required"`
	TTL   time.Duration `json:"ttl_seconds"` // TTL in seconds
}

// SetKey sets a key-value pair in Redis
// @Summary Set a key-value pair in Redis
// @Description Set a key with optional TTL (in seconds)
// @Tags redis
// @Accept json
// @Produce json
// @Param input body SetKeyRequest true "Key-Value pair with optional TTL"
// @Success 200 {object} map[string]interface{}
// @Router /api/redis/set [post]
func (h *RedisHandler) SetKey(c *gin.Context) {
	var req SetKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	err := h.redisClient.Set(ctx, req.Key, req.Value, req.TTL*time.Second).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"key":    req.Key,
		"value":  req.Value,
		"ttl":    req.TTL,
	})
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
