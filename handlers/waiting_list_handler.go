package handlers

import (
	"asset-diary/models"
	"asset-diary/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WaitingListHandler struct {
	service services.WaitingListServiceInterface
}

func NewWaitingListHandler(service services.WaitingListServiceInterface) *WaitingListHandler {
	return &WaitingListHandler{service: service}
}

func (h *WaitingListHandler) Join(c *gin.Context) {
	var entry models.WaitingList
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, models.NewAppError(
			models.ErrCodeInvalidRequest,
			err.Error(),
		))
		return
	}

	if err := h.service.JoinWaitingList(&entry); err != nil {
		if models.IsDuplicateError(err, "waiting_lists_email_project_key") {
			c.JSON(http.StatusConflict, models.NewAppError(
				models.ErrCodeDuplicateEmail,
				"You are already on the waiting list for this project",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewAppError(
			models.ErrCodeInternal,
			"Failed to join waiting list",
		))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Successfully joined the waiting list",
		"id":      entry.ID,
	})
}
