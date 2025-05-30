package services

import (
	"os"
	"strconv"
	"sync"
	"time"
)

type cachedTickerInfo struct {
	info      *TickerInfo
	expiresAt time.Time
}

type priceServiceCacheDecorator struct {
	service  AssetPriceServiceInterface
	cache    map[string]cachedTickerInfo
	cacheTTL time.Duration
	mu       sync.RWMutex
}

// NewPriceServiceCacheDecorator creates a new caching decorator for AssetPriceService
func NewPriceServiceCacheDecorator(service AssetPriceServiceInterface) *priceServiceCacheDecorator {
	// Default cache TTL of 20 minutes
	cacheTTL := 20 * time.Minute
	
	// Try to get TTL from environment
	envTTL := os.Getenv("PRICE_CACHE_TTL_MINUTES")
	if envTTL != "" {
		if ttl, err := strconv.Atoi(envTTL); err == nil && ttl > 0 {
			cacheTTL = time.Duration(ttl) * time.Minute
		}
	}

	return &priceServiceCacheDecorator{
		service:  service,
		cache:    make(map[string]cachedTickerInfo),
		cacheTTL: cacheTTL,
	}
}

func (d *priceServiceCacheDecorator) getFromCache(key string) (*TickerInfo, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	cached, exists := d.cache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(cached.expiresAt) {
		// Cache expired
		return nil, false
	}

	return cached.info, true
}

func (d *priceServiceCacheDecorator) setInCache(key string, info *TickerInfo) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.cache[key] = cachedTickerInfo{
		info:      info,
		expiresAt: time.Now().Add(d.cacheTTL),
	}
}

func (d *priceServiceCacheDecorator) GetStockPrice(symbol string) (*TickerInfo, error) {
	cacheKey := "stock_" + symbol
	if cached, found := d.getFromCache(cacheKey); found {
		return cached, nil
	}

	info, err := d.service.GetStockPrice(symbol)
	if err != nil {
		return nil, err
	}

	d.setInCache(cacheKey, info)
	return info, nil
}

func (d *priceServiceCacheDecorator) GetCryptoPrice(symbol string) (*TickerInfo, error) {
	cacheKey := "crypto_" + symbol
	if cached, found := d.getFromCache(cacheKey); found {
		return cached, nil
	}

	info, err := d.service.GetCryptoPrice(symbol)
	if err != nil {
		return nil, err
	}

	d.setInCache(cacheKey, info)
	return info, nil
}
