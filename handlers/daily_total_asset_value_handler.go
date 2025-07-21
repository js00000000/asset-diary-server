package handlers

import (
	"net/http"
	"time"

	"asset-diary/models"
	"asset-diary/services"

	"github.com/gin-gonic/gin"
)

type DailyTotalAssetValueHandler struct {
	DailyTotalAssetValueService services.DailyTotalAssetValueServiceInterface
}

func NewDailyTotalAssetValueHandler(service services.DailyTotalAssetValueServiceInterface) *DailyTotalAssetValueHandler {
	return &DailyTotalAssetValueHandler{
		DailyTotalAssetValueService: service,
	}
}

type GetUserDailyTotalAssetValuesRequest struct {
	StartDate string `form:"start_date" binding:"required,datetime=2006-01-02"`
	EndDate   string `form:"end_date" binding:"required,datetime=2006-01-02"`
}

func (h *DailyTotalAssetValueHandler) GetUserDailyTotalAssetValues(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewAppError(models.ErrCodeUnauthorized, "Unauthorized"))
		return
	}

	var req GetUserDailyTotalAssetValuesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewAppError(models.ErrCodeInvalidRequest, err.Error()))
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewAppError(models.ErrCodeInvalidRequest, "Invalid start_date format. Use YYYY-MM-DD"))
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewAppError(models.ErrCodeInvalidRequest, "Invalid end_date format. Use YYYY-MM-DD"))
		return
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, models.NewAppError(models.ErrCodeInvalidRequest, "end_date must be after or equal to start_date"))
		return
	}

	assetValues, err := h.DailyTotalAssetValueService.GetUserDailyTotalAssetValues(userID.(string), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAppError(models.ErrCodeInternal, "Failed to fetch daily total asset values"))
		return
	}

	c.JSON(http.StatusOK, assetValues)
}
