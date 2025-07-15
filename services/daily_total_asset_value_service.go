package services

import (
	"log"
	"time"

	"asset-diary/models"
	"asset-diary/repositories"
)

type DailyTotalAssetValueServiceInterface interface {
	RecordDailyTotalAssetValue() error
	GetUserDailyTotalAssetValues(userID string, startDate, endDate time.Time) ([]models.UserDailyTotalAssetValue, error)
}

type DailyTotalAssetValueService struct {
	dailyAssetRepo repositories.UserDailyTotalAssetValueRepositoryInterface
	holdingSvc     HoldingServiceInterface
	accountSvc     AccountServiceInterface
	exchangeSvc    ExchangeRateServiceInterface
	profileSvc     ProfileServiceInterface
	userSvc        UserServiceInterface
}

func NewDailyTotalAssetValueService(
	dailyAssetRepo repositories.UserDailyTotalAssetValueRepositoryInterface,
	holdingSvc HoldingServiceInterface,
	accountSvc AccountServiceInterface,
	exchangeSvc ExchangeRateServiceInterface,
	profileSvc ProfileServiceInterface,
	userSvc UserServiceInterface,
) *DailyTotalAssetValueService {
	return &DailyTotalAssetValueService{
		dailyAssetRepo: dailyAssetRepo,
		holdingSvc:     holdingSvc,
		accountSvc:     accountSvc,
		exchangeSvc:    exchangeSvc,
		profileSvc:     profileSvc,
		userSvc:        userSvc,
	}
}

func (s *DailyTotalAssetValueService) RecordDailyTotalAssetValue() error {
	userIDs, err := s.userSvc.GetAllUserIDs()
	if err != nil {
		return err
	}

	today := time.Now().UTC()

	for _, userID := range userIDs {
		err := s.recordUserDailyTotalAssetValues(userID, today)
		if err != nil {
			log.Printf("Failed to record daily assets for user %s: %v", userID, err)
		}
	}

	return nil
}

func (s *DailyTotalAssetValueService) GetUserDailyTotalAssetValues(userID string, startDate, endDate time.Time) ([]models.UserDailyTotalAssetValue, error) {
	return s.dailyAssetRepo.GetUserDailyTotalAssetValues(userID, startDate, endDate)
}

func (s *DailyTotalAssetValueService) recordUserDailyTotalAssetValues(userID string, date time.Time) error {
	defaultCurrency, err := s.profileSvc.GetDefaultCurrency(userID)
	if err != nil {
		return err
	}

	holdings, err := s.holdingSvc.ListHoldings(userID)
	if err != nil {
		return err
	}
	holdingItems := make([]struct {
		Amount   float64
		Currency string
	}, len(holdings))
	for i, h := range holdings {
		holdingItems[i] = struct {
			Amount   float64
			Currency string
		}{
			Amount:   h.Quantity * h.Price,
			Currency: h.Currency,
		}
	}

	accounts, err := s.accountSvc.ListAccounts(userID)
	if err != nil {
		return err
	}
	accountItems := make([]struct {
		Amount   float64
		Currency string
	}, len(accounts))
	for i, a := range accounts {
		accountItems[i] = struct {
			Amount   float64
			Currency string
		}{
			Amount:   a.Balance,
			Currency: a.Currency,
		}
	}

	allItems := append(holdingItems, accountItems...)
	totalValue, err := s.calculateTotalValue(allItems, defaultCurrency)
	if err != nil {
		return err
	}

	record := &models.UserDailyTotalAssetValue{
		UserID:     userID,
		Date:       date,
		TotalValue: totalValue,
		Currency:   defaultCurrency,
	}

	return s.dailyAssetRepo.CreateOrUpdate(record)
}

func (s *DailyTotalAssetValueService) calculateTotalValue(items []struct {
	Amount   float64
	Currency string
}, targetCurrency string) (float64, error) {
	totalValue := 0.0

	rates, err := s.exchangeSvc.GetRatesByBaseCurrency(targetCurrency)
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		if item.Currency == targetCurrency {
			totalValue += item.Amount
			continue
		}

		rate, ok := rates[item.Currency]
		if !ok {
			log.Printf("No exchange rate found for %s to %s", item.Currency, targetCurrency)
			continue
		}
		totalValue += item.Amount / rate
	}

	return totalValue, nil
}
