package repositories

import (
	"log"

	"asset-diary/models"

	"gorm.io/gorm"
)

type AccountRepositoryInterface interface {
	ListAccounts(userID string) ([]models.Account, error)
	CreateAccount(userID string, acc *models.Account) error
	UpdateAccount(userID, accID string, req models.AccountUpdateRequest) (*models.Account, error)
	DeleteAccount(userID, accID string) error
}

type AccountRepository struct {
	DB *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{DB: db}
}

func (r *AccountRepository) ListAccounts(userID string) ([]models.Account, error) {
	var gormAccounts []models.Account
	result := r.DB.Where(&models.Account{UserID: userID}).Find(&gormAccounts)
	if result.Error != nil {
		log.Println("Failed to fetch accounts:", result.Error)
		return nil, result.Error
	}

	accounts := make([]models.Account, len(gormAccounts))
	for i, gormAcc := range gormAccounts {
		accounts[i] = models.Account{
			ID:       gormAcc.ID,
			Name:     gormAcc.Name,
			Currency: gormAcc.Currency,
			Balance:  gormAcc.Balance,
		}
	}
	return accounts, nil
}

func (r *AccountRepository) CreateAccount(userID string, acc *models.Account) error {
	gormAcc := models.Account{
		ID:       acc.ID,
		UserID:   userID,
		Name:     acc.Name,
		Currency: acc.Currency,
		Balance:  acc.Balance,
	}

	result := r.DB.Create(&gormAcc)
	if result.Error != nil {
		log.Println("Failed to create account:", result.Error)
		return result.Error
	}
	return nil
}

func (r *AccountRepository) UpdateAccount(userID, accID string, req models.AccountUpdateRequest) (*models.Account, error) {
	var gormAccount models.Account
	result := r.DB.Where(&models.Account{ID: accID, UserID: userID}).First(&gormAccount)
	if result.Error != nil {
		log.Println("Failed to find account:", result.Error)
		return nil, result.Error
	}

	// Update fields from request
	gormAccount.Name = req.Name
	gormAccount.Currency = req.Currency
	gormAccount.Balance = req.Balance

	result = r.DB.Save(&gormAccount)
	if result.Error != nil {
		log.Println("Failed to update account:", result.Error)
		return nil, result.Error
	}

	return &models.Account{
		ID:       gormAccount.ID,
		Name:     gormAccount.Name,
		Currency: gormAccount.Currency,
		Balance:  gormAccount.Balance,
	}, nil
}

func (r *AccountRepository) DeleteAccount(userID, accID string) error {
	result := r.DB.Where(&models.Account{ID: accID, UserID: userID}).Delete(&models.Account{})
	if result.Error != nil {
		log.Println("Failed to delete account:", result.Error)
		return result.Error
	}
	return nil
}
