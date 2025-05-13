package services

import (
	"asset-dairy/models"
	"asset-dairy/repositories"
)

type TradeServiceInterface interface {
	ListTrades(userID string) ([]models.Trade, error)
	CreateTrade(userID string, trade models.Trade) (*models.Trade, error)
	UpdateTrade(userID, tradeID string, req models.TradeUpdateRequest) (*models.Trade, error)
	DeleteTrade(userID, tradeID string) (bool, error)
	IsAccountOwnedByUser(accountID, userID string) (bool, error)
	IsTradeOwnedByUser(tradeID, userID string) (bool, error)
}

type TradeService struct {
	repo repositories.TradeRepositoryInterface
}

// NewTradeService creates a new TradeService instance with a repository
func NewTradeService(repo repositories.TradeRepositoryInterface) *TradeService {
	return &TradeService{repo: repo}
}

// ListTrades retrieves all trades for a given user
func (s *TradeService) ListTrades(userID string) ([]models.Trade, error) {
	return s.repo.ListTrades(userID)
}

func (s *TradeService) CreateTrade(userID string, trade models.Trade) (*models.Trade, error) {
	return s.repo.CreateTrade(userID, trade)
}

func (s *TradeService) UpdateTrade(userID, tradeID string, req models.TradeUpdateRequest) (*models.Trade, error) {
	return s.repo.UpdateTrade(userID, tradeID, req)
}

func (s *TradeService) DeleteTrade(userID, tradeID string) (bool, error) {
	return s.repo.DeleteTrade(userID, tradeID)
}

func (s *TradeService) IsAccountOwnedByUser(accountID, userID string) (bool, error) {
	return s.repo.IsAccountOwnedByUser(accountID, userID)
}

func (s *TradeService) IsTradeOwnedByUser(tradeID, userID string) (bool, error) {
	return s.repo.IsTradeOwnedByUser(tradeID, userID)
}
