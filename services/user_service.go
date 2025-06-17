package services

import (
	"asset-diary/repositories"
)

type UserServiceInterface interface {
	DeleteUser(userID string) error
	GetAllUserIDs() ([]string, error)
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

// GetAllUserIDs returns a list of all user IDs in the system
func (s *UserService) GetAllUserIDs() ([]string, error) {
	return s.userRepo.ListAllUserIDs()
}
