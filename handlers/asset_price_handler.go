package handlers

import (
	"asset-diary/models"
	"asset-diary/services/interfaces"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AssetPriceHandler handles requests related to asset prices (stocks and cryptocurrencies)
type AssetPriceHandler struct {
	assetPriceService interfaces.AssetPriceServiceInterface
}

// NewAssetPriceHandler creates a new instance of AssetPriceHandler
func NewAssetPriceHandler(assetPriceService interfaces.AssetPriceServiceInterface) *AssetPriceHandler {
	return &AssetPriceHandler{
		assetPriceService: assetPriceService,
	}
}

// GetStockPrice handles GET /stock/price/:symbol
// Supports stock symbols and Taiwan stock codes
// Examples: /stock/price/AAPL or /stock/price/2330
func (h *AssetPriceHandler) GetStockPrice(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.NewAppError(models.ErrCodeInvalidRequest, "Symbol is required"))
		return
	}

	tickerInfo, err := h.assetPriceService.GetStockPrice(symbol)
	if err != nil {
		h.handlePriceError(c, err)
		return
	}

	c.JSON(http.StatusOK, tickerInfo)
}

// GetCryptoPrice handles GET /crypto/price/:symbol
// Example: /crypto/price/BTC or /crypto/price/ETH
// Always returns price in USDT
func (h *AssetPriceHandler) GetCryptoPrice(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.NewAppError(models.ErrCodeInvalidRequest, "Symbol is required"))
		return
	}

	tickerInfo, err := h.assetPriceService.GetCryptoPrice(symbol)
	if err != nil {
		h.handlePriceError(c, err)
		return
	}

	c.JSON(http.StatusOK, tickerInfo)
}

// handlePriceError handles common price-related errors
func (h *AssetPriceHandler) handlePriceError(c *gin.Context, err error) {
	errMsg := err.Error()

	// More specific error handling
	switch {
	case strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "no data"):
		c.JSON(http.StatusNotFound, models.NewAppError(models.ErrCodeInvalidRequest, errMsg))
	case strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "failed to fetch"):
		c.JSON(http.StatusBadRequest, models.NewAppError(models.ErrCodeInvalidRequest, errMsg))
	default:
		c.JSON(http.StatusInternalServerError, models.NewAppError(models.ErrCodeInternal, errMsg))
	}
}
