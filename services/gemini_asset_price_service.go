package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"asset-diary/models"
)

// GeminiAssetPriceService is an implementation of interfaces.AssetPriceServiceInterface that uses Gemini
type GeminiAssetPriceService struct {
	geminiService *GeminiChatService
}

// NewGeminiAssetPriceService creates a new instance of GeminiAssetPriceService
func NewGeminiAssetPriceService(geminiService *GeminiChatService) *GeminiAssetPriceService {
	return &GeminiAssetPriceService{
		geminiService: geminiService,
	}
}

// marketCurrencies maps market codes to their expected currency codes
var marketCurrencies = map[string]string{
	"US": "USD",
	"TW": "TWD",
}

// GetStockPrice gets the price of a stock using Gemini
func (s *GeminiAssetPriceService) GetStockPrice(symbol string) (*models.TickerInfo, error) {
	market := "US"
	isTaiwanStock, _ := regexp.MatchString(`^\d`, symbol)
	if isTaiwanStock {
		market = "TW"
	}
	expectedCurrency, ok := marketCurrencies[market]
	if !ok {
		expectedCurrency = "USD" // default to USD for unknown markets
	}

	prompt := fmt.Sprintf(`today is %s, I will provide a stock symbol (e.g., "AAPL", "BTC", "Tesla") and a market identifier (e.g., TW for Taiwan, US for United States). Please return the real-time price data in the following JSON format, using the provided symbol in the symbol field (or standardizing it, e.g., Tesla to TSLA).  Return ONLY the specified JSON and NO additional text or explanations:

{
  "price": <current price, number format, rounded to two decimal places>,
  "symbol": "<standardized stock symbol, uppercase>",
  "name": "<full name of the company, string format>",
  "currency": "<currency of the price, e.g., USD, EUR, etc.>",
}

Requirements:
1. price must be a number, formatted to up to four decimal places without trailing zeros (e.g., 200.63, 530.32, not 530.3200). For assets with very small prices (e.g., PEPE or other low-priced stocks/cryptocurrencies), use higher precision (e.g., 8 decimal places, such as 0.00001234) based on the asset's typical price scale.
2. symbol must use the provided input code (uppercase). If the input is a company name (e.g., "Tesla"), convert it to the standard stock code (e.g., "TSLA"). Examples:
   - Input "AAPL" → symbol: "AAPL"
   - Input "BTC" → symbol: "BTC"
   - Input "Tesla" → symbol: "TSLA"
3. name must be the official full name of the company (e.g., "Tesla" → "Tesla, Inc.", "BTC" → "Bitcoin").
4. currency must reflect the actual currency of the asset's price (e.g., USD for AAPL, USD for BTC, or appropriate currency for other assets).
5. If the input code or name is invalid for the specified market or data cannot be retrieved, return EXACTLY the following JSON and NOTHING else: { "error": "Invalid stock or asset code for the specified market, or data unavailable" }.
6. Ensure the JSON is valid, with no extra spaces or newlines.

Please generate the JSON response based on the input stock code and market appended below:
Market: %s, Input code: %s`, time.Now().Format(time.RFC3339), market, strings.ToUpper(symbol))

	response, err := s.geminiService.GenerateContentWithJSON(prompt)
	if err != nil {
		return nil, err
	}

	var tickerInfo models.TickerInfo
	if err := json.Unmarshal([]byte(response), &tickerInfo); err != nil {
		panic(err)
	}

	if tickerInfo.Currency != expectedCurrency {
		return nil, errors.New(InvalidSymbolError)
	}

	tickerInfo.LastUpdated = time.Now().Format(time.RFC3339)

	return &tickerInfo, nil
}

// GetCryptoPrice gets the price of a cryptocurrency using Gemini
func (s *GeminiAssetPriceService) GetCryptoPrice(symbol string) (*models.TickerInfo, error) {
	prompt := fmt.Sprintf(`today is %s, I will provide a crypto symbol (e.g., "BTC", "ETH", "SOL"). Please return the real-time price data in the following JSON format, using the provided symbol in the symbol field:

{
  "price": <current price, number format, rounded to two decimal places>,
  "symbol": "<standardized crypto symbol, uppercase>",
  "name": "<full name of the crypto, string format>",
  "currency": "<currency of the price, e.g., USD, EUR, etc.>",
}

Requirements:
1. price must be a number, formatted to up to four decimal places without trailing zeros (e.g., 200.63, 530.32, not 530.3200). For assets with very small prices (e.g., PEPE or other low-priced stocks/cryptocurrencies), use higher precision (e.g., 8 decimal places, such as 0.00001234) based on the asset's typical price scale.
2. symbol must use the provided input symbol (uppercase).
3. name must be the official full name of the crypto (e.g., "BTC" → "Bitcoin").
4. currency must reflect the actual currency of the asset's price (e.g., USDT for BTC).
5. If the input symbol is invalid or data cannot be retrieved, return an error message in the format: {"error": "Invalid crypto or asset code, or data unavailable"}.
6. Ensure the JSON is valid, with no extra spaces or newlines.

Please generate the JSON response based on the input crypto symbol appended below:
Input symbol: %s`, time.Now().Format(time.RFC3339), strings.ToUpper(symbol))

	response, err := s.geminiService.GenerateContentWithJSON(prompt)
	if err != nil {
		return nil, err
	}

	var tickerInfo models.TickerInfo
	if err := json.Unmarshal([]byte(response), &tickerInfo); err != nil {
		panic(err)
	}

	tickerInfo.LastUpdated = time.Now().Format(time.RFC3339)

	return &tickerInfo, nil
}
