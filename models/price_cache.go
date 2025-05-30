package models

import "time"

type PriceCache struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CacheKey  string    `gorm:"uniqueIndex;not null" json:"cache_key"`
	Symbol    string    `gorm:"not null" json:"symbol"`
	Name      string    `gorm:"not null" json:"name"`
	Price     float64   `gorm:"type:decimal(24,8);not null" json:"price"`
	Currency  string    `gorm:"not null" json:"currency"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
