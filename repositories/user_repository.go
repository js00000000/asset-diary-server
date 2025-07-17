package repositories

import (
	"asset-diary/models"

	"gorm.io/gorm"
)

// UserRepositoryInterface 定義了使用者和個人檔案資料庫操作的介面
type UserRepositoryInterface interface {
	DeleteUser(userID string) error
	ListAllUserIDs() ([]string, error)
	FindUserByEmail(email string) (*models.User, error)
}

// UserRepository 實作了 UserRepositoryInterface
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 建立 UserRepository 實例
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// DeleteUser 刪除使用者及其相關資料
func (r *UserRepository) DeleteUser(userID string) error {
	// Delete user - cascade will handle related records
	result := r.db.Where(&models.User{ID: userID}).Delete(&models.User{})
	return result.Error
}

// ListAllUserIDs 獲取所有使用者的 ID
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

// FindUserByEmail 根據電子郵件查找用戶
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
