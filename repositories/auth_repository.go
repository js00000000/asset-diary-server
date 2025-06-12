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
	StoreRefreshToken(userID, tokenHash string, expiresAt time.Time) error
	FindRefreshToken(tokenHash string) (*models.RefreshToken, error)
	RevokeRefreshToken(tokenHash string) error
	RevokeAllUserRefreshTokens(userID string) error
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
		log.Println("Failed to create user:", result.Error)
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

	result := r.DB.Where(&verificationCode).First(&verificationCode)

	if result.Error == gorm.ErrRecordNotFound {
		verificationCode.Code = code
		verificationCode.ExpiresAt = time.Now().Add(expiry)
		result = r.DB.Create(&verificationCode)
	} else {
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

	if time.Now().After(verificationCode.ExpiresAt) {
		return false, nil
	}

	r.DB.Delete(&verificationCode)

	return true, nil
}

func (r *AuthRepository) StoreRefreshToken(userID, tokenHash string, expiresAt time.Time) error {
	token := &models.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	return r.DB.Create(token).Error
}

func (r *AuthRepository) FindRefreshToken(tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.DB.Where("token_hash = ? AND revoked = ? AND expires_at > ?",
		tokenHash, false, time.Now()).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *AuthRepository) RevokeRefreshToken(tokenHash string) error {
	return r.DB.Model(&models.RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Update("revoked", true).Error
}

func (r *AuthRepository) RevokeAllUserRefreshTokens(userID string) error {
	return r.DB.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true).Error
}
