package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthCheckHandler struct {
	// Add any dependencies here if needed in the future
}

func NewHealthCheckHandler() *HealthCheckHandler {
	return &HealthCheckHandler{}
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Returns the health status of the server
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/healthz [get]
func (h *HealthCheckHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Server is running",
	})
}
