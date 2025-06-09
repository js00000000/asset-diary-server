package handlers

import (
	"asset-diary/models"
	"asset-diary/services"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func getRefreshTokenExpirySeconds() int {
	// Default to 7 days (in seconds) if not set
	defaultExpiry := 7 * 24 * 3600
	expiryStr := os.Getenv("REFRESH_TOKEN_EXPIRY")
	if expiryStr == "" {
		return defaultExpiry
	}

	duration, err := time.ParseDuration(expiryStr)
	if err != nil {
		return defaultExpiry
	}

	return int(duration.Seconds())
}

type AuthHandler struct {
	authService services.AuthServiceInterface
}

func NewAuthHandler(authService services.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) SignUp(c *gin.Context) {
	var req models.UserSignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.SignUp(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.setRefreshTokenCookie(c, response.RefreshToken)

	c.JSON(http.StatusCreated, gin.H{
		"token":        response.Token,
		"refreshToken": response.RefreshToken,
		"user":         response.User,
	})
}

func (h *AuthHandler) SignIn(c *gin.Context) {
	var req models.UserSignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.SignIn(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.setRefreshTokenCookie(c, response.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"token":        response.Token,
		"refreshToken": response.RefreshToken,
		"user":         response.User,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		// Fallback to cookie for backward compatibility
		var err error
		header, err = c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token missing"})
			return
		}
	} else {
		header = strings.TrimPrefix(header, "Bearer ")
	}
	refreshToken := header

	accessToken, newRefreshToken, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		if err == services.ErrInvalidToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new tokens"})
		}
		return
	}

	h.setRefreshTokenCookie(c, newRefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"token":        accessToken,
		"refreshToken": newRefreshToken,
	})
}

func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, newRefreshToken string) {
	cookieMaxAge := getRefreshTokenExpirySeconds()
	frontendDomain := os.Getenv("FRONTEND_DOMAIN")
	c.SetCookie("refresh_token", newRefreshToken, cookieMaxAge, "/", frontendDomain, true, true)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Get refresh token from Authorization header or cookie
	header := c.GetHeader("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		header, _ = c.Cookie("refresh_token")
	} else {
		header = strings.TrimPrefix(header, "Bearer ")
	}

	if header != "" {
		// Try to revoke the refresh token if it exists
		if err := h.authService.RevokeRefreshToken(header); err != nil {
			// Log the error but don't fail the request
			log.Printf("Failed to revoke refresh token: %v", err)
		}
	}

	// Clear the refresh token cookie
	frontendDomain := os.Getenv("FRONTEND_DOMAIN")
	c.SetCookie("refresh_token", "", -1, "/", frontendDomain, true, true)

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.ForgotPassword(req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) VerifyResetCode(c *gin.Context) {
	var req models.VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.VerifyResetCode(req.Email, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
