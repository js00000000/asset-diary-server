package repositories

import (
	"log"
	"time"

	"asset-dairy/models"

	"gorm.io/gorm"
)

// TradeRepositoryInterface defines methods for trade-related database operations
type TradeRepositoryInterface interface {
	ListTrades(userID string) ([]models.Trade, error)
	CreateTrade(userID string, trade models.Trade) (*models.Trade, error)
	UpdateTrade(userID, tradeID string, req models.TradeUpdateRequest) (*models.Trade, error)
	DeleteTrade(userID, tradeID string) (bool, error)
	IsAccountOwnedByUser(accountID, userID string) (bool, error)
	IsTradeOwnedByUser(tradeID, userID string) (bool, error)
}

// TradeRepository implements TradeRepositoryInterface
type TradeRepository struct {
	db *gorm.DB
}

// NewTradeRepository creates a new TradeRepository instance
func NewTradeRepository(db *gorm.DB) *TradeRepository {
	return &TradeRepository{db: db}
}

// ListTrades retrieves all trades for a given user
func (r *TradeRepository) ListTrades(userID string) ([]models.Trade, error) {
	var gormTrades []models.Trade
	result := r.db.Where(&models.Trade{UserID: userID}).Find(&gormTrades)
	if result.Error != nil {
		log.Println("TradeRepository: Failed to fetch trades:", result.Error)
		return nil, result.Error
	}

	trades := []models.Trade{}
	for _, gormTrade := range gormTrades {
		trade := models.Trade{
			ID:        gormTrade.ID,
			Type:      gormTrade.Type,
			AssetType: gormTrade.AssetType,
			Ticker:    gormTrade.Ticker,
			TradeDate: gormTrade.TradeDate,
			Quantity:  gormTrade.Quantity,
			Price:     gormTrade.Price,
			Currency:  gormTrade.Currency,
			AccountID: gormTrade.AccountID,
			Reason:    gormTrade.Reason,
			CreatedAt: gormTrade.CreatedAt,
		}
		trades = append(trades, trade)
	}

	return trades, nil
}

func (r *TradeRepository) IsAccountOwnedByUser(accountID, userID string) (bool, error) {
	var count int64
	result := r.db.Model(&models.Account{}).Where(&models.Account{ID: accountID, UserID: userID}).Count(&count)
	return count > 0, result.Error
}

func (r *TradeRepository) IsTradeOwnedByUser(tradeID, userID string) (bool, error) {
	var count int64
	result := r.db.Model(&models.Trade{}).Joins("JOIN accounts a ON trades.account_id = a.id").Where("trades.id = ? AND a.user_id = ?", tradeID, userID).Count(&count)
	return count > 0, result.Error
}

func (r *TradeRepository) CreateTrade(userID string, trade models.Trade) (*models.Trade, error) {
	gormTrade := &models.Trade{
		ID:        trade.ID,
		UserID:    userID,
		Type:      trade.Type,
		AssetType: trade.AssetType,
		Ticker:    trade.Ticker,
		TradeDate: trade.TradeDate,
		Quantity:  trade.Quantity,
		Price:     trade.Price,
		Currency:  trade.Currency,
		AccountID: trade.AccountID,
		Reason:    trade.Reason,
		CreatedAt: time.Now(),
	}

	result := r.db.Create(gormTrade)
	if result.Error != nil {
		return nil, result.Error
	}

	// Fetch the created trade to get all fields populated by DB
	var createdTrade models.Trade
	if err := r.db.First(&createdTrade, "id = ?", gormTrade.ID).Error; err != nil {
		return nil, err
	}

	return &createdTrade, nil
}

func (r *TradeRepository) UpdateTrade(userID, tradeID string, req models.TradeUpdateRequest) (*models.Trade, error) {
	var gormTrade models.Trade
	result := r.db.Where(&models.Trade{ID: tradeID, UserID: userID}).First(&gormTrade)
	if result.Error != nil {
		log.Println("Failed to find trade:", result.Error)
		return nil, result.Error
	}

	// Update fields if provided
	if req.Type != "" {
		gormTrade.Type = req.Type
	}
	if req.AssetType != "" {
		gormTrade.AssetType = req.AssetType
	}
	if req.TradeDate != "" {
		tradeDate, err := time.Parse("2006-01-02", req.TradeDate)
		if err != nil {
			return nil, err
		}
		gormTrade.TradeDate = tradeDate
	}
	if req.Ticker != "" {
		gormTrade.Ticker = req.Ticker
	}
	if req.Quantity != 0 {
		gormTrade.Quantity = req.Quantity
	}
	if req.Price != 0 {
		gormTrade.Price = req.Price
	}
	if req.Currency != "" {
		gormTrade.Currency = req.Currency
	}
	if req.Reason != nil {
		gormTrade.Reason = req.Reason
	}

	result = r.db.Save(&gormTrade)
	if result.Error != nil {
		log.Println("Failed to update trade:", result.Error)
		return nil, result.Error
	}

	return &models.Trade{
		ID:        gormTrade.ID,
		Type:      gormTrade.Type,
		AssetType: gormTrade.AssetType,
		Ticker:    gormTrade.Ticker,
		TradeDate: gormTrade.TradeDate,
		Quantity:  gormTrade.Quantity,
		Price:     gormTrade.Price,
		Currency:  gormTrade.Currency,
		AccountID: gormTrade.AccountID,
		Reason:    gormTrade.Reason,
	}, nil
}

func (r *TradeRepository) DeleteTrade(userID, tradeID string) (bool, error) {
	result := r.db.Where("id = ? AND user_id = ?", tradeID, userID).Delete(&models.Trade{})
	if result.Error != nil {
		log.Println("Failed to delete trade:", result.Error)
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}
