package routes

import (
	"asset-diary/handlers"
	"asset-diary/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine,
	authHandler *handlers.AuthHandler,
	profileHandler *handlers.ProfileHandler,
	accountHandler *handlers.AccountHandler,
	tradeHandler *handlers.TradeHandler,
	holdingHandler *handlers.HoldingHandler,
) {
	// Public routes
	public := r.Group("/auth")
	{
		public.POST("/sign-in", authHandler.SignIn)
		public.POST("/sign-up", authHandler.SignUp)
		public.POST("/refresh", authHandler.RefreshToken)
		public.POST("/forgot-password", authHandler.ForgotPassword)
		public.POST("/verify-reset-code", authHandler.VerifyResetCode)
	}

	// Protected routes
	protected := r.Group("/")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		auth := protected.Group("/auth")
		{
			auth.POST("/logout", authHandler.Logout)
		}

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

		// Asset routes
		protected.GET("/holdings", holdingHandler.ListHoldings)
	}
}
