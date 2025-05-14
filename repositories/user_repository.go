package repositories

import (
	"asset-diary/models"

	"gorm.io/gorm"
)

// UserRepositoryInterface 定義了使用者和個人檔案資料庫操作的介面
type UserRepositoryInterface interface {
	DeleteUser(userID string) error
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
