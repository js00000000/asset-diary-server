package main

import (
	"asset-diary/db"
	"asset-diary/handlers"
	"asset-diary/repositories"
	"asset-diary/routes"
	"asset-diary/services"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	dotenvFile := fmt.Sprintf(".env.%s", env)
	if err := godotenv.Load(dotenvFile); err != nil {
		log.Printf("No %s file found or error loading %s", dotenvFile, dotenvFile)
	}

	dbConn := db.InitDB()

	// Run DB migrations before starting the server
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	sqlDB, err := dbConn.DB()
	if err != nil {
		log.Fatalf("Failed to get SQL database: %v", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create DB driver for migrations: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("Failed to create migration instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Database migrated successfully")

	r := gin.Default()

	// Configure trusted proxies based on environment
	if env == "production" {
		// In production, trust X-Forwarded-* headers from Netlify
		err = r.SetTrustedProxies([]string{
			"127.0.0.1",     // Localhost
			"::1",           // IPv6 localhost
			// Netlify serverless functions IP ranges
			"34.138.0.0/15", // Netlify serverless functions
			"34.149.0.0/16", // Netlify serverless functions
			"35.201.0.0/16", // Netlify serverless functions
		})
	} else {
		// In local development, only trust localhost
		err = r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	}

	if err != nil {
		log.Printf("Warning: Failed to set trusted proxies: %v", err)
	} else {
		log.Printf("[%s] Proxy configuration applied", env)
	}

	// CORS middleware for frontend on port 5173
	r.Use(func(c *gin.Context) {
		allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
		originList := strings.Split(allowedOrigins, ",")
		requestOrigin := c.Request.Header.Get("Origin")
		allowed := false
		for _, o := range originList {
			if strings.TrimSpace(o) == requestOrigin {
				allowed = true
				break
			}
		}
		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", requestOrigin)
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Initialize repositories
	profileRepo := repositories.NewProfileRepository(dbConn)
	tradeRepo := repositories.NewTradeRepository(dbConn)
	accountRepo := repositories.NewAccountRepository(dbConn)
	authRepo := repositories.NewAuthRepository(dbConn)
	userRepo := repositories.NewUserRepository(dbConn)
	priceCacheRepo := repositories.NewPriceCacheRepository(dbConn)

	// Initialize services
	authService := services.NewAuthService(authRepo)
	profileService := services.NewProfileService(profileRepo)
	accountService := services.NewAccountService(accountRepo)
	tradeService := services.NewTradeService(tradeRepo)
	holdingService := services.NewHoldingService(tradeService)
	userService := services.NewUserService(userRepo)
	geminiChatService := services.NewGeminiChatService()
	geminiAssetPriceService := services.NewGeminiAssetPriceService(geminiChatService)
	assetPriceService := services.NewAssetPriceService()
	fallbackPriceService := services.NewFallbackPriceService(assetPriceService, geminiAssetPriceService)
	assetPriceServiceCacheDecorator := services.NewPriceServiceCacheDecorator(fallbackPriceService, priceCacheRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	profileHandler := handlers.NewProfileHandler(profileService, userService)
	accountHandler := handlers.NewAccountHandler(accountService)
	tradeHandler := handlers.NewTradeHandler(tradeService)
	holdingHandler := handlers.NewHoldingHandler(holdingService)
	assetPriceHandler := handlers.NewAssetPriceHandler(assetPriceServiceCacheDecorator)
	geminiTestHandler := handlers.NewGeminiTestHandler(geminiChatService, geminiAssetPriceService)

	routes.SetupRoutes(&r.RouterGroup, authHandler, profileHandler, accountHandler, tradeHandler, holdingHandler, assetPriceHandler, geminiTestHandler)

	r.GET("/swagger/*any", ginSwaggerHandler())

	r.Run(":3000")
}

// ginSwaggerHandler is a placeholder for Swagger docs
func ginSwaggerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Swagger docs not implemented yet"})
	}
}
