package repositories

import (
	"log"

	"asset-diary/models"

	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"
)

// ProfileRepositoryInterface defines methods for profile-related database operations
type ProfileRepositoryInterface interface {
	GetProfile(userID string) (*models.Profile, error)
	ChangePassword(userID string, currentPassword, newPassword string) error
	UpdateProfile(userID string, req *models.UserUpdateRequest) (*models.Profile, error)
}

// ProfileRepository implements ProfileRepositoryInterface
type ProfileRepository struct {
	db *gorm.DB
}

// NewProfileRepository creates a new ProfileRepository instance
func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

// GetProfile retrieves user profile and investment profile from the database
func (r *ProfileRepository) GetProfile(userID string) (*models.Profile, error) {
	var user models.User
	result := r.db.Where(&models.User{ID: userID}).First(&user)
	if result.Error != nil {
		log.Println("Failed to fetch user:", result.Error)
		return nil, result.Error
	}

	// Fetch investment profile
	var investmentProfile models.InvestmentProfile
	result = r.db.Where(&models.InvestmentProfile{UserID: userID}).First(&investmentProfile)
	if result.Error == gorm.ErrRecordNotFound {
		return &models.Profile{
			Email:    user.Email,
			Username: user.Username,
		}, nil
	}

	if result.Error != nil {
		log.Println("Failed to fetch investment profile:", result.Error)
		return nil, result.Error
	}

	return &models.Profile{
		Email:    user.Email,
		Username: user.Username,
		InvestmentProfile: &models.InvestmentProfile{
			Age:                                  int(investmentProfile.Age),
			MaxAcceptableShortTermLossPercentage: int(investmentProfile.MaxAcceptableShortTermLossPercentage),
			ExpectedAnnualizedRateOfReturn:       int(investmentProfile.ExpectedAnnualizedRateOfReturn),
			TimeHorizon:                          investmentProfile.TimeHorizon,
			YearsInvesting:                       int(investmentProfile.YearsInvesting),
			MonthlyCashFlow:                      investmentProfile.MonthlyCashFlow,
			DefaultCurrency:                      investmentProfile.DefaultCurrency,
		},
	}, nil
}

// ChangePassword updates the user's password after verifying the current password
func (r *ProfileRepository) ChangePassword(userID string, currentPassword, newPassword string) error {
	var user models.User
	result := r.db.Where(&models.User{ID: userID}).First(&user)
	if result.Error != nil {
		log.Println("Failed to find user:", result.Error)
		return result.Error
	}

	// Check current password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password_Hash), []byte(currentPassword))
	if err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	user.Password_Hash = string(hashedPassword)
	result = r.db.Save(&user)
	if result.Error != nil {
		log.Println("Failed to update password:", result.Error)
		return result.Error
	}

	return nil
}

func (r *ProfileRepository) UpdateProfile(userID string, req *models.UserUpdateRequest) (*models.Profile, error) {
	// Find and update user
	var user models.User
	result := r.db.Where(&models.User{ID: userID}).First(&user)
	if result.Error != nil {
		log.Println("Failed to find user:", result.Error)
		return nil, result.Error
	}

	// Update fields
	user.Username = req.Username

	result = r.db.Save(&user)
	if result.Error != nil {
		log.Println("Failed to update user:", result.Error)
		return nil, result.Error
	}

	// Upsert investment profile
	if req.InvestmentProfile != nil {
		// Check if investment profile already exists
		var existingProfile models.InvestmentProfile
		result = r.db.Where(&models.InvestmentProfile{UserID: userID}).First(&existingProfile)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new investment profile
			newProfile := models.InvestmentProfile{
				UserID:                               userID,
				Age:                                  int(req.InvestmentProfile.Age),
				MaxAcceptableShortTermLossPercentage: int(req.InvestmentProfile.MaxAcceptableShortTermLossPercentage),
				ExpectedAnnualizedRateOfReturn:       int(req.InvestmentProfile.ExpectedAnnualizedRateOfReturn),
				TimeHorizon:                          req.InvestmentProfile.TimeHorizon,
				YearsInvesting:                       int(req.InvestmentProfile.YearsInvesting),
				MonthlyCashFlow:                      req.InvestmentProfile.MonthlyCashFlow,
				DefaultCurrency:                      req.InvestmentProfile.DefaultCurrency,
			}
			existingProfile = newProfile
			result = r.db.Create(&newProfile)
		} else if result.Error == nil {
			// Update existing investment profile
			existingProfile.Age = int(req.InvestmentProfile.Age)
			existingProfile.MaxAcceptableShortTermLossPercentage = int(req.InvestmentProfile.MaxAcceptableShortTermLossPercentage)
			existingProfile.ExpectedAnnualizedRateOfReturn = int(req.InvestmentProfile.ExpectedAnnualizedRateOfReturn)
			existingProfile.TimeHorizon = req.InvestmentProfile.TimeHorizon
			existingProfile.YearsInvesting = int(req.InvestmentProfile.YearsInvesting)
			existingProfile.MonthlyCashFlow = req.InvestmentProfile.MonthlyCashFlow
			existingProfile.DefaultCurrency = req.InvestmentProfile.DefaultCurrency
			result = r.db.Save(&existingProfile)
		} else {
			log.Println("Failed to process investment profile:", result.Error)
			return nil, result.Error
		}

		if result.Error != nil {
			log.Println("Failed to save investment profile:", result.Error)
			return nil, result.Error
		}

		return &models.Profile{
			Email:    user.Email,
			Username: user.Username,
			InvestmentProfile: &models.InvestmentProfile{
				Age:                                  int(existingProfile.Age),
				MaxAcceptableShortTermLossPercentage: int(existingProfile.MaxAcceptableShortTermLossPercentage),
				ExpectedAnnualizedRateOfReturn:       int(existingProfile.ExpectedAnnualizedRateOfReturn),
				TimeHorizon:                          existingProfile.TimeHorizon,
				YearsInvesting:                       int(existingProfile.YearsInvesting),
				MonthlyCashFlow:                      existingProfile.MonthlyCashFlow,
				DefaultCurrency:                      existingProfile.DefaultCurrency,
			},
		}, nil
	}

	return &models.Profile{
		Email:    user.Email,
		Username: user.Username,
	}, nil
}
