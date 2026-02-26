package services

import (
	"asset-diary/models"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWaitingListRepository is a mock implementation of WaitingListRepositoryInterface
type MockWaitingListRepository struct {
	mock.Mock
}

func (m *MockWaitingListRepository) Create(entry *models.WaitingList) error {
	args := m.Called(entry)
	return args.Error(0)
}

func TestJoinWaitingList(t *testing.T) {
	mockRepo := new(MockWaitingListRepository)
	service := NewWaitingListService(mockRepo)

	t.Run("successful join", func(t *testing.T) {
		entry := &models.WaitingList{
			Email:   "test@example.com",
			Project: "test-project",
		}

		mockRepo.On("Create", entry).Return(nil).Once()

		err := service.JoinWaitingList(entry)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		entry := &models.WaitingList{
			Email:   "error@example.com",
			Project: "test-project",
		}

		expectedErr := errors.New("db error")
		mockRepo.On("Create", entry).Return(expectedErr).Once()

		err := service.JoinWaitingList(entry)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}
