package services

import (
	"asset-diary/models"
	"asset-diary/repositories"
	"asset-diary/services/interfaces"
	"fmt"
	"os"
	"strconv"
	"time"
)

type priceServiceCacheDecorator struct {
	service   interfaces.AssetPriceServiceInterface
	cacheRepo repositories.PriceCacheRepositoryInterface
	cacheTTL  time.Duration
}

// NewPriceServiceCacheDecorator creates a new caching decorator for AssetPriceService
func NewPriceServiceCacheDecorator(service interfaces.AssetPriceServiceInterface, cacheRepo repositories.PriceCacheRepositoryInterface) *priceServiceCacheDecorator {
	// Default cache TTL of 20 minutes
	cacheTTL := 20 * time.Minute

	// Try to get TTL from environment
	envTTL := os.Getenv("PRICE_CACHE_TTL_MINUTES")
	if envTTL != "" {
		if ttl, err := strconv.Atoi(envTTL); err == nil && ttl > 0 {
			cacheTTL = time.Duration(ttl) * time.Minute
		}
	}

	// Start a background goroutine to clean up expired cache entries
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			_ = cacheRepo.DeleteExpired()
		}
	}()

	return &priceServiceCacheDecorator{
		service:   service,
		cacheRepo: cacheRepo,
		cacheTTL:  cacheTTL,
	}
}

func (d *priceServiceCacheDecorator) getFromCache(key string) (*models.TickerInfo, bool) {
	cached, err := d.cacheRepo.Get(key)
	if err != nil || cached == nil {
		return nil, false
	}

	if time.Now().After(cached.ExpiresAt) {
		// Cache expired
		return nil, false
	}

	return &models.TickerInfo{
		Price:       cached.Price,
		Symbol:      cached.Symbol,
		Name:        cached.Name,
		Currency:    cached.Currency,
		LastUpdated: cached.UpdatedAt.Format(time.RFC3339),
	}, true
}

func (d *priceServiceCacheDecorator) setInCache(key string, info *models.TickerInfo) error {
	expiresAt := time.Now().Add(d.cacheTTL)
	fmt.Println("Setting cache for key:", key)

	cache := &models.PriceCache{
		CacheKey:  key,
		Symbol:    info.Symbol,
		Name:      info.Name,
		Price:     info.Price,
		Currency:  info.Currency,
		ExpiresAt: expiresAt,
	}

	return d.cacheRepo.Set(cache)
}

func (d *priceServiceCacheDecorator) GetStockPrice(symbol string) (*models.TickerInfo, error) {
	cacheKey := "stock_" + symbol
	if cached, found := d.getFromCache(cacheKey); found {
		return cached, nil
	}

	info, err := d.service.GetStockPrice(symbol)
	if err != nil {
		return nil, err
	}

	_ = d.setInCache(cacheKey, info)
	return info, nil
}

func (d *priceServiceCacheDecorator) GetCryptoPrice(symbol string) (*models.TickerInfo, error) {
	cacheKey := "crypto_" + symbol
	if cached, found := d.getFromCache(cacheKey); found {
		return cached, nil
	}

	info, err := d.service.GetCryptoPrice(symbol)
	if err != nil {
		return nil, err
	}

	_ = d.setInCache(cacheKey, info)
	return info, nil
}
