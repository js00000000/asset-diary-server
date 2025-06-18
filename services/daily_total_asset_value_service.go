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
	exchangeSvc    ExchangeRateServiceInterface
	profileSvc     ProfileServiceInterface
	userSvc        UserServiceInterface
}

func NewDailyTotalAssetValueService(
	dailyAssetRepo repositories.UserDailyTotalAssetValueRepositoryInterface,
	holdingSvc HoldingServiceInterface,
	exchangeSvc ExchangeRateServiceInterface,
	profileSvc ProfileServiceInterface,
	userSvc UserServiceInterface,
) *DailyTotalAssetValueService {
	return &DailyTotalAssetValueService{
		dailyAssetRepo: dailyAssetRepo,
		holdingSvc:     holdingSvc,
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

	totalValue, err := s.calculateTotalValue(holdings, defaultCurrency)
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

func (s *DailyTotalAssetValueService) calculateTotalValue(holdings []models.Holding, targetCurrency string) (float64, error) {
	totalValue := 0.0

	rates, err := s.exchangeSvc.GetRatesByBaseCurrency(targetCurrency)
	if err != nil {
		return 0, err
	}

	for _, holding := range holdings {
		holdingValue := holding.Quantity * holding.Price
		rate, ok := rates[holding.Currency]
		if !ok {
			log.Printf("No exchange rate found for %s to %s", holding.Currency, targetCurrency)
			continue
		}
		totalValue += holdingValue / rate
	}

	return totalValue, nil
}
