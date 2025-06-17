package models

type Account struct {
	ID       string  `gorm:"primaryKey;type:uuid" json:"id"`
	UserID   string  `gorm:"type:uuid;not null;index" json:"user_id"`
	User     User    `gorm:"foreignKey:UserID;references:ID;onUpdate:CASCADE;onDelete:CASCADE" json:"user"`
	Name     string  `gorm:"not null" json:"name"`
	Currency string  `gorm:"not null" json:"currency"`
	Balance  float64 `gorm:"not null" json:"balance"`
}

func (Account) TableName() string {
	return "accounts"
}

type AccountCreateRequest struct {
	Name     string  `json:"name" binding:"required"`
	Currency string  `json:"currency" binding:"required"`
	Balance  float64 `json:"balance" binding:"required"`
}

type AccountUpdateRequest struct {
	Name     string  `json:"name"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

type AccountResponse struct {
	ID                       string  `json:"id"`
	Name                     string  `json:"name"`
	Currency                 string  `json:"currency"`
	Balance                  float64 `json:"balance"`
	BalanceInDefaultCurrency float64 `json:"balanceInDefaultCurrency"`
}
