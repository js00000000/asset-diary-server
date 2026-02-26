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
	return s.repo.Create(entry)
}
