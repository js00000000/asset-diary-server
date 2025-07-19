package services

import (
	"asset-diary/models"
	"asset-diary/repositories"
	"asset-diary/services/interfaces"
	"os"
	"strconv"
	"time"
)

type priceServiceCacheDecorator struct {
	service   interfaces.AssetPriceServiceInterface
	cacheRepo repositories.PriceCacheRepositoryInterface
	cacheTTL  time.Duration
}

func NewPriceServiceCacheDecorator(service interfaces.AssetPriceServiceInterface, cacheRepo repositories.PriceCacheRepositoryInterface) *priceServiceCacheDecorator {
	cacheTTL := 5 * time.Minute

	envTTL := os.Getenv("PRICE_CACHE_TTL_MINUTES")
	if envTTL != "" {
		if ttl, err := strconv.Atoi(envTTL); err == nil && ttl > 0 {
			cacheTTL = time.Duration(ttl) * time.Minute
		}
	}

	return &priceServiceCacheDecorator{
		service:   service,
		cacheRepo: cacheRepo,
		cacheTTL:  cacheTTL,
	}
}

func (d *priceServiceCacheDecorator) getFromCache(assetType string, symbol string) (*models.TickerInfo, bool) {
	cached, err := d.cacheRepo.Get(assetType, symbol)
	if err != nil || cached == nil {
		return nil, false
	}

	if time.Now().After(cached.ExpiresAt) {
		return nil, false
	}

	return &models.TickerInfo{
		AssetType:   cached.AssetType,
		Price:       cached.Price,
		Symbol:      cached.Symbol,
		Name:        cached.Name,
		Currency:    cached.Currency,
		LastUpdated: cached.UpdatedAt.Format(time.RFC3339),
	}, true
}

func (d *priceServiceCacheDecorator) setInCache(info *models.TickerInfo) error {
	expiresAt := time.Now().Add(d.cacheTTL)

	cache := &models.PriceCache{
		AssetType: info.AssetType,
		Symbol:    info.Symbol,
		Name:      info.Name,
		Price:     info.Price,
		Currency:  info.Currency,
		ExpiresAt: expiresAt,
	}

	return d.cacheRepo.Set(cache)
}

func (d *priceServiceCacheDecorator) GetStockPrice(symbol string) (*models.TickerInfo, error) {
	if cached, found := d.getFromCache("stock", symbol); found {
		return cached, nil
	}

	info, err := d.service.GetStockPrice(symbol)
	if err != nil {
		return nil, err
	}

	_ = d.setInCache(info)
	return info, nil
}

func (d *priceServiceCacheDecorator) GetCryptoPrice(symbol string) (*models.TickerInfo, error) {
	if cached, found := d.getFromCache("crypto", symbol); found {
		return cached, nil
	}

	info, err := d.service.GetCryptoPrice(symbol)
	if err != nil {
		return nil, err
	}

	_ = d.setInCache(info)
	return info, nil
}
