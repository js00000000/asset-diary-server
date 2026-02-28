package handlers

import (
	"asset-diary/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWaitingListService struct {
	mock.Mock
}

func (m *MockWaitingListService) JoinWaitingList(entry *models.WaitingList) error {
	args := m.Called(entry)
	return args.Error(0)
}

func TestWaitingListHandler_Join(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful join", func(t *testing.T) {
		mockService := new(MockWaitingListService)
		handler := NewWaitingListHandler(mockService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		entry := models.WaitingList{
			Email:   "test@example.com",
			Project: "read-together",
		}
		body, _ := json.Marshal(entry)
		c.Request, _ = http.NewRequest(http.MethodPost, "/waiting-list", bytes.NewBuffer(body))

		mockService.On("JoinWaitingList", mock.AnythingOfType("*models.WaitingList")).Return(nil).Once()

		handler.Join(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Successfully joined the waiting list", response["message"])
		mockService.AssertExpectations(t)
	})

	t.Run("project not allowed", func(t *testing.T) {
		mockService := new(MockWaitingListService)
		handler := NewWaitingListHandler(mockService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		entry := models.WaitingList{
			Email:   "test@example.com",
			Project: "unauthorized-project",
		}
		body, _ := json.Marshal(entry)
		c.Request, _ = http.NewRequest(http.MethodPost, "/waiting-list", bytes.NewBuffer(body))

		appErr := models.NewAppError(models.ErrCodeProjectNotAllowed, "This project is not currently accepting new signups for the waiting list.")
		mockService.On("JoinWaitingList", mock.AnythingOfType("*models.WaitingList")).Return(appErr).Once()

		handler.Join(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response models.AppError
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, models.ErrCodeProjectNotAllowed, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("duplicate email", func(t *testing.T) {
		mockService := new(MockWaitingListService)
		handler := NewWaitingListHandler(mockService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		entry := models.WaitingList{
			Email:   "test@example.com",
			Project: "read-together",
		}
		body, _ := json.Marshal(entry)
		c.Request, _ = http.NewRequest(http.MethodPost, "/waiting-list", bytes.NewBuffer(body))

		// Simulating a DB error for duplicates
		mockService.On("JoinWaitingList", mock.AnythingOfType("*models.WaitingList")).Return(&DuplicateError{}).Once()

		handler.Join(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		var response models.AppError
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, models.ErrCodeDuplicateEmail, response.Code)
		mockService.AssertExpectations(t)
	})
}

// DuplicateError is a helper to simulate GORM duplicate key errors
type DuplicateError struct{}

func (e *DuplicateError) Error() string {
	return `duplicate key value violates unique constraint "waiting_lists_email_project_key"`
}
