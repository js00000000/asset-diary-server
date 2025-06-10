package repositories

import (
	"asset-diary/models"

	"gorm.io/gorm"
)

type ExchangeRateRepositoryInterface interface {
	GetRatesByBaseCurrency(baseCurrency string) ([]*models.ExchangeRate, error)
	Upsert(rate *models.ExchangeRate) error
}

type ExchangeRateRepository struct {
	db *gorm.DB
}

func NewExchangeRateRepository(db *gorm.DB) *ExchangeRateRepository {
	return &ExchangeRateRepository{db: db}
}

func (r *ExchangeRateRepository) GetRatesByBaseCurrency(baseCurrency string) ([]*models.ExchangeRate, error) {
	var rates []*models.ExchangeRate
	err := r.db.Where("base_currency = ?", baseCurrency).
		Order("target_currency").
		Find(&rates).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return rates, nil
}

// Upsert updates an existing exchange rate or creates a new one if it doesn't exist
func (r *ExchangeRateRepository) Upsert(rate *models.ExchangeRate) error {
	return r.db.Where(models.ExchangeRate{
		BaseCurrency:   rate.BaseCurrency,
		TargetCurrency: rate.TargetCurrency,
	}).
		Assign(map[string]interface{}{
			"rate":         rate.Rate,
			"last_updated": rate.LastUpdated,
		}).
		FirstOrCreate(rate).Error
}
