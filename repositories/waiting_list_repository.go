package repositories

import (
	"asset-diary/models"
	"gorm.io/gorm"
)

type WaitingListRepositoryInterface interface {
	Create(entry *models.WaitingList) error
	IsProjectAllowed(name string) (bool, error)
}

type WaitingListRepository struct {
	db *gorm.DB
}

func NewWaitingListRepository(db *gorm.DB) *WaitingListRepository {
	return &WaitingListRepository{db: db}
}

func (r *WaitingListRepository) Create(entry *models.WaitingList) error {
	return r.db.Create(entry).Error
}

func (r *WaitingListRepository) IsProjectAllowed(name string) (bool, error) {
	var count int64
	err := r.db.Model(&models.AllowedProject{}).Where("name = ? AND is_active = ?", name, true).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
