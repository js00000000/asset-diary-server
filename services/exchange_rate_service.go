package services

import (
	"asset-diary/models"
	"asset-diary/repositories"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type ExchangeRateServiceInterface interface {
	FetchAndStoreRates() error
	GetRatesByBaseCurrency(baseCurrency string) (map[string]float64, error)
}

type ExchangeRateService struct {
	repo                repositories.ExchangeRateRepositoryInterface
	supportedCurrencies []string
}

func NewExchangeRateService(repo repositories.ExchangeRateRepositoryInterface, supportedCurrencies []string) *ExchangeRateService {
	return &ExchangeRateService{
		repo:                repo,
		supportedCurrencies: supportedCurrencies,
	}
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
	var totalSuccessCount int
	var lastErr error

	for _, baseCurrency := range s.supportedCurrencies {
		// For each base currency, we'll compare against other base currencies
		compareCurrencies := []string{}
		for _, other := range s.supportedCurrencies {
			if other != baseCurrency {
				compareCurrencies = append(compareCurrencies, other)
			}
		}

		if len(compareCurrencies) == 0 {
			continue // Skip if no currencies to compare
		}

		url := fmt.Sprintf("https://open.er-api.com/v6/latest/%s", baseCurrency)
		resp, err := http.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("failed to fetch exchange rates for %s: %w", baseCurrency, err)
			log.Println(lastErr)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status code for %s: %d", baseCurrency, resp.StatusCode)
			log.Println(lastErr)
			continue
		}

		var apiResponse ExchangeRateResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("failed to decode response for %s: %w", baseCurrency, err)
			log.Println(lastErr)
			continue
		}
		resp.Body.Close()

		if apiResponse.Result != "success" {
			lastErr = fmt.Errorf("api returned non-success status for %s: %s", baseCurrency, apiResponse.Result)
			log.Println(lastErr)
			continue
		}

		lastUpdated := time.Unix(apiResponse.LastUpdated, 0)
		successCount := 0

		// TODO: remove this after USDT is added to the API
		// add USDT to api Response same rate with USD but base on base currency
		if baseCurrency == "USD" {
			apiResponse.Rates["USDT"] = 1.0
		} else {
			apiResponse.Rates["USDT"] = apiResponse.Rates["USD"]
		}

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
			log.Printf("Successfully updated exchange rate for %s/%s at %s\n",
				baseCurrency, currency, lastUpdated.Format(time.RFC3339))
		}

		totalSuccessCount += successCount
	}

	if totalSuccessCount == 0 && lastErr != nil {
		return fmt.Errorf("failed to store any exchange rates. Last error: %w", lastErr)
	}

	if lastErr != nil {
		// We had some successes but also some failures
		log.Printf("Partial success: %v\n", lastErr)
	}

	return nil
}
