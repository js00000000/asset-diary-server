package handlers

import (
	"asset-diary/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HoldingHandler struct {
	holdingService services.HoldingServiceInterface
}

func NewHoldingHandler(holdingService services.HoldingServiceInterface) *HoldingHandler {
	return &HoldingHandler{
		holdingService: holdingService,
	}
}

// ListHoldings handles GET /holdings
func (h *HoldingHandler) ListHoldings(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	holdings, err := h.holdingService.ListHoldings(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, holdings)
}
