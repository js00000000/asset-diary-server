package services

import (
	"asset-diary/models"
	"asset-diary/services/interfaces"
	"log"
)

type HoldingServiceInterface interface {
	ListHoldings(userID string) ([]models.Holding, error)
}

type HoldingService struct {
	tradeService TradeServiceInterface
	priceService interfaces.AssetPriceServiceInterface
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

func NewHoldingService(tradeService TradeServiceInterface, priceService interfaces.AssetPriceServiceInterface) *HoldingService {
	return &HoldingService{
		tradeService: tradeService,
		priceService: priceService,
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
			tickerInfo, err := s.getCurrentPrice(asset.Ticker, asset.AssetType)
			if err != nil {
				log.Printf("Error fetching price for %s %s: %v", asset.AssetType, asset.Ticker, err)
				asset.Price = 0
			} else {
				asset.Price = tickerInfo.Price
				// TODO: handle currency conversion
				// If the asset's currency is different from the price's currency, we might want to convert it
				// For now, we'll just use the price as is and handle conversion in the frontend if needed
			}
			assets = append(assets, *asset)
		}
	}

	return assets, nil
}
