package services

import (
	"log"
	"time"

	"asset-diary/models"
	"asset-diary/repositories"
)

// DailyTotalAssetValueServiceInterface defines the interface for daily asset service
type DailyTotalAssetValueServiceInterface interface {
	RecordDailyTotalAssetValue() error
	GetUserDailyTotalAssetValues(userID string, startDate, endDate time.Time) ([]models.UserDailyTotalAssetValue, error)
}

// DailyTotalAssetValueService implements DailyTotalAssetValueServiceInterface
type DailyTotalAssetValueService struct {
	dailyAssetRepo repositories.UserDailyTotalAssetValueRepositoryInterface
	holdingSvc     HoldingServiceInterface
	exchangeSvc    ExchangeRateServiceInterface
	profileSvc     ProfileServiceInterface
	userSvc        UserServiceInterface
}

// NewDailyTotalAssetValueService creates a new DailyTotalAssetValueService
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

// RecordDailyTotalAssetValue records the daily asset values for all users
func (s *DailyTotalAssetValueService) RecordDailyTotalAssetValue() error {
	// Get all users (you might want to implement pagination for large user bases)
	// For now, we'll assume we have a method to get all user IDs
	userIDs, err := s.userSvc.GetAllUserIDs()
	if err != nil {
		return err
	}

	today := time.Now().UTC()

	for _, userID := range userIDs {
		err := s.recordUserDailyTotalAssetValues(userID, today)
		if err != nil {
			// Log the error but continue with other users
			log.Printf("Failed to record daily assets for user %s: %v", userID, err)
		}
	}

	return nil
}

// GetUserDailyTotalAssetValues retrieves daily asset records for a user within a date range
func (s *DailyTotalAssetValueService) GetUserDailyTotalAssetValues(userID string, startDate, endDate time.Time) ([]models.UserDailyTotalAssetValue, error) {
	return s.dailyAssetRepo.GetUserDailyTotalAssetValues(userID, startDate, endDate)
}

// recordUserDailyTotalAssetValues records the daily asset value for a single user
func (s *DailyTotalAssetValueService) recordUserDailyTotalAssetValues(userID string, date time.Time) error {
	// Get user's default currency from profile
	profile, err := s.profileSvc.GetProfile(userID)
	if err != nil {
		return err
	}

	if profile.InvestmentProfile == nil || profile.InvestmentProfile.DefaultCurrency == "" {
		log.Printf("User %s has no default currency set, using USD as fallback", userID)
		profile.InvestmentProfile = &models.InvestmentProfile{DefaultCurrency: "USD"}
	}

	// Get user's holdings
	holdings, err := s.holdingSvc.ListHoldings(userID)
	if err != nil {
		return err
	}

	// Calculate total value in user's default currency
	totalValue, err := s.calculateTotalValue(holdings, profile.InvestmentProfile.DefaultCurrency)
	if err != nil {
		return err
	}

	// Create daily asset record
	record := &models.UserDailyTotalAssetValue{
		UserID:     userID,
		Date:       date,
		TotalValue: totalValue,
		Currency:   profile.InvestmentProfile.DefaultCurrency,
	}

	return s.dailyAssetRepo.CreateOrUpdate(record)
}

// calculateTotalValue calculates the total value of holdings in the target currency
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
			// If we don't have the exchange rate, log a warning and skip this holding
			log.Printf("No exchange rate found for %s to %s", holding.Currency, targetCurrency)
			continue
		}
		totalValue += holdingValue / rate
	}

	return totalValue, nil
}
