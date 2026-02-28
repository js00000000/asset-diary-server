package services

import (
	"asset-diary/models"
	"asset-diary/repositories"
)

type WaitingListServiceInterface interface {
	JoinWaitingList(entry *models.WaitingList) error
}

type WaitingListService struct {
	repo repositories.WaitingListRepositoryInterface
}

func NewWaitingListService(repo repositories.WaitingListRepositoryInterface) *WaitingListService {
	return &WaitingListService{repo: repo}
}

func (s *WaitingListService) JoinWaitingList(entry *models.WaitingList) error {
	allowed, err := s.repo.IsProjectAllowed(entry.Project)
	if err != nil {
		return err
	}
	if !allowed {
		return models.NewAppError(models.ErrCodeProjectNotAllowed, "This project is not currently accepting new signups for the waiting list.")
	}
	return s.repo.Create(entry)
}
