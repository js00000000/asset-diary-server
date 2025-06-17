package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"asset-diary/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SQLiteUserDailyTotalAssetValue is a simplified version of UserDailyTotalAssetValue for SQLite testing
type SQLiteUserDailyTotalAssetValue struct {
	ID         string    `gorm:"primaryKey"`
	UserID     string    `gorm:"not null;index"`
	Date       time.Time `gorm:"not null"`
	TotalValue float64   `gorm:"not null"`
	Currency   string    `gorm:"not null;size:3"`
	CreatedAt  time.Time `gorm:"not null;default:current_timestamp"`
}

// TableName specifies the table name for the SQLite model
func (SQLiteUserDailyTotalAssetValue) TableName() string {
	return "user_daily_assets"
}

// sqliteDailyAssetRepo is a custom repository for SQLite testing
type sqliteDailyAssetRepo struct {
	db *gorm.DB
}

// CreateOrUpdate implements the repository interface
func (r *sqliteDailyAssetRepo) CreateOrUpdate(asset *models.UserDailyTotalAssetValue) error {
	sqliteAsset := &SQLiteUserDailyTotalAssetValue{
		ID:         asset.ID,
		UserID:     asset.UserID,
		Date:       asset.Date,
		TotalValue: asset.TotalValue,
		Currency:   asset.Currency,
		CreatedAt:  asset.CreatedAt,
	}

	var existing SQLiteUserDailyTotalAssetValue
	result := r.db.Where("user_id = ? AND date = ?", asset.UserID, asset.Date.Format("2006-01-02")).First(&existing)
	if result.Error == nil {
		// Update existing
		return r.db.Model(&SQLiteUserDailyTotalAssetValue{}).
			Where("id = ?", existing.ID).
			Updates(map[string]interface{}{
				"total_value": sqliteAsset.TotalValue,
				"currency":    sqliteAsset.Currency,
			}).Error
	} else if result.Error == gorm.ErrRecordNotFound {
		// Create new
		return r.db.Create(sqliteAsset).Error
	}
	return result.Error
}

// GetUserDailyTotalAssetValues implements the repository interface
func (r *sqliteDailyAssetRepo) GetUserDailyTotalAssetValues(userID string, startDate, endDate time.Time) ([]models.UserDailyTotalAssetValue, error) {
	var assets []SQLiteUserDailyTotalAssetValue
	query := r.db.Where("user_id = ? AND date BETWEEN ? AND ?",
		userID,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02")).Order("date")
	err := query.Find(&assets).Error
	if err != nil {
		return nil, err
	}

	result := make([]models.UserDailyTotalAssetValue, len(assets))
	for i, a := range assets {
		result[i] = models.UserDailyTotalAssetValue{
			ID:         a.ID,
			UserID:     a.UserID,
			Date:       a.Date,
			TotalValue: a.TotalValue,
			Currency:   a.Currency,
			CreatedAt:  a.CreatedAt,
		}
	}
	return result, nil
}

// GetLatestUserDailyTotalAssetValue implements the repository interface
func (r *sqliteDailyAssetRepo) GetLatestUserDailyTotalAssetValue(userID string) (*models.UserDailyTotalAssetValue, error) {
	var asset SQLiteUserDailyTotalAssetValue
	err := r.db.Where("user_id = ?", userID).Order("date DESC").First(&asset).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &models.UserDailyTotalAssetValue{
		ID:         asset.ID,
		UserID:     asset.UserID,
		Date:       asset.Date,
		TotalValue: asset.TotalValue,
		Currency:   asset.Currency,
		CreatedAt:  asset.CreatedAt,
	}, nil
}

func main() {
	// Initialize in-memory SQLite database
	dbConn, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create a simple table for testing
	err = dbConn.AutoMigrate(&SQLiteUserDailyTotalAssetValue{})
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Create repository instance
	dailyAssetRepo := &sqliteDailyAssetRepo{db: dbConn}

	// Create a test user ID for testing
	testUserID := "test-user-123"

	// Test recording daily assets for the test user
	log.Println("Testing daily asset recording...")

	// Manually create test records
	today := time.Now()
	testRecords := []*models.UserDailyTotalAssetValue{
		{
			ID:         "test-id-1",
			UserID:     testUserID,
			Date:       today.AddDate(0, 0, -2), // 2 days ago
			TotalValue: 1400.0,
			Currency:   "USD",
			CreatedAt:  time.Now(),
		},
		{
			ID:         "test-id-2",
			UserID:     testUserID,
			Date:       today.AddDate(0, 0, -1), // yesterday
			TotalValue: 1450.0,
			Currency:   "USD",
			CreatedAt:  time.Now(),
		},
		{
			ID:         "test-id-3",
			UserID:     testUserID,
			Date:       today,  // today
			TotalValue: 1505.0, // 10 * 150.5 (from our mock holdings)
			Currency:   "USD",
			CreatedAt:  time.Now(),
		},
	}

	// Save test records
	for _, record := range testRecords {
		err = dailyAssetRepo.CreateOrUpdate(record)
		if err != nil {
			log.Fatalf("Error creating test record: %v", err)
		}
	}

	log.Printf("Successfully recorded %d test daily assets!", len(testRecords))

	// Query and print the recorded data
	startDate := today.AddDate(0, 0, -7) // Last 7 days
	records, err := dailyAssetRepo.GetUserDailyTotalAssetValues(testUserID, startDate, today)
	if err != nil {
		log.Printf("Error fetching records: %v", err)
		os.Exit(1)
	}

	log.Printf("\n=== Daily Asset Snapshot Test Results ===")
	log.Printf("Found %d records for user %s\n", len(records), testUserID)

	log.Println("Date         | Total Value | Currency")
	log.Println("------------------------------------")
	for _, r := range records {
		log.Printf("%-12s | %11.2f | %s",
			r.Date.Format("2006-01-02"),
			r.TotalValue,
			r.Currency,
		)
	}
	log.Println("\nTest completed successfully!")
}

// Mock services for testing

type mockHoldingService struct {
	holdings []models.Holding
}

func (m *mockHoldingService) ListHoldings(userID string) ([]models.Holding, error) {
	return m.holdings, nil
}

type mockProfileService struct {
	profile *models.Profile
}

func (m *mockProfileService) GetProfile(userID string) (*models.Profile, error) {
	return m.profile, nil
}

func (m *mockProfileService) ChangePassword(userID string, currentPassword, newPassword string) error {
	// For testing purposes, we'll just return nil
	return nil
}

func (m *mockProfileService) UpdateProfile(userID string, req *models.UserUpdateRequest) (*models.Profile, error) {
	// For testing purposes, we'll just return the existing profile
	return m.profile, nil
}
