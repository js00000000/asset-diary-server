package handlers

import (
	"net/http"

	"asset-diary/services"

	"github.com/gin-gonic/gin"
)

type CronHandler struct {
	exchangeRateService    services.ExchangeRateServiceInterface
	assetValueService      services.DailyTotalAssetValueServiceInterface
}

func NewCronHandler(
	exchangeRateService services.ExchangeRateServiceInterface,
	assetValueService services.DailyTotalAssetValueServiceInterface,
) *CronHandler {
	return &CronHandler{
		exchangeRateService:    exchangeRateService,
		assetValueService:      assetValueService,
	}
}

// UpdateExchangeRates godoc
// @Summary Update exchange rates
// @Description Updates all exchange rates from the external API
// @Tags cron
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cron/update-exchange-rates [post]
func (h *CronHandler) UpdateExchangeRates(c *gin.Context) {
	if err := h.exchangeRateService.FetchAndStoreRates(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update exchange rates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exchange rates updated successfully"})
}

// RecordDailyAssets godoc
// @Summary Record daily asset values
// @Description Records the current total asset values for all users
// @Tags cron
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cron/record-daily-assets [post]
func (h *CronHandler) RecordDailyAssets(c *gin.Context) {
	if err := h.assetValueService.RecordDailyTotalAssetValue(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record daily assets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Daily assets recorded successfully"})
}
