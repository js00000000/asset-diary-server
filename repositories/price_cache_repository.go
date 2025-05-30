package repositories

import (
	"asset-diary/models"
	"time"

	"gorm.io/gorm"
)

type PriceCacheRepositoryInterface interface {
	Get(cacheKey string) (*models.PriceCache, error)
	Set(cache *models.PriceCache) error
	DeleteExpired() error
}

type PriceCacheRepository struct {
	db *gorm.DB
}

func NewPriceCacheRepository(db *gorm.DB) *PriceCacheRepository {
	return &PriceCacheRepository{db: db}
}

func (r *PriceCacheRepository) Get(cacheKey string) (*models.PriceCache, error) {
	var cache models.PriceCache
	err := r.db.Where("cache_key = ? AND expires_at > ?", cacheKey, time.Now()).First(&cache).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &cache, nil
}

func (r *PriceCacheRepository) Set(cache *models.PriceCache) error {
	// Use upsert to update if exists, insert if not
	return r.db.Where(models.PriceCache{CacheKey: cache.CacheKey}).
		Assign(map[string]interface{}{
			"symbol":     cache.Symbol,
			"name":       cache.Name,
			"price":      cache.Price,
			"currency":   cache.Currency,
			"expires_at": cache.ExpiresAt,
		}).
		FirstOrCreate(cache).Error
}

func (r *PriceCacheRepository) DeleteExpired() error {
	return r.db.Where("expires_at <= ?", time.Now()).Delete(&models.PriceCache{}).Error
}
