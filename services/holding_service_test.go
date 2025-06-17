package services

import (
	"asset-diary/models"
	"fmt"
	"sort"
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

type MockProfileService struct {
	mock.Mock
}

func (m *MockProfileService) GetProfile(userID string) (*models.Profile, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Profile), args.Error(1)
}

func (m *MockProfileService) ChangePassword(userID string, currentPassword, newPassword string) error {
	args := m.Called(userID, currentPassword, newPassword)
	return args.Error(0)
}

func (m *MockProfileService) UpdateProfile(userID string, req *models.UserUpdateRequest) (*models.Profile, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Profile), args.Error(1)
}

func (m *MockProfileService) GetDefaultCurrency(userID string) (string, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

type MockExchangeRateService struct {
	mock.Mock
}

func (m *MockExchangeRateService) GetRatesByBaseCurrency(baseCurrency string) (map[string]float64, error) {
	args := m.Called(baseCurrency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *MockExchangeRateService) FetchAndStoreRates() error {
	args := m.Called()
	return args.Error(0)
}

type MockPriceService struct {
	mock.Mock
}

func (m *MockPriceService) GetStockPrice(symbol string) (*models.TickerInfo, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TickerInfo), args.Error(1)
}

func (m *MockPriceService) GetCryptoPrice(symbol string) (*models.TickerInfo, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TickerInfo), args.Error(1)
}

// byTicker implements sort.Interface for []models.Holding based on the Ticker field
type byTicker []models.Holding

func (a byTicker) Len() int           { return len(a) }
func (a byTicker) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTicker) Less(i, j int) bool { return a[i].Ticker < a[j].Ticker }

func TestListHoldings(t *testing.T) {
	tests := []struct {
		name           string
		trades         []models.Trade
		expectedAssets []models.Holding
		expectedError  error
		setupMocks     func(*MockPriceService)
	}{
		{
			name:           "no trades should return empty list",
			trades:         []models.Trade{},
			expectedAssets: []models.Holding{},
			expectedError:  nil,
			setupMocks:     func(*MockPriceService) {},
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
			setupMocks:     func(*MockPriceService) {},
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
					Ticker:      "AAPL",
					Quantity:    10,
					AverageCost: (5*100 + 5*300) / 10,
					Price:       100,
					AssetType:   "stock",
					Currency:    "USD",
				},
			},
			expectedError: nil,
			setupMocks:    func(*MockPriceService) {},
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
					Ticker:      "AAPL",
					Quantity:    10,
					AverageCost: (5*100 + 5*200) / 10,
					Price:       100,
					AssetType:   "stock",
					Currency:    "USD",
				},
			},
			expectedError: nil,
			setupMocks:    func(*MockPriceService) {},
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
				{Ticker: "AAPL", Quantity: 10, AverageCost: 100, Price: 100, AssetType: "stock", Currency: "USD"},
				{Ticker: "BTC", Quantity: 1, AverageCost: 50000, Price: 200, AssetType: "crypto", Currency: "USD"},
			},
			expectedError: nil,
			setupMocks:    func(*MockPriceService) {},
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
				{Ticker: "AAPL", Quantity: 10, AverageCost: 100, Price: 100, AssetType: "stock", Currency: "USD"},
				{Ticker: "AAPL", Quantity: 5, AverageCost: 80, Price: 100, AssetType: "stock", Currency: "EUR"},
			},
			expectedError: nil,
			setupMocks:    func(*MockPriceService) {},
		},
		{
			name: "price service returns error",
			trades: []models.Trade{
				{
					Type:      "buy",
					AssetType: "stock",
					Ticker:    "AAPL",
					Quantity:  10,
					Price:     100,
					Currency:  "USD",
				},
			},
			expectedAssets: nil,
			expectedError:  fmt.Errorf("failed to get price"),
			setupMocks: func(ps *MockPriceService) {
				ps.On("GetStockPrice", "AAPL").Return((*models.TickerInfo)(nil), fmt.Errorf("price service error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockTradeService := new(MockTradeService)
			mockTradeService.On("ListTrades", "user1").Return(tt.trades, tt.expectedError)

			priceService := new(MockPriceService)
			// Set up mock responses for price service based on test case
			for _, trade := range tt.trades {
				switch trade.AssetType {
				case "stock":
					if tt.expectedError == nil {
						priceService.On("GetStockPrice", trade.Ticker).Return(&models.TickerInfo{Price: 100.0}, nil)
					}
				case "crypto":
					if tt.expectedError == nil {
						priceService.On("GetCryptoPrice", trade.Ticker).Return(&models.TickerInfo{Price: 200.0}, nil)
					}
				}
			}
			// Run any test-specific mock setup
			tt.setupMocks(priceService)

			// Set up mock profile service
			mockProfileService := new(MockProfileService)
			mockProfileService.On("GetProfile", "user1").Return(&models.Profile{
				InvestmentProfile: &models.InvestmentProfile{
					DefaultCurrency: "USD",
				},
			}, nil)

			// Set up mock exchange rate service
			mockExchangeService := new(MockExchangeRateService)
			mockExchangeService.On("GetRatesByBaseCurrency", "USD").Return(map[string]float64{
				"TWD": 30.0,
				"USD": 1.0,
			}, nil)

			service := NewHoldingService(mockTradeService, priceService, mockProfileService, mockExchangeService)

			// Call the method under test
			holdings, err := service.ListHoldings("user1")

			// Assert results
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, holdings)
			} else {
				assert.NoError(t, err)
				// Sort both slices by ticker and currency before comparison
				sort.Slice(holdings, func(i, j int) bool {
					if holdings[i].Ticker == holdings[j].Ticker {
						return holdings[i].Currency < holdings[j].Currency
					}
					return holdings[i].Ticker < holdings[j].Ticker
				})
				sort.Slice(tt.expectedAssets, func(i, j int) bool {
					if tt.expectedAssets[i].Ticker == tt.expectedAssets[j].Ticker {
						return tt.expectedAssets[i].Currency < tt.expectedAssets[j].Currency
					}
					return tt.expectedAssets[i].Ticker < tt.expectedAssets[j].Ticker
				})

				// Compare each field individually for better test failure messages
				assert.Equal(t, len(tt.expectedAssets), len(holdings), "number of holdings mismatch")
				for i := range tt.expectedAssets {
					expected := tt.expectedAssets[i]
					actual := holdings[i]
					assert.Equal(t, expected.Ticker, actual.Ticker, "ticker mismatch")
					assert.Equal(t, expected.Quantity, actual.Quantity, "quantity mismatch for "+expected.Ticker)
					assert.Equal(t, expected.AverageCost, actual.AverageCost, "average price mismatch for "+expected.Ticker)
					assert.Equal(t, expected.AssetType, actual.AssetType, "asset type mismatch for "+expected.Ticker)
					assert.Equal(t, expected.Currency, actual.Currency, "currency mismatch for "+expected.Ticker)
				}
			}

			// Verify mock expectations
			mockTradeService.AssertExpectations(t)
		})
	}
}
