package models

// Holding represents a user's current asset holdings
type Holding struct {
	Ticker       string  `json:"ticker" db:"ticker"`
	TickerName   string  `json:"tickerName" db:"ticker_name"`
	Quantity     float64 `json:"quantity" db:"quantity"`
	AveragePrice float64 `json:"averagePrice" db:"average_price"`
	Price        float64 `json:"price,omitempty" db:"-"` // Current market price, not stored in DB
	AssetType    string  `json:"assetType" db:"asset_type"`
	Currency     string  `json:"currency" db:"currency"`
}
