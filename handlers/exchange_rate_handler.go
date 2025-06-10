package handlers

import (
	"net/http"

	"asset-diary/services"

	"github.com/gin-gonic/gin"
)

type ExchangeRateHandler struct {
	service services.ExchangeRateServiceInterface
}

func NewExchangeRateHandler(service services.ExchangeRateServiceInterface) *ExchangeRateHandler {
	return &ExchangeRateHandler{service: service}
}

type ExchangeRateResponse struct {
	BaseCurrency string             `json:"base_currency"`
	Rates        map[string]float64 `json:"rates"`
}

// GetRatesByBaseCurrency godoc
// @Summary Get exchange rates by base currency
// @Description Get all exchange rates for a specific base currency
// @Tags exchange-rates
// @Accept json
// @Produce json
// @Param base_currency path string true "Base Currency Code (e.g., TWD, USD)"
// @Success 200 {object} ExchangeRateResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /exchange-rates/{base_currency} [get]
func (h *ExchangeRateHandler) GetRatesByBaseCurrency(c *gin.Context) {
	baseCurrency := c.Param("base_currency")
	if baseCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "base_currency is required"})
		return
	}

	rates, err := h.service.GetRatesByBaseCurrency(baseCurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch exchange rates"})
		return
	}

	if len(rates) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no exchange rates found for the specified base currency"})
		return
	}

	response := ExchangeRateResponse{
		BaseCurrency: baseCurrency,
		Rates:        rates,
	}

	c.JSON(http.StatusOK, response)
}
