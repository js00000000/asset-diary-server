package models

import "time"

type PriceCache struct {
	AssetType string    `json:"asset_type"`
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	ExpiresAt time.Time `json:"expires_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *PriceCache) GetRedisKey() string {
	return p.AssetType + "_" + p.Symbol
}
