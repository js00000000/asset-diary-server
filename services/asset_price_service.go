package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"asset-diary/models"
)

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

func (s *AssetPriceService) GetStockPrice(symbol string) (*models.TickerInfo, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))

	if matched, _ := regexp.MatchString(`^\d`, symbol); matched {
		return s.getTaiwanStockPrice(symbol)
	}
	return s.getUSStockPrice(symbol)
}

func (s *AssetPriceService) GetCryptoPrice(symbol string) (*models.TickerInfo, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	return s.getCryptoPrice(symbol)
}

func (s *AssetPriceService) getTaiwanStockPrice(symbol string) (*models.TickerInfo, error) {
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
		return nil, errors.New(InvalidSymbolError)
	}

	stockInfo := data.MsgArray[0]
	price, err := strconv.ParseFloat(stockInfo.Z, 64)
	if err != nil || price <= 0 {
		return nil, errors.New(InvalidSymbolError)
	}

	return &models.TickerInfo{
		AssetType:   "stock",
		Price:       price,
		Symbol:      symbol,
		Name:        stockInfo.N,
		Currency:    "TWD",
		LastUpdated: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *AssetPriceService) getCryptoPrice(symbol string) (*models.TickerInfo, error) {
	url := fmt.Sprintf("https://data-api.binance.vision/api/v3/ticker/price?symbol=%sUSDT", symbol)

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
		if resp.StatusCode == http.StatusBadRequest {
			return nil, errors.New(InvalidSymbolError)
		}
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

	return &models.TickerInfo{
		AssetType:   "crypto",
		Price:       price,
		Symbol:      symbol,
		Name:        symbol,
		Currency:    "USDT",
		LastUpdated: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *AssetPriceService) getUSStockPrice(symbol string) (*models.TickerInfo, error) {
	apiKey := os.Getenv("FMP_API_KEY")
	url := fmt.Sprintf("https://financialmodelingprep.com/stable/quote?symbol=%s&apikey=%s", symbol, apiKey)

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
		return nil, errors.New(InvalidSymbolError)
	}

	quote := quotes[0]
	return &models.TickerInfo{
		AssetType:   "stock",
		Price:       quote.Price,
		Symbol:      quote.Symbol,
		Name:        quote.Name,
		Currency:    "USD",
		LastUpdated: time.Now().Format(time.RFC3339),
	}, nil
}
