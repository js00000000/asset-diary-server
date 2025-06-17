package models

type Holding struct {
	Ticker                      string  `json:"ticker" db:"ticker"`
	TickerName                  string  `json:"tickerName" db:"ticker_name"`
	Quantity                    float64 `json:"quantity" db:"quantity"`
	AveragePrice                float64 `json:"averagePrice" db:"average_price"`
	Price                       float64 `json:"price,omitempty" db:"-"`
	AssetType                   string  `json:"assetType" db:"asset_type"`
	Currency                    string  `json:"currency" db:"currency"`
	TotalCost                   float64 `json:"totalCost" db:"total_cost"`
	TotalValue                  float64 `json:"totalValue" db:"total_value"`
	TotalValueInDefaultCurrency float64 `json:"totalValueInDefaultCurrency" db:"total_value_in_default_currency"`
	GainLoss                    float64 `json:"gainLoss" db:"gain_loss"`
	GainLossPercentage          float64 `json:"gainLossPercentage" db:"gain_loss_percentage"`
}
