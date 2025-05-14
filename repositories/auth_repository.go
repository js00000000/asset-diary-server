package repositories

import (
	"log"
	"time"

	"asset-diary/models"

	"gorm.io/gorm"
)

type AuthRepositoryInterface interface {
	CreateUser(user *models.User, hashedPassword string) error
	FindUserByEmail(email string) (*models.User, error)
	StoreVerificationCode(email, code string, expiry time.Duration) error
	ValidateVerificationCode(email, code string) (bool, error)
	UpdateUserPassword(email, hashedPassword string) error
}

type AuthRepository struct {
	DB *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{DB: db}
}

func (r *AuthRepository) CreateUser(user *models.User, hashedPassword string) error {
	user.Password_Hash = hashedPassword

	result := r.DB.Create(user)
	if result.Error != nil {
		log.Println("Failed to insert user:", result.Error)
		return result.Error
	}
	return nil
}

func (r *AuthRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.DB.Where(&models.User{Email: email}).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) UpdateUserPassword(email, hashedPassword string) error {
	var user models.User
	result := r.DB.Where(&models.User{Email: email}).First(&user)
	if result.Error != nil {
		log.Println("Failed to find user:", result.Error)
		return result.Error
	}

	user.Password_Hash = hashedPassword
	result = r.DB.Save(&user)
	if result.Error != nil {
		log.Println("Failed to update password:", result.Error)
		return result.Error
	}
	return nil
}

func (r *AuthRepository) StoreVerificationCode(email, code string, expiry time.Duration) error {
	verificationCode := models.VerificationCode{
		Email: email,
	}

	// First, try to find an existing verification code for this email
	result := r.DB.Where(&verificationCode).First(&verificationCode)

	// Update or create the verification code
	if result.Error == gorm.ErrRecordNotFound {
		// If not found, create a new record
		verificationCode.Code = code
		verificationCode.ExpiresAt = time.Now().Add(expiry)
		result = r.DB.Create(&verificationCode)
	} else {
		// If found, update the existing record
		verificationCode.Code = code
		verificationCode.ExpiresAt = time.Now().Add(expiry)
		result = r.DB.Save(&verificationCode)
	}

	if result.Error != nil {
		log.Println("Failed to store verification code:", result.Error)
		return result.Error
	}
	return nil
}

func (r *AuthRepository) ValidateVerificationCode(email, code string) (bool, error) {
	var verificationCode models.VerificationCode
	result := r.DB.Where(&models.VerificationCode{
		Email: email,
		Code:  code,
	}).First(&verificationCode)

	if result.Error != nil {
		return false, result.Error
	}

	// Check if the verification code has expired
	if time.Now().After(verificationCode.ExpiresAt) {
		return false, nil
	}

	// Optional: Delete the verification code after successful validation
	r.DB.Delete(&verificationCode)

	return true, nil
}
