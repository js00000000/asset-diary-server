package repositories

import (
	"asset-diary/models"
	"gorm.io/gorm"
)

type WaitingListRepositoryInterface interface {
	Create(entry *models.WaitingList) error
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
