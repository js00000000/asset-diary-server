package handlers

import (
	"asset-diary/models"
	"asset-diary/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GeminiTestHandler handles requests for testing the Gemini services
type GeminiTestHandler struct {
	geminiChatService  *services.GeminiChatService
	geminiPriceService *services.GeminiAssetPriceService
}

// NewGeminiTestHandler creates a new instance of GeminiTestHandler
func NewGeminiTestHandler(geminiChatService *services.GeminiChatService, geminiPriceService *services.GeminiAssetPriceService) *GeminiTestHandler {
	return &GeminiTestHandler{
		geminiChatService:  geminiChatService,
		geminiPriceService: geminiPriceService,
	}
}

// TestGenerateContentRequest represents the request body for testing GenerateContent
type TestGenerateContentRequest struct {
	Message string `json:"message" binding:"required"`
}

// TestGenerateContentResponse represents the response from the test endpoint
type TestGenerateContentResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// TestGenerateContent tests the GenerateContent method of GeminiChatService
func (h *GeminiTestHandler) TestGenerateContent(c *gin.Context) {
	var req TestGenerateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewAppError(models.ErrCodeInvalidRequest, "Invalid request body"))
		return
	}

	response, err := h.geminiChatService.GenerateContent(req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAppError(models.ErrCodeInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, TestGenerateContentResponse{
		Success: true,
		Message: response,
	})
}

// TestGenerateContentWithHistoryRequest represents the request body for testing GenerateContentWithHistory
type TestGenerateContentWithHistoryRequest struct {
	Messages []struct {
		Role    string `json:"role" binding:"required"`
		Content string `json:"content" binding:"required"`
	} `json:"messages" binding:"required,min=1"`
}

// TestAssetPriceRequest represents the request body for testing asset price
type TestAssetPriceRequest struct {
	Symbol string `json:"symbol" binding:"required"`
}

// TestAssetPrice tests the GeminiAssetPriceService
func (h *GeminiTestHandler) TestAssetPrice(c *gin.Context) {
	var req struct {
		Symbol string `json:"symbol" binding:"required"`
		Type   string `json:"type" binding:"required,oneof=stock crypto"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewAppError(models.ErrCodeInvalidRequest, "Invalid request: " + err.Error()))
		return
	}

	var (
		tickerInfo interface{}
		err        error
	)

	switch req.Type {
	case "stock":
		tickerInfo, err = h.geminiPriceService.GetStockPrice(req.Symbol)
	case "crypto":
		tickerInfo, err = h.geminiPriceService.GetCryptoPrice(req.Symbol)
	default:
		c.JSON(http.StatusBadRequest, models.NewAppError(models.ErrCodeInvalidRequest, "Invalid asset type. Must be 'stock' or 'crypto'"))
		return
	}

	if err != nil {
		if err.Error() == services.InvalidSymbolError {
			c.JSON(http.StatusBadRequest, models.NewAppError(models.ErrCodeInvalidRequest, "Invalid symbol"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewAppError(models.ErrCodeInternal, "failed to get stock price: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tickerInfo,
	})
}

// RegisterRoutes registers the test routes for the Gemini services
func (h *GeminiTestHandler) RegisterRoutes(router *gin.RouterGroup) {
	geminiGroup := router.Group("/gemini-test")
	{
		// Chat service endpoints
		geminiGroup.POST("/generate-content", h.TestGenerateContent)

		// Asset price service endpoints
		geminiGroup.POST("/asset-price", h.TestAssetPrice)
	}
}
