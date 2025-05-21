package handlers

import (
	"net/http"
	"time"

	"asset-diary/models"
	"asset-diary/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TradeHandler struct {
	service services.TradeServiceInterface
}

func NewTradeHandler(tradeService services.TradeServiceInterface) *TradeHandler {
	return &TradeHandler{
		service: tradeService,
	}
}

// List all trades for a given account or user
func (h *TradeHandler) ListTrades(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	trades, err := h.service.ListTrades(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trades"})
		return
	}

	tradeResponses := []models.TradeResponse{}
	for _, trade := range trades {
		tradeResponses = append(tradeResponses, models.TradeResponse{
			ID:         trade.ID,
			Type:       trade.Type,
			AssetType:  trade.AssetType,
			Ticker:     trade.Ticker,
			TickerName: trade.TickerName,
			TradeDate:  trade.TradeDate,
			Quantity:   trade.Quantity,
			Price:      trade.Price,
			Currency:   trade.Currency,
			AccountID:  trade.AccountID,
			Reason:     trade.Reason,
			CreatedAt:  trade.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, tradeResponses)
}

// Create a trade
func (h *TradeHandler) CreateTrade(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var req models.TradeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Verify account belongs to user
	okAcc, err := h.service.IsAccountOwnedByUser(req.AccountID, userID.(string))
	if err != nil || !okAcc {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or unauthorized account_id"})
		return
	}
	tradeDate, err := time.Parse("2006-01-02", req.TradeDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tradeDate format, use YYYY-MM-DD"})
		return
	}
	trade := models.Trade{
		ID:         uuid.New().String(),
		Type:       req.Type,
		AssetType:  req.AssetType,
		Ticker:     req.Ticker,
		TickerName: req.TickerName,
		TradeDate:  tradeDate,
		Quantity:   req.Quantity,
		Price:      req.Price,
		Currency:   req.Currency,
		AccountID:  req.AccountID,
		Reason:     req.Reason,
	}
	createdTrade, err := h.service.CreateTrade(userID.(string), trade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trade"})
		return
	}
	tradeResponse := models.TradeResponse{
		ID:         createdTrade.ID,
		Type:       createdTrade.Type,
		AssetType:  createdTrade.AssetType,
		Ticker:     createdTrade.Ticker,
		TickerName: createdTrade.TickerName,
		TradeDate:  createdTrade.TradeDate,
		Quantity:   createdTrade.Quantity,
		Price:      createdTrade.Price,
		Currency:   createdTrade.Currency,
		AccountID:  createdTrade.AccountID,
		Reason:     createdTrade.Reason,
		CreatedAt:  createdTrade.CreatedAt,
	}
	c.JSON(http.StatusCreated, tradeResponse)
}

// Update a trade
func (h *TradeHandler) UpdateTrade(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	id := c.Param("id")
	// Check ownership of trade
	okTrade, err := h.service.IsTradeOwnedByUser(id, userID.(string))
	if err != nil || !okTrade {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trade not found or unauthorized"})
		return
	}
	var req models.TradeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.AccountID != "" {
		okAcc, err := h.service.IsAccountOwnedByUser(req.AccountID, userID.(string))
		if err != nil || !okAcc {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or unauthorized account_id"})
			return
		}
	}
	updatedTrade, err := h.service.UpdateTrade(userID.(string), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update trade"})
		return
	}
	if updatedTrade == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	tradeResponse := models.TradeResponse{
		ID:         updatedTrade.ID,
		Type:       updatedTrade.Type,
		AssetType:  updatedTrade.AssetType,
		Ticker:     updatedTrade.Ticker,
		TickerName: updatedTrade.TickerName,
		TradeDate:  updatedTrade.TradeDate,
		Quantity:   updatedTrade.Quantity,
		Price:      updatedTrade.Price,
		Currency:   updatedTrade.Currency,
		AccountID:  updatedTrade.AccountID,
		Reason:     updatedTrade.Reason,
		CreatedAt:  updatedTrade.CreatedAt,
	}
	c.JSON(http.StatusOK, tradeResponse)
}

// Delete a trade
func (h *TradeHandler) DeleteTrade(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	id := c.Param("id")
	deleted, err := h.service.DeleteTrade(userID.(string), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete trade"})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trade not found or unauthorized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "deleted": true})
}
