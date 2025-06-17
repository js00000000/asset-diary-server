package services

import (
	"asset-diary/models"
	"asset-diary/services/interfaces"
	"fmt"
	"log"
)

type HoldingServiceInterface interface {
	ListHoldings(userID string) ([]models.Holding, error)
}

type HoldingService struct {
	tradeService    TradeServiceInterface
	priceService    interfaces.AssetPriceServiceInterface
	profileService  ProfileServiceInterface
	exchangeService ExchangeRateServiceInterface
}

func (s *HoldingService) getCurrentPrice(ticker, assetType string) (*models.TickerInfo, error) {
	switch assetType {
	case "stock":
		return s.priceService.GetStockPrice(ticker)
	case "crypto":
		return s.priceService.GetCryptoPrice(ticker)
	default:
		return nil, nil
	}
}

// Lot represents a batch of shares bought at a specific price
type Lot struct {
	Quantity     float64
	Price        float64
	RemainingQty float64
}

func NewHoldingService(
	tradeService TradeServiceInterface,
	priceService interfaces.AssetPriceServiceInterface,
	profileService ProfileServiceInterface,
	exchangeService ExchangeRateServiceInterface,
) *HoldingService {
	return &HoldingService{
		tradeService:    tradeService,
		priceService:    priceService,
		profileService:  profileService,
		exchangeService: exchangeService,
	}
}

func (s *HoldingService) ListHoldings(userID string) ([]models.Holding, error) {
	trades, err := s.tradeService.ListTrades(userID)
	if err != nil {
		return nil, err
	}

	// Map to track assets by ticker and currency
	assetMap := make(map[string]*models.Holding)
	// Map to track lots for each asset
	lotsMap := make(map[string][]*Lot)

	for _, trade := range trades {
		key := trade.Ticker + "_" + trade.Currency
		asset, exists := assetMap[key]
		if !exists {
			asset = &models.Holding{
				Ticker:     trade.Ticker,
				TickerName: trade.TickerName,
				AssetType:  trade.AssetType,
				Currency:   trade.Currency,
			}
			assetMap[key] = asset
			lotsMap[key] = []*Lot{}
		}

		if trade.Type == "buy" {
			// Add new lot for buy
			lot := &Lot{
				Quantity:     trade.Quantity,
				Price:        trade.Price,
				RemainingQty: trade.Quantity,
			}
			lotsMap[key] = append(lotsMap[key], lot)
			asset.Quantity += trade.Quantity
		} else if trade.Type == "sell" {
			// Implement FIFO for sells
			remainingSellQty := trade.Quantity
			for _, lot := range lotsMap[key] {
				if remainingSellQty <= 0 {
					break
				}
				if lot.RemainingQty > 0 {
					if lot.RemainingQty >= remainingSellQty {
						lot.RemainingQty -= remainingSellQty
						remainingSellQty = 0
					} else {
						remainingSellQty -= lot.RemainingQty
						lot.RemainingQty = 0
					}
				}
			}
			asset.Quantity -= trade.Quantity
		}
	}

	// Calculate average price based on remaining lots
	for key, asset := range assetMap {
		if asset.Quantity > 0 {
			var totalCost float64
			var totalRemainingQty float64
			for _, lot := range lotsMap[key] {
				if lot.RemainingQty > 0 {
					totalCost += lot.Price * lot.RemainingQty
					totalRemainingQty += lot.RemainingQty
				}
			}
			if totalRemainingQty > 0 {
				asset.AverageCost = totalCost / totalRemainingQty
			}
		}
	}

	// Get user's default currency from profile
	profile, err := s.profileService.GetProfile(userID)
	if err != nil {
		log.Printf("Error getting user profile: %v", err)
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	defaultCurrency := "USD" // Default fallback
	if profile.InvestmentProfile != nil && profile.InvestmentProfile.DefaultCurrency != "" {
		defaultCurrency = profile.InvestmentProfile.DefaultCurrency
	}

	// Get exchange rates for all involved currencies
	rates, err := s.exchangeService.GetRatesByBaseCurrency(defaultCurrency)
	if err != nil {
		log.Printf("Error getting exchange rates: %v", err)
		return nil, fmt.Errorf("failed to get exchange rates: %w", err)
	}
	log.Printf("Exchange rates: %v", rates)

	// Convert map to slice and filter out zero quantity assets
	assets := []models.Holding{}
	for _, asset := range assetMap {
		if asset.Quantity > 0 {
			tickerInfo, err := s.getCurrentPrice(asset.Ticker, asset.AssetType)
			if err != nil {
				log.Printf("Error fetching price for %s %s: %v", asset.AssetType, asset.Ticker, err)
				asset.Price = 0
				asset.TotalValue = 0
			} else {
				asset.Price = tickerInfo.Price
				asset.TotalValue = asset.Price * asset.Quantity
			}
			asset.TotalCost = asset.AverageCost * asset.Quantity
			asset.TotalValueInDefaultCurrency = asset.TotalValue
			asset.GainLoss = asset.TotalValue - asset.TotalCost
			asset.GainLossPercentage = (asset.GainLoss / asset.TotalCost) * 100
			if asset.Currency != defaultCurrency {
				asset.TotalValueInDefaultCurrency = asset.TotalValue / rates[asset.Currency]
			}
			assets = append(assets, *asset)
		}
	}

	return assets, nil
}
