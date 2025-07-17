package repositories

import (
	"asset-diary/models"

	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	DeleteUser(userID string) error
	ListAllUserIDs() ([]string, error)
	FindUserByID(userID string) (*models.User, error)
	FindUserByEmail(email string) (*models.User, error)
	FindUserByGoogleID(googleID string) (*models.User, error)
	LinkGoogleAccount(userID, googleID, googleEmail string) error
	UnlinkGoogleAccount(userID string) error
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) DeleteUser(userID string) error {
	result := r.db.Where(&models.User{ID: userID}).Delete(&models.User{})
	return result.Error
}

func (r *UserRepository) ListAllUserIDs() ([]string, error) {
	var users []models.User
	result := r.db.Select("id").Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	userIDs := make([]string, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	return userIDs, nil
}

func (r *UserRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) FindUserByID(userID string) (*models.User, error) {
	var user models.User
	result := r.db.Where("id = ?", userID).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) FindUserByGoogleID(googleID string) (*models.User, error) {
	var user models.User
	result := r.db.Where("google_id = ?", googleID).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) LinkGoogleAccount(userID, googleID, googleEmail string) error {
	user, err := r.FindUserByID(userID)
	if err != nil {
		return err
	}
	user.LinkGoogleAccount(googleID, googleEmail)
	result := r.db.Save(&user)
	return result.Error
}

func (r *UserRepository) UnlinkGoogleAccount(userID string) error {
	user, err := r.FindUserByID(userID)
	if err != nil {
		return err
	}
	user.UnlinkGoogleAccount()
	result := r.db.Save(&user)
	return result.Error
}
