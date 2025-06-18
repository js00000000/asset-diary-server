package routes

import (
	"asset-diary/handlers"
	"asset-diary/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.RouterGroup,
	authHandler *handlers.AuthHandler,
	profileHandler *handlers.ProfileHandler,
	accountHandler *handlers.AccountHandler,
	tradeHandler *handlers.TradeHandler,
	holdingHandler *handlers.HoldingHandler,
	assetPriceHandler *handlers.AssetPriceHandler,
	geminiTestHandler *handlers.GeminiTestHandler,
	exchangeRateHandler *handlers.ExchangeRateHandler,
	healthCheckHandler *handlers.HealthCheckHandler,
	dailyTotalAssetValueHandler *handlers.DailyTotalAssetValueHandler,
) {
	router.GET("/healthz", healthCheckHandler.HealthCheck)
	public := router.Group("/auth")
	{
		public.POST("/sign-in", authHandler.SignIn)
		public.POST("/sign-up", authHandler.SignUp)
		public.POST("/refresh", authHandler.RefreshToken)
		public.POST("/logout", authHandler.Logout)
		public.POST("/forgot-password", authHandler.ForgotPassword)
		public.POST("/verify-reset-code", authHandler.VerifyResetCode)
	}

	geminiTestHandler.RegisterRoutes(router.Group(""))

	// Public exchange rate endpoints
	exchangeRates := router.Group("/exchange-rates")
	{
		exchangeRates.GET("/:base_currency", exchangeRateHandler.GetRatesByBaseCurrency)
	}

	protected := router.Group("/")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		profile := protected.Group("/profile")
		{
			profile.POST("/change-password", profileHandler.ChangePassword)
			profile.GET("", profileHandler.GetProfile)
			profile.PUT("", profileHandler.UpdateProfile)
			profile.DELETE("", profileHandler.DeleteProfile)
		}

		accounts := protected.Group("/accounts")
		{
			accounts.GET("", accountHandler.ListAccounts)
			accounts.POST("", accountHandler.CreateAccount)
			accounts.PUT("/:id", accountHandler.UpdateAccount)
			accounts.DELETE("/:id", accountHandler.DeleteAccount)
		}

		trades := protected.Group("/trades")
		{
			trades.GET("", tradeHandler.ListTrades)
			trades.POST("", tradeHandler.CreateTrade)
			trades.PUT("/:id", tradeHandler.UpdateTrade)
			trades.DELETE("/:id", tradeHandler.DeleteTrade)
		}

		protected.GET("/holdings", holdingHandler.ListHoldings)
		protected.GET("/stock/price/:symbol", assetPriceHandler.GetStockPrice)
		protected.GET("/crypto/price/:symbol", assetPriceHandler.GetCryptoPrice)
		protected.GET("/daily-total-assets", dailyTotalAssetValueHandler.GetUserDailyTotalAssetValues)
	}
}
