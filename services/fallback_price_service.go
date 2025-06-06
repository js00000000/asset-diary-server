package services

import (
	"asset-diary/models"
	"asset-diary/services/interfaces"
)

type FallbackPriceService struct {
	primary  interfaces.AssetPriceServiceInterface
	fallback interfaces.AssetPriceServiceInterface
}

func NewFallbackPriceService(
	primary interfaces.AssetPriceServiceInterface,
	fallback interfaces.AssetPriceServiceInterface,
) *FallbackPriceService {
	return &FallbackPriceService{
		primary:  primary,
		fallback: fallback,
	}
}

func (s *FallbackPriceService) GetStockPrice(symbol string) (*models.TickerInfo, error) {
	result, err := s.primary.GetStockPrice(symbol)
	if err != nil {
		if s.fallback != nil && err.Error() != InvalidSymbolError {
			return s.fallback.GetStockPrice(symbol)
		}
		return nil, err
	}
	return result, err
}

func (s *FallbackPriceService) GetCryptoPrice(symbol string) (*models.TickerInfo, error) {
	result, err := s.primary.GetCryptoPrice(symbol)
	if err != nil {
		if s.fallback != nil && err.Error() != InvalidSymbolError {
			return s.fallback.GetCryptoPrice(symbol)
		}
		return nil, err
	}
	return result, err
}
