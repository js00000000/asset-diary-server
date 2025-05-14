package handlers

import (
	"asset-diary/models"
	"asset-diary/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProfileHandler handles profile-related HTTP requests
type ProfileHandler struct {
	profileService services.ProfileServiceInterface
	userService    services.UserServiceInterface
}

// NewProfileHandler creates a new ProfileHandler instance
func NewProfileHandler(profileService services.ProfileServiceInterface, userService services.UserServiceInterface) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		userService:    userService,
	}
}

// GetProfile returns the current user's profile
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	profile, err := h.profileService.GetProfile(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if profile.InvestmentProfile != nil {
		c.JSON(http.StatusOK, models.ProfileResponse{
			Email:    profile.Email,
			Name:     profile.Name,
			Username: profile.Username,
			InvestmentProfile: &models.InvestmentProfileResponse{
				Age:                                  profile.InvestmentProfile.Age,
				MaxAcceptableShortTermLossPercentage: profile.InvestmentProfile.MaxAcceptableShortTermLossPercentage,
				ExpectedAnnualizedRateOfReturn:       profile.InvestmentProfile.ExpectedAnnualizedRateOfReturn,
				TimeHorizon:                          profile.InvestmentProfile.TimeHorizon,
				YearsInvesting:                       profile.InvestmentProfile.YearsInvesting,
				MonthlyCashFlow:                      profile.InvestmentProfile.MonthlyCashFlow,
				DefaultCurrency:                      profile.InvestmentProfile.DefaultCurrency,
			},
		})
	} else {
		c.JSON(http.StatusOK, models.ProfileResponse{
			Email:    profile.Email,
			Name:     profile.Name,
			Username: profile.Username,
		})
	}
}

// ChangePassword changes the current user's password
func (h *ProfileHandler) ChangePassword(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.profileService.ChangePassword(userID.(string), req.CurrentPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateProfile updates the current user's profile
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.profileService.UpdateProfile(userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ProfileResponse{
		Name:     profile.Name,
		Email:    profile.Email,
		Username: profile.Username,
		InvestmentProfile: &models.InvestmentProfileResponse{
			Age:                                  profile.InvestmentProfile.Age,
			MaxAcceptableShortTermLossPercentage: profile.InvestmentProfile.MaxAcceptableShortTermLossPercentage,
			ExpectedAnnualizedRateOfReturn:       profile.InvestmentProfile.ExpectedAnnualizedRateOfReturn,
			TimeHorizon:                          profile.InvestmentProfile.TimeHorizon,
			YearsInvesting:                       profile.InvestmentProfile.YearsInvesting,
			MonthlyCashFlow:                      profile.InvestmentProfile.MonthlyCashFlow,
			DefaultCurrency:                      profile.InvestmentProfile.DefaultCurrency,
		},
	})
}

// DeleteProfile deletes the current user's profile and all associated data
func (h *ProfileHandler) DeleteProfile(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := h.userService.DeleteUser(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
