package services

import (
	"asset-diary/models"
	"asset-diary/repositories"
)

type UserServiceInterface interface {
	DeleteUser(userID string) error
	GetAllUserIDs() ([]string, error)
	FindUserByEmail(email string) (*models.User, error)
}

type UserService struct {
	userRepo repositories.UserRepositoryInterface
}

func NewUserService(userRepo repositories.UserRepositoryInterface) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) DeleteUser(userID string) error {
	return s.userRepo.DeleteUser(userID)
}

func (s *UserService) GetAllUserIDs() ([]string, error) {
	return s.userRepo.ListAllUserIDs()
}

func (s *UserService) FindUserByEmail(email string) (*models.User, error) {
	return s.userRepo.FindUserByEmail(email)
}
