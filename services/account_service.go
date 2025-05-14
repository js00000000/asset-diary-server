package services

import (
	"asset-diary/models"
	"asset-diary/repositories"

	"github.com/google/uuid"
)

type AccountServiceInterface interface {
	ListAccounts(userID string) ([]models.Account, error)
	CreateAccount(userID string, req models.AccountCreateRequest) (*models.Account, error)
	UpdateAccount(userID, accID string, req models.AccountUpdateRequest) (*models.Account, error)
	DeleteAccount(userID, accID string) error
}

type AccountService struct {
	repo repositories.AccountRepositoryInterface
}

func NewAccountService(repo repositories.AccountRepositoryInterface) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) ListAccounts(userID string) ([]models.Account, error) {
	return s.repo.ListAccounts(userID)
}

func (s *AccountService) CreateAccount(userID string, req models.AccountCreateRequest) (*models.Account, error) {
	id := uuid.New().String()
	acc := &models.Account{
		ID:       id,
		Name:     req.Name,
		Currency: req.Currency,
		Balance:  req.Balance,
	}

	err := s.repo.CreateAccount(userID, acc)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

func (s *AccountService) UpdateAccount(userID, accID string, req models.AccountUpdateRequest) (*models.Account, error) {
	return s.repo.UpdateAccount(userID, accID, req)
}

func (s *AccountService) DeleteAccount(userID, accID string) error {
	return s.repo.DeleteAccount(userID, accID)
}
