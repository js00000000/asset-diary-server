package services

import (
	"asset-diary/models"
	"asset-diary/repositories"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type ExchangeRateServiceInterface interface {
	FetchAndStoreRates() error
	GetRatesByBaseCurrency(baseCurrency string) (map[string]float64, error)
}

type ExchangeRateService struct {
	repo repositories.ExchangeRateRepositoryInterface
}

func NewExchangeRateService(repo repositories.ExchangeRateRepositoryInterface) *ExchangeRateService {
	return &ExchangeRateService{repo: repo}
}

type ExchangeRateResponse struct {
	Result      string             `json:"result"`
	Provider    string             `json:"provider"`
	BaseCode    string             `json:"base_code"`
	Rates       map[string]float64 `json:"rates"`
	LastUpdated int64              `json:"time_last_update_unix"`
}

func (s *ExchangeRateService) GetRatesByBaseCurrency(baseCurrency string) (map[string]float64, error) {
	rates, err := s.repo.GetRatesByBaseCurrency(baseCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rates: %w", err)
	}

	result := make(map[string]float64)
	for _, rate := range rates {
		result[rate.TargetCurrency] = rate.Rate
	}

	return result, nil
}

func (s *ExchangeRateService) FetchAndStoreRates() error {
	// TWD is the base currency we're interested in
	baseCurrency := "TWD"
	compareCurrencies := []string{"USD"}
	url := fmt.Sprintf("https://open.er-api.com/v6/latest/%s", baseCurrency)

	// Fetch the latest rates
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch exchange rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse ExchangeRateResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResponse.Result != "success" {
		return fmt.Errorf("API returned non-success status: %s", apiResponse.Result)
	}

	lastUpdated := time.Unix(apiResponse.LastUpdated, 0).UTC()
	var successCount int

	// Store each currency pair
	for _, currency := range compareCurrencies {
		rate, exists := apiResponse.Rates[currency]
		if !exists {
			log.Printf("Warning: No rate found for currency %s", currency)
			continue
		}

		exchangeRate := &models.ExchangeRate{
			BaseCurrency:   baseCurrency,
			TargetCurrency: currency,
			Rate:           rate,
			LastUpdated:    lastUpdated,
		}

		// Store or update in database
		if err := s.repo.Upsert(exchangeRate); err != nil {
			log.Printf("Failed to update exchange rate for %s/%s: %v", baseCurrency, currency, err)
			continue
		}

		successCount++
	}

	log.Printf("Successfully updated %d exchange rates for %s at %s\n",
		successCount, baseCurrency, lastUpdated.Format(time.RFC3339))

	if successCount == 0 {
		return fmt.Errorf("failed to store any exchange rates")
	}

	return nil
}
