package services

import (
	"asset-diary/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTradeService is a mock implementation of TradeServiceInterface
type MockTradeService struct {
	mock.Mock
}

func (m *MockTradeService) ListTrades(userID string) ([]models.Trade, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Trade), args.Error(1)
}

// Add stub methods to satisfy TradeServiceInterface
func (m *MockTradeService) CreateTrade(userID string, trade models.Trade) (*models.Trade, error) {
	args := m.Called(userID, trade)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Trade), args.Error(1)
}
func (m *MockTradeService) UpdateTrade(userID, tradeID string, req models.TradeUpdateRequest) (*models.Trade, error) {
	panic("not implemented")
}
func (m *MockTradeService) DeleteTrade(userID, tradeID string) (bool, error) {
	panic("not implemented")
}
func (m *MockTradeService) IsAccountOwnedByUser(accountID, userID string) (bool, error) {
	panic("not implemented")
}
func (m *MockTradeService) IsTradeOwnedByUser(tradeID, userID string) (bool, error) {
	panic("not implemented")
}

func TestListAssets(t *testing.T) {
	tests := []struct {
		name           string
		trades         []models.Trade
		expectedAssets []models.Holding
		expectedError  error
	}{
		{
			name:           "no trades should return empty list",
			trades:         []models.Trade{},
			expectedAssets: []models.Holding{},
			expectedError:  nil,
		},
		{
			name: "zero quantity holdings should not be included",
			trades: []models.Trade{
				{
					Type:      "buy",
					AssetType: "stock",
					Ticker:    "AAPL",
					Quantity:  10,
					Price:     100,
					Currency:  "USD",
				},
				{
					Type:      "sell",
					AssetType: "stock",
					Ticker:    "AAPL",
					Quantity:  10,
					Price:     150,
					Currency:  "USD",
				},
			},
			expectedAssets: []models.Holding{},
			expectedError:  nil,
		},
		{
			name: "multiple buys should calculate correct average price",
			trades: []models.Trade{
				{
					Type:      "buy",
					AssetType: "stock",
					Ticker:    "AAPL",
					Quantity:  5,
					Price:     100,
					Currency:  "USD",
				},
				{
					Type:      "buy",
					AssetType: "stock",
					Ticker:    "AAPL",
					Quantity:  5,
					Price:     300,
					Currency:  "USD",
				},
			},
			expectedAssets: []models.Holding{
				{
					Ticker:       "AAPL",
					Quantity:     10,
					AveragePrice: (5*100 + 5*300) / 10,
					AssetType:    "stock",
					Currency:     "USD",
				},
			},
			expectedError: nil,
		},
		{
			name: "sell should not affect average price of remaining shares",
			trades: []models.Trade{
				{
					Type:      "buy",
					AssetType: "stock",
					Ticker:    "AAPL",
					Quantity:  10,
					Price:     100,
					Currency:  "USD",
				},
				{
					Type:      "buy",
					AssetType: "stock",
					Ticker:    "AAPL",
					Quantity:  5,
					Price:     200,
					Currency:  "USD",
				},
				{
					Type:      "sell",
					AssetType: "stock",
					Ticker:    "AAPL",
					Quantity:  5,
					Price:     300,
					Currency:  "USD",
				},
			},
			expectedAssets: []models.Holding{
				{
					Ticker:       "AAPL",
					Quantity:     10,
					AveragePrice: (5*100 + 5*200) / 10,
					AssetType:    "stock",
					Currency:     "USD",
				},
			},
			expectedError: nil,
		},
		{
			name: "multiple assets should be handled correctly",
			trades: []models.Trade{
				{
					Type:      "buy",
					AssetType: "stock",
					Ticker:    "AAPL",
					Quantity:  10,
					Price:     100,
					Currency:  "USD",
				},
				{
					Type:      "buy",
					AssetType: "crypto",
					Ticker:    "BTC",
					Quantity:  1,
					Price:     50000,
					Currency:  "USD",
				},
			},
			expectedAssets: []models.Holding{
				{
					Ticker:       "AAPL",
					Quantity:     10,
					AveragePrice: 100,
					AssetType:    "stock",
					Currency:     "USD",
				},
				{
					Ticker:       "BTC",
					Quantity:     1,
					AveragePrice: 50000,
					AssetType:    "crypto",
					Currency:     "USD",
				},
			},
			expectedError: nil,
		},
		{
			name: "different currencies should be treated as different assets",
			trades: []models.Trade{
				{
					Type:      "buy",
					AssetType: "stock",
					Ticker:    "AAPL",
					Quantity:  10,
					Price:     100,
					Currency:  "USD",
				},
				{
					Type:      "buy",
					AssetType: "stock",
					Ticker:    "AAPL",
					Quantity:  5,
					Price:     80,
					Currency:  "EUR",
				},
			},
			expectedAssets: []models.Holding{
				{
					Ticker:       "AAPL",
					Quantity:     10,
					AveragePrice: 100,
					AssetType:    "stock",
					Currency:     "USD",
				},
				{
					Ticker:       "AAPL",
					Quantity:     5,
					AveragePrice: 80,
					AssetType:    "stock",
					Currency:     "EUR",
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockTradeService := new(MockTradeService)
			mockTradeService.On("ListTrades", "test-user").Return(tt.trades, tt.expectedError)

			// Create service with mock
			service := NewHoldingService(mockTradeService)

			// Call service method
			assets, err := service.ListHoldings("test-user")

			// Assert results
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedAssets, assets)

			// Verify mock expectations
			mockTradeService.AssertExpectations(t)
		})
	}
}
