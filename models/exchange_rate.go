package models

import (
	"time"

	"gorm.io/gorm"
)

type ExchangeRate struct {
	gorm.Model
	BaseCurrency   string    `json:"base_currency" gorm:"not null;index"`
	TargetCurrency string    `json:"target_currency" gorm:"not null;index"`
	Rate          float64   `json:"rate" gorm:"type:float8;not null"`
	LastUpdated   time.Time `json:"last_updated" gorm:"not null"`
}
