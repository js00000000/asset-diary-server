package handlers

import (
	"net/http"

	"asset-diary/models"
	"asset-diary/services"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	AccountService      services.AccountServiceInterface
	ExchangeRateService services.ExchangeRateServiceInterface
	ProfileService      services.ProfileServiceInterface
}

func NewAccountHandler(accountService services.AccountServiceInterface, exchangeRateService services.ExchangeRateServiceInterface, profileService services.ProfileServiceInterface) *AccountHandler {
	return &AccountHandler{AccountService: accountService, ExchangeRateService: exchangeRateService, ProfileService: profileService}
}

func (h *AccountHandler) ListAccounts(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewAppError(models.ErrCodeUnauthorized, "Unauthorized"))
		return
	}
	accounts, err := h.AccountService.ListAccounts(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAppError(models.ErrCodeInternal, "Failed to fetch accounts"))
		return
	}
	defaultCurrency, err := h.ProfileService.GetDefaultCurrency(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAppError(models.ErrCodeInternal, "Failed to fetch profile"))
		return
	}
	exchangeRates, err := h.ExchangeRateService.GetRatesByBaseCurrency(defaultCurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAppError(models.ErrCodeInternal, "Failed to fetch exchange rates"))
		return
	}
	responses := []models.AccountResponse{}
	for _, acc := range accounts {
		balanceInDefaultCurrency := acc.Balance
		if acc.Balance != 0 {
			balanceInDefaultCurrency = acc.Balance / exchangeRates[acc.Currency]
		}
		responses = append(responses, models.AccountResponse{
			ID:                       acc.ID,
			Name:                     acc.Name,
			Currency:                 acc.Currency,
			Balance:                  acc.Balance,
			BalanceInDefaultCurrency: balanceInDefaultCurrency,
		})
	}
	c.JSON(http.StatusOK, responses)
}

func (h *AccountHandler) CreateAccount(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewAppError(models.ErrCodeUnauthorized, "Unauthorized"))
		return
	}
	var req models.AccountCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewAppError(models.ErrCodeInvalidRequest, err.Error()))
		return
	}
	acc, err := h.AccountService.CreateAccount(userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAppError(models.ErrCodeInternal, "Failed to create account"))
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

func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewAppError(models.ErrCodeUnauthorized, "Unauthorized"))
		return
	}
	accID := c.Param("id")
	var req models.AccountUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewAppError(models.ErrCodeInvalidRequest, err.Error()))
		return
	}
	acc, err := h.AccountService.UpdateAccount(userID.(string), accID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAppError(models.ErrCodeInternal, "Failed to update account"))
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

func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewAppError(models.ErrCodeUnauthorized, "Unauthorized"))
		return
	}
	accID := c.Param("id")
	err := h.AccountService.DeleteAccount(userID.(string), accID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAppError(models.ErrCodeInternal, "Failed to delete account"))
		return
	}
	c.Status(http.StatusNoContent)
}
