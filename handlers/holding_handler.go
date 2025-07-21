package handlers

import (
	"asset-diary/models"
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
		c.JSON(http.StatusUnauthorized, models.NewAppError(models.ErrCodeUnauthorized, "Unauthorized"))
		return
	}

	holdings, err := h.holdingService.ListHoldings(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAppError(models.ErrCodeInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, holdings)
}
