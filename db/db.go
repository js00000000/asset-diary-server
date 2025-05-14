package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("Warning: DATABASE_URL not set, falling back to default DSN")
		dsn = "postgres://postgres:postgres@localhost:5432/asset_diary?sslmode=disable"
	}

	// Configure logger
	dbLogger := logger.Default.LogMode(logger.Info)
	if env == "production" {
		dbLogger = logger.Default.LogMode(logger.Error)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// Set global DB instance
	DB = db

	return db
}

// GetDB returns the global DB instance
func GetDB() *gorm.DB {
	return DB
}
