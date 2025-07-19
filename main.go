package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"asset-diary/db"
	"asset-diary/handlers"
	"asset-diary/repositories"
	"asset-diary/routes"
	"asset-diary/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Set the timezone to Asia/Taipei (UTC+8)
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		log.Fatalf("Error loading location: %v", err)
	}
	time.Local = loc // Set the default timezone for the application

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	dotenvFile := fmt.Sprintf(".env.%s", env)
	if err := godotenv.Load(dotenvFile); err != nil {
		log.Printf("No %s file found or error loading %s", dotenvFile, dotenvFile)
	}

	dbConn := db.InitDB()

	// Initialize Redis
	redisClient, err := db.InitRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

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

	app := gin.New()

	if env == "production" {
		// In production, trust X-Forwarded-* headers from Netlify
		err = app.SetTrustedProxies([]string{
			"127.0.0.1", // Localhost
			"::1",       // IPv6 localhost
			// Netlify serverless functions IP ranges
			"34.138.0.0/15", // Netlify serverless functions
			"34.149.0.0/16", // Netlify serverless functions
			"35.201.0.0/16", // Netlify serverless functions
		})
	} else {
		// In local development, only trust localhost
		err = app.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	}

	if err != nil {
		log.Printf("Warning: Failed to set trusted proxies: %v", err)
	} else {
		log.Printf("[%s] Proxy configuration applied", env)
	}

	// Middleware
	app.Use(gin.Logger())
	app.Use(gin.Recovery())
	app.Use(func(c *gin.Context) {
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

	// Initialize exchange rate service first since it's needed for holding service
	supportedCurrencies := []string{"TWD", "USD", "USDT"} // Default values
	if currencies := os.Getenv("SUPPORTED_CURRENCIES"); currencies != "" {
		supportedCurrencies = strings.Split(currencies, ",")
	}

	// Get server URL from environment variable or use default
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:3000" // Default value
	}

	// Initialize repositories
	profileRepo := repositories.NewProfileRepository(dbConn)
	tradeRepo := repositories.NewTradeRepository(dbConn)
	accountRepo := repositories.NewAccountRepository(dbConn)
	authRepo := repositories.NewAuthRepository(dbConn)
	userRepo := repositories.NewUserRepository(dbConn)
	priceCacheRepo := repositories.NewPriceCacheRepository(redisClient)
	exchangeRateRepo := repositories.NewExchangeRateRepository(dbConn)
	userDailyTotalAssetValueRepo := repositories.NewUserDailyTotalAssetValueRepository(dbConn)

	// Initialize services
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(authRepo, userService)
	profileService := services.NewProfileService(profileRepo)
	accountService := services.NewAccountService(accountRepo)
	tradeService := services.NewTradeService(tradeRepo)
	geminiChatService := services.NewGeminiChatService()
	geminiAssetPriceService := services.NewGeminiAssetPriceService(geminiChatService)
	assetPriceService := services.NewAssetPriceService()
	fallbackPriceService := services.NewFallbackPriceService(assetPriceService, geminiAssetPriceService)
	assetPriceServiceCacheDecorator := services.NewPriceServiceCacheDecorator(fallbackPriceService, priceCacheRepo)
	exchangeRateService := services.NewExchangeRateService(exchangeRateRepo, supportedCurrencies)
	holdingService := services.NewHoldingService(
		tradeService,
		assetPriceServiceCacheDecorator,
		profileService,
		exchangeRateService,
	)
	dailyAssetService := services.NewDailyTotalAssetValueService(
		userDailyTotalAssetValueRepo,
		holdingService,
		accountService,
		exchangeRateService,
		profileService,
		userService,
	)

	// Initialize handlers
	cronHandler := handlers.NewCronHandler(exchangeRateService, dailyAssetService)
	authHandler := handlers.NewAuthHandler(authService, userService)
	profileHandler := handlers.NewProfileHandler(profileService, userService)
	accountHandler := handlers.NewAccountHandler(accountService, exchangeRateService, profileService)
	tradeHandler := handlers.NewTradeHandler(tradeService)
	holdingHandler := handlers.NewHoldingHandler(holdingService)
	assetPriceHandler := handlers.NewAssetPriceHandler(assetPriceServiceCacheDecorator)
	geminiTestHandler := handlers.NewGeminiTestHandler(geminiChatService, geminiAssetPriceService)
	healthCheckHandler := handlers.NewHealthCheckHandler()
	exchangeRateHandler := handlers.NewExchangeRateHandler(exchangeRateService)
	dailyTotalAssetValueHandler := handlers.NewDailyTotalAssetValueHandler(dailyAssetService)

	// Initialize Redis handler
	redisHandler := handlers.NewRedisHandler()

	// Set up routes with all handlers
	routes.SetupRoutes(app.Group("/api"),
		authHandler,
		profileHandler,
		accountHandler,
		tradeHandler,
		holdingHandler,
		assetPriceHandler,
		geminiTestHandler,
		exchangeRateHandler,
		healthCheckHandler,
		dailyTotalAssetValueHandler,
		cronHandler,
		redisHandler,
	)

	app.GET("/kaithhealthcheck", healthCheckHandler.HealthCheck) // for leapcell
	app.GET("/swagger/*any", ginSwaggerHandler())

	app.Run(":3000")
}

// ginSwaggerHandler is a placeholder for Swagger docs
func ginSwaggerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Swagger docs not implemented yet"})
	}
}
