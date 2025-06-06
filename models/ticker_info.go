package models

type TickerInfo struct {
	Price       float64 `json:"price"`
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	Currency    string  `json:"currency"`
	LastUpdated string  `json:"lastUpdated"`
}
