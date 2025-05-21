package services

import (
	"asset-diary/models"
)

type HoldingServiceInterface interface {
	ListHoldings(userID string) ([]models.Holding, error)
}

type HoldingService struct {
	tradeService TradeServiceInterface
}

// Lot represents a batch of shares bought at a specific price
type Lot struct {
	Quantity     float64
	Price        float64
	RemainingQty float64
}

func NewHoldingService(tradeService TradeServiceInterface) *HoldingService {
	return &HoldingService{
		tradeService: tradeService,
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
				asset.AveragePrice = totalCost / totalRemainingQty
			}
		}
	}

	// Convert map to slice and filter out zero quantity assets
	assets := []models.Holding{}
	for _, asset := range assetMap {
		if asset.Quantity > 0 {
			assets = append(assets, *asset)
		}
	}

	return assets, nil
}
