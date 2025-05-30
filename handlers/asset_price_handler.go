package handlers

import (
	"asset-diary/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AssetPriceHandler handles requests related to asset prices (stocks and cryptocurrencies)
type AssetPriceHandler struct {
	assetPriceService services.AssetPriceServiceInterface
}

// NewAssetPriceHandler creates a new instance of AssetPriceHandler
func NewAssetPriceHandler(assetPriceService services.AssetPriceServiceInterface) *AssetPriceHandler {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
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
	statusCode := http.StatusInternalServerError
	errMsg := err.Error()

	// More specific error handling
	switch {
	case strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "no data"):
		statusCode = http.StatusNotFound
	case strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "failed to fetch"):
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, gin.H{"error": errMsg})
}
