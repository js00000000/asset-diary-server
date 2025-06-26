package models

type Holding struct {
	Ticker                      string  `json:"ticker"`
	TickerName                  string  `json:"tickerName"`
	Quantity                    float64 `json:"quantity"`
	AverageCost                 float64 `json:"averageCost"`
	Price                       float64 `json:"price"`
	AssetType                   string  `json:"assetType"`
	Currency                    string  `json:"currency"`
	TotalCost                   float64 `json:"totalCost"`
	TotalValue                  float64 `json:"totalValue"`
	TotalValueInDefaultCurrency float64 `json:"totalValueInDefaultCurrency"`
	GainLoss                    float64 `json:"gainLoss"`
	GainLossPercentage          float64 `json:"gainLossPercentage"`
}
