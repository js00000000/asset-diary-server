package handlers

import (
	"net/http"

	"asset-diary/models"
	"asset-diary/services"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	AccountService services.AccountServiceInterface
}

func NewAccountHandler(accountService services.AccountServiceInterface) *AccountHandler {
	return &AccountHandler{AccountService: accountService}
}

// ListAccounts returns all accounts for the current user
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	accounts, err := h.AccountService.ListAccounts(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accounts"})
		return
	}
	responses := []models.AccountResponse{}
	for _, acc := range accounts {
		responses = append(responses, models.AccountResponse{
			ID:       acc.ID,
			Name:     acc.Name,
			Currency: acc.Currency,
			Balance:  acc.Balance,
		})
	}
	c.JSON(http.StatusOK, responses)
}

// CreateAccount creates a new account for the current user
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var req models.AccountCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	acc, err := h.AccountService.CreateAccount(userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}
	response := models.AccountResponse{
		ID:       acc.ID,
		Name:     acc.Name,
		Currency: acc.Currency,
		Balance:  acc.Balance,
	}
	c.JSON(http.StatusCreated, response)
}

// UpdateAccount updates an account by id for the current user
func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	accID := c.Param("id")
	var req models.AccountUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	acc, err := h.AccountService.UpdateAccount(userID.(string), accID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
		return
	}
	response := models.AccountResponse{
		ID:       acc.ID,
		Name:     acc.Name,
		Currency: acc.Currency,
		Balance:  acc.Balance,
	}
	c.JSON(http.StatusOK, response)
}

// DeleteAccount deletes an account by id for the current user
func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	accID := c.Param("id")
	err := h.AccountService.DeleteAccount(userID.(string), accID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}
	c.Status(http.StatusNoContent)
}
