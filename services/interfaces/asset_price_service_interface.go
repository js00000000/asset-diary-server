package interfaces

import "asset-diary/models"

type AssetPriceServiceInterface interface {
	GetStockPrice(symbol string) (*models.TickerInfo, error)
	GetCryptoPrice(symbol string) (*models.TickerInfo, error)
}
