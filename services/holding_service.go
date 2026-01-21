package services

import (
	"asset-diary/models"
	"asset-diary/services/interfaces"
	"fmt"
	"log"
	"sort"
	"sync"
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

	trades, err := s.tradeService.ListTrades(userID)
	if err != nil {
		return nil, err
	}

	tradesMap := make(map[string][]models.Trade)
	for _, trade := range trades {
		key := fmt.Sprintf("%s_%s_%s", trade.AssetType, trade.Ticker, trade.Currency)
		tradesMap[key] = append(tradesMap[key], trade)
	}

	holdings := make(map[string]*models.Holding)
	for _, trades := range tradesMap {
		key := fmt.Sprintf("%s_%s_%s", trades[0].AssetType, trades[0].Ticker, trades[0].Currency)

		holding, err := calculateHolding(trades)
		if err != nil {
			log.Printf("Error calculating holding for trade %s: %v", trades[0].ID, err)
			continue
		}

		if holding.Quantity > 0 {
			holdings[key] = holding
		}
	}

	assets := make([]models.Holding, 0, len(holdings))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, holding := range holdings {
		wg.Add(1)
		go func(h *models.Holding) {
			defer wg.Done()

			tickerInfo, err := s.getCurrentPrice(h.Ticker, h.AssetType)
			if err != nil {
				log.Printf("Error fetching price for %s %s: %v", h.AssetType, h.Ticker, err)
				h.Price = 0
				h.TotalValue = 0
			} else {
				h.Price = tickerInfo.Price
				h.TotalValue = h.Price * h.Quantity
			}
			h.TotalCost = h.AverageCost * h.Quantity
			h.GainLoss = h.TotalValue - h.TotalCost
			if h.TotalCost > 0 {
				h.GainLossPercentage = (h.GainLoss / h.TotalCost) * 100
			}
			if rate, ok := rates[h.Currency]; ok && rate > 0 {
				h.TotalValueInDefaultCurrency = h.TotalValue / rate
			}

			mu.Lock()
			assets = append(assets, *h)
			mu.Unlock()
		}(holding)
	}

	wg.Wait()

	return assets, nil
}

func calculateHolding(trades []models.Trade) (*models.Holding, error) {
	if len(trades) == 0 {
		return nil, fmt.Errorf("no trades provided")
	}

	holding := &models.Holding{
		Ticker:     trades[0].Ticker,
		TickerName: trades[0].TickerName,
		AssetType:  trades[0].AssetType,
		Currency:   trades[0].Currency,
	}

	sort.Slice(trades, func(i, j int) bool {
		return trades[i].TradeDate.Before(trades[j].TradeDate)
	})

	type fifoBuyTrade struct {
		Quantity float64
		Price    float64
	}

	var buyQueue []fifoBuyTrade

	for _, trade := range trades {
		switch trade.Type {
		case "buy":
			buyQueue = append(buyQueue, fifoBuyTrade{
				Quantity: trade.Quantity,
				Price:    trade.Price,
			})
			holding.TotalCost += trade.Quantity * trade.Price
			holding.Quantity += trade.Quantity
			holding.AverageCost = holding.TotalCost / holding.Quantity

		case "sell":
			if trade.Quantity <= 0 {
				return nil, fmt.Errorf("sell quantity must be positive, got %.2f", trade.Quantity)
			}

			sellQuantity := trade.Quantity
			if holding.Quantity < sellQuantity {
				return nil, fmt.Errorf("insufficient quantity to sell %s, attempted to sell %.2f but only have %.2f",
					trade.Ticker, sellQuantity, holding.Quantity)
			}

			for sellQuantity > 0 && len(buyQueue) > 0 {
				oldestBuy := &buyQueue[0]
				if oldestBuy.Quantity <= sellQuantity {
					holding.TotalCost -= oldestBuy.Quantity * oldestBuy.Price
					sellQuantity -= oldestBuy.Quantity
					buyQueue = buyQueue[1:]
				} else {
					holding.TotalCost -= sellQuantity * oldestBuy.Price
					oldestBuy.Quantity -= sellQuantity
					sellQuantity = 0
				}
			}

			holding.Quantity -= trade.Quantity
			if holding.Quantity > 0 {
				holding.AverageCost = holding.TotalCost / holding.Quantity
			} else {
				holding.AverageCost = 0
				holding.TotalCost = 0
			}

		default:
			return nil, fmt.Errorf("unsupported trade type: %s", trade.Type)
		}
	}

	return holding, nil
}
