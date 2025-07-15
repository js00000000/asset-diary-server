package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// APIKeyAuthMiddleware checks for a valid API key in the X-API-Key header
func APIKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		expectedKey := os.Getenv("CRON_API_KEY")

		if expectedKey == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "server configuration error"})
			return
		}

		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			return
		}

		if apiKey != expectedKey {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid API key"})
			return
		}

		c.Next()
	}
}
