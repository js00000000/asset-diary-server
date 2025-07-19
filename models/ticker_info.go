package models

type TickerInfo struct {
	AssetType   string  `json:"asset_type"`
	Price       float64 `json:"price"`
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	Currency    string  `json:"currency"`
	LastUpdated string  `json:"lastUpdated"`
}

func (t *TickerInfo) GetRedisKey() string {
	return t.AssetType + "_" + t.Symbol
}
