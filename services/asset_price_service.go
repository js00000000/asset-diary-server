package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type AssetPriceServiceInterface interface {
	GetStockPrice(symbol string) (*TickerInfo, error)
	GetCryptoPrice(baseCurrency string) (*TickerInfo, error)
}

type AssetPriceService struct {
	httpClient *http.Client
}

func NewAssetPriceService() *AssetPriceService {
	return &AssetPriceService{
		httpClient: &http.Client{},
	}
}

type TaiwanStockResponse struct {
	MsgArray []struct {
		C string `json:"c"` // stock code
		N string `json:"n"` // stock name
		Z string `json:"z"` // latest price
	} `json:"msgArray"`
}

type TickerInfo struct {
	Price  float64 `json:"price"`
	Ticker string  `json:"ticker"`
	Name   string  `json:"name"`
}

func (s *AssetPriceService) GetStockPrice(symbol string) (*TickerInfo, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))

	// Check if it's a Taiwan stock (starts with a number)
	if matched, _ := regexp.MatchString(`^\d`, symbol); matched {
		return s.getTaiwanStockPrice(symbol)
	}

	// Otherwise, treat as US stock
	return s.getUSStockPrice(symbol)
}

// GetCryptoPrice gets the price of a cryptocurrency in USDT
func (s *AssetPriceService) GetCryptoPrice(baseCurrency string) (*TickerInfo, error) {
	if baseCurrency == "" {
		return nil, fmt.Errorf("base currency is required")
	}

	// Always use USDT as the quote currency
	quoteCurrency := "USDT"
	symbol := fmt.Sprintf("%s%s", baseCurrency, quoteCurrency)
	return s.getCryptoPrice(symbol)
}

func (s *AssetPriceService) getTaiwanStockPrice(symbol string) (*TickerInfo, error) {
	targetUrl := fmt.Sprintf("https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch=tse_%s.tw&_=CURRENT_TIME", symbol)
	proxyUrl := fmt.Sprintf("https://api.allorigins.win/raw?url=%s", targetUrl)

	req, err := http.NewRequest("GET", proxyUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Taiwan stock data: %s", resp.Status)
	}

	var data TaiwanStockResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if len(data.MsgArray) == 0 {
		return nil, fmt.Errorf("no data found for symbol: %s", symbol)
	}

	stockInfo := data.MsgArray[0]
	price, err := strconv.ParseFloat(stockInfo.Z, 64)
	if err != nil || price <= 0 {
		return nil, fmt.Errorf("invalid price for symbol: %s", symbol)
	}

	return &TickerInfo{
		Price:  price,
		Ticker: symbol,
		Name:   stockInfo.N,
	}, nil
}

func (s *AssetPriceService) getCryptoPrice(symbol string) (*TickerInfo, error) {
	// Convert symbol to Binance format (e.g., BTC-USDT -> BTCUSDT)
	pair := strings.ReplaceAll(symbol, "-", "")
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", pair)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch crypto price: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch crypto price: %s", resp.Status)
	}

	var result struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode crypto price response: %v", err)
	}

	price, err := strconv.ParseFloat(result.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid price format: %v", err)
	}

	// Extract base currency (e.g., BTC from BTC-USDT)
	baseCurrency := strings.Split(symbol, "-")[0]

	return &TickerInfo{
		Price:  price,
		Ticker: baseCurrency,
		Name:   baseCurrency,
	}, nil
}

func (s *AssetPriceService) getUSStockPrice(symbol string) (*TickerInfo, error) {
	// Using Financial Modeling Prep API
	apiKey := os.Getenv("FMP_API_KEY") // In production, this should be in environment variables
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/quote/%s?apikey=%s", symbol, apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch US stock data: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var quotes []struct {
		Symbol string  `json:"symbol"`
		Name   string  `json:"name"`
		Price  float64 `json:"price"`
	}

	if err := json.Unmarshal(body, &quotes); err != nil {
		return nil, err
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("no data found for symbol: %s", symbol)
	}

	quote := quotes[0]
	if quote.Price <= 0 {
		return nil, fmt.Errorf("invalid price for symbol: %s", symbol)
	}

	return &TickerInfo{
		Price:  quote.Price,
		Ticker: quote.Symbol,
		Name:   quote.Name,
	}, nil
}
