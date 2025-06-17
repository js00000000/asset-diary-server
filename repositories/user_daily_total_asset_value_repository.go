package repositories

import (
	"log"
	"time"

	"asset-diary/models"

	"gorm.io/gorm"
)

// UserDailyTotalAssetValueRepositoryInterface defines the interface for daily asset repository
type UserDailyTotalAssetValueRepositoryInterface interface {
	CreateOrUpdate(record *models.UserDailyTotalAssetValue) error
	GetUserDailyTotalAssetValues(userID string, startDate, endDate time.Time) ([]models.UserDailyTotalAssetValue, error)
	GetLatestUserDailyTotalAssetValue(userID string) (*models.UserDailyTotalAssetValue, error)
}

// UserDailyTotalAssetValueRepository implements DailyAssetRepositoryInterface
type UserDailyTotalAssetValueRepository struct {
	db *gorm.DB
}

// NewUserDailyTotalAssetValueRepository creates a new UserDailyTotalAssetValueRepository
func NewUserDailyTotalAssetValueRepository(db *gorm.DB) *UserDailyTotalAssetValueRepository {
	return &UserDailyTotalAssetValueRepository{db: db}
}

// CreateOrUpdate creates a new daily asset record or updates if it already exists for the user and date
func (r *UserDailyTotalAssetValueRepository) CreateOrUpdate(record *models.UserDailyTotalAssetValue) error {
	// Use ON CONFLICT to update the total_value if the record already exists
	result := r.db.Exec(`
		INSERT INTO user_daily_total_asset_values (user_id, date, total_value, currency, updated_at)
		VALUES (?, ?, ?, ?, NOW())
		ON CONFLICT (user_id, date) 
		DO UPDATE SET 
			total_value = EXCLUDED.total_value,
			updated_at = EXCLUDED.updated_at
	`,
		record.UserID,
		record.Date.Format("2006-01-02"),
		record.TotalValue,
		record.Currency,
	)

	if result.Error != nil {
		log.Printf("Failed to create/update daily asset record: %v", result.Error)
		return result.Error
	}

	return nil
}

// GetUserDailyTotalAssetValues retrieves daily asset records for a user within a date range
func (r *UserDailyTotalAssetValueRepository) GetUserDailyTotalAssetValues(userID string, startDate, endDate time.Time) ([]models.UserDailyTotalAssetValue, error) {
	var assets []models.UserDailyTotalAssetValue
	result := r.db.Where("user_id = ? AND date BETWEEN ? AND ?",
		userID,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	).Order("date ASC").Find(&assets)

	if result.Error != nil {
		log.Printf("Failed to get user daily assets: %v", result.Error)
		return nil, result.Error
	}

	return assets, nil
}

// GetLatestUserDailyTotalAssetValue gets the most recent daily asset record for a user
func (r *UserDailyTotalAssetValueRepository) GetLatestUserDailyTotalAssetValue(userID string) (*models.UserDailyTotalAssetValue, error) {
	var asset models.UserDailyTotalAssetValue
	result := r.db.Where("user_id = ?", userID).Order("date DESC").First(&asset)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Printf("Failed to get latest user daily asset: %v", result.Error)
		return nil, result.Error
	}

	return &asset, nil
}
