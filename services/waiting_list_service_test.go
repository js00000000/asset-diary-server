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

func (m *MockWaitingListRepository) IsProjectAllowed(name string) (bool, error) {
	args := m.Called(name)
	return args.Bool(0), args.Error(1)
}

func TestJoinWaitingList(t *testing.T) {
	mockRepo := new(MockWaitingListRepository)
	service := NewWaitingListService(mockRepo)

	t.Run("successful join", func(t *testing.T) {
		entry := &models.WaitingList{
			Email:   "test@example.com",
			Project: "read-together",
		}

		mockRepo.On("IsProjectAllowed", "read-together").Return(true, nil).Once()
		mockRepo.On("Create", entry).Return(nil).Once()

		err := service.JoinWaitingList(entry)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("project not allowed", func(t *testing.T) {
		entry := &models.WaitingList{
			Email:   "test@example.com",
			Project: "not-allowed",
		}

		mockRepo.On("IsProjectAllowed", "not-allowed").Return(false, nil).Once()

		err := service.JoinWaitingList(entry)

		assert.Error(t, err)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok)
		assert.Equal(t, models.ErrCodeProjectNotAllowed, appErr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error on IsProjectAllowed", func(t *testing.T) {
		entry := &models.WaitingList{
			Email:   "test@example.com",
			Project: "test-project",
		}

		expectedErr := errors.New("db error")
		mockRepo.On("IsProjectAllowed", "test-project").Return(false, expectedErr).Once()

		err := service.JoinWaitingList(entry)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error on Create", func(t *testing.T) {
		entry := &models.WaitingList{
			Email:   "error@example.com",
			Project: "test-project",
		}

		expectedErr := errors.New("db error")
		mockRepo.On("IsProjectAllowed", "test-project").Return(true, nil).Once()
		mockRepo.On("Create", entry).Return(expectedErr).Once()

		err := service.JoinWaitingList(entry)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}
