package services

import (
	"asset-diary/models"
	"asset-diary/services/interfaces"
	"fmt"
	"log"
	"sort"
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
		return nil, fmt.Errorf("invalid asset type: %s", assetType)
	}
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

	assetMap, err := calculateHoldings(trades)
	if err != nil {
		return nil, err
	}

	// Get user's default currency from profile
	defaultCurrency, err := s.profileService.GetDefaultCurrency(userID)
	if err != nil {
		log.Printf("Error getting user profile: %v", err)
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	rates, err := s.exchangeService.GetRatesByBaseCurrency(defaultCurrency)
	if err != nil {
		log.Printf("Error getting exchange rates: %v", err)
		return nil, fmt.Errorf("failed to get exchange rates: %w", err)
	}

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
			asset.GainLoss = asset.TotalValue - asset.TotalCost
			asset.GainLossPercentage = (asset.GainLoss / asset.TotalCost) * 100
			asset.TotalValueInDefaultCurrency = asset.TotalValue / rates[asset.Currency]
			assets = append(assets, *asset)
		}
	}

	return assets, nil
}

func calculateHoldings(trades []models.Trade) (map[string]*models.Holding, error) {
	type fifoBuyTrade struct {
		Quantity float64
		Price    float64
		TradeID  string
	}

	sort.Slice(trades, func(i, j int) bool {
		return trades[i].TradeDate.Before(trades[j].TradeDate)
	})

	buyQueues := make(map[string][]fifoBuyTrade)

	currentHoldings := make(map[string]*models.Holding)

	for _, trade := range trades {
		key := fmt.Sprintf("%s_%s_%s", trade.AssetType, trade.Ticker, trade.Currency)

		if _, ok := currentHoldings[key]; !ok {
			currentHoldings[key] = &models.Holding{
				Ticker:     trade.Ticker,
				TickerName: trade.TickerName,
				AssetType:  trade.AssetType,
				Currency:   trade.Currency,
			}
		}

		holding := currentHoldings[key]

		switch trade.Type {
		case "buy":
			buyQueues[key] = append(buyQueues[key], fifoBuyTrade{
				Quantity: trade.Quantity,
				Price:    trade.Price,
				TradeID:  trade.ID,
			})

			holding.TotalCost += trade.Quantity * trade.Price
			holding.Quantity += trade.Quantity

		case "sell":
			sellQuantity := trade.Quantity

			if holding.Quantity < sellQuantity {
				return nil, fmt.Errorf("holding %s sell quantity %.2f exceeds holding quantity %.2f, trade ID: %s", key, sellQuantity, holding.Quantity, trade.ID)
			}

			for sellQuantity > 0 && len(buyQueues[key]) > 0 {
				oldestBuy := &buyQueues[key][0]

				if oldestBuy.Quantity <= sellQuantity {
					holding.TotalCost -= oldestBuy.Quantity * oldestBuy.Price
					holding.Quantity -= oldestBuy.Quantity
					sellQuantity -= oldestBuy.Quantity
					buyQueues[key] = buyQueues[key][1:] // remove the oldest buy record
				} else {
					holding.TotalCost -= sellQuantity * oldestBuy.Price
					holding.Quantity -= sellQuantity
					oldestBuy.Quantity -= sellQuantity
					sellQuantity = 0
				}
			}
		}

		if holding.Quantity > 0 {
			holding.AverageCost = holding.TotalCost / holding.Quantity
		} else {
			holding.AverageCost = 0
			holding.TotalCost = 0
		}

		currentHoldings[key] = holding
	}

	finalHoldings := make(map[string]*models.Holding)
	for key, h := range currentHoldings {
		if h.Quantity > 0 {
			finalHoldings[key] = h
		}
	}

	return finalHoldings, nil
}
