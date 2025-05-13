package models

import "time"

// Trade represents a trade transaction.
type Trade struct {
	ID        string    `gorm:"primaryKey;type:uuid" json:"id" db:"id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id" db:"user_id"`
	User      User      `gorm:"foreignKey:UserID;references:ID;onUpdate:CASCADE;onDelete:CASCADE" json:"user"`
	Type      string    `gorm:"not null" json:"type" db:"type"`            // buy or sell
	AssetType string    `gorm:"not null" json:"assetType" db:"asset_type"` // stock or crypto
	Ticker    string    `gorm:"not null" json:"ticker" db:"ticker"`
	TradeDate time.Time `gorm:"not null" json:"tradeDate" db:"trade_date"`
	Quantity  float64   `gorm:"not null" json:"quantity" db:"quantity"`
	Price     float64   `gorm:"not null" json:"price" db:"price"`
	Currency  string    `gorm:"not null" json:"currency" db:"currency"` // e.g., USD, TWD
	AccountID string    `gorm:"type:uuid;not null;index" json:"accountId" db:"account_id"`
	Account   Account   `gorm:"foreignKey:AccountID;references:ID;onUpdate:CASCADE" json:"account"`
	Reason    *string   `gorm:"nullable" json:"reason,omitempty" db:"reason"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp" json:"createdAt" db:"created_at"`
}

func (Trade) TableName() string {
	return "trades"
}

// TradeCreateRequest for creating a trade
// (optional: can be used for binding in handlers)
type TradeCreateRequest struct {
	Type      string  `json:"type" binding:"required,oneof=buy sell"`
	AssetType string  `json:"assetType" binding:"required,oneof=stock crypto"`
	Ticker    string  `json:"ticker" binding:"required"`
	TradeDate string  `json:"tradeDate" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
	Currency  string  `json:"currency" binding:"required"`
	AccountID string  `json:"accountId" binding:"required"`
	Reason    *string `json:"reason"`
}

type TradeUpdateRequest struct {
	Type      string  `json:"type" binding:"omitempty,oneof=buy sell"`
	AssetType string  `json:"assetType" binding:"omitempty,oneof=stock crypto"`
	Ticker    string  `json:"ticker" binding:"omitempty"`
	TradeDate string  `json:"tradeDate" binding:"omitempty"`
	Quantity  float64 `json:"quantity" binding:"omitempty"`
	Price     float64 `json:"price" binding:"omitempty"`
	Currency  string  `json:"currency" binding:"omitempty"`
	AccountID string  `json:"accountId" binding:"omitempty"`
	Reason    *string `json:"reason"`
}

type TradeResponse struct {
	ID        string    `json:"id" db:"id"`
	Type      string    `json:"type" db:"type"`            // buy or sell
	AssetType string    `json:"assetType" db:"asset_type"` // stock or crypto
	Ticker    string    `json:"ticker" db:"ticker"`
	TradeDate time.Time `json:"tradeDate" db:"trade_date"`
	Quantity  float64   `json:"quantity" db:"quantity"`
	Price     float64   `json:"price" db:"price"`
	Currency  string    `json:"currency" db:"currency"` // e.g., USD, TWD
	AccountID string    `json:"accountId" db:"account_id"`
	Reason    *string   `json:"reason,omitempty" db:"reason"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
