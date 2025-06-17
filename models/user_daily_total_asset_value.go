package models

import "time"

// UserDailyTotalAssetValue represents a daily snapshot of a user's total asset value
type UserDailyTotalAssetValue struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID     string    `gorm:"type:uuid;not null;index" json:"userId"`
	User       User      `gorm:"foreignKey:UserID;references:ID;onUpdate:CASCADE;onDelete:CASCADE" json:"-"`
	Date       time.Time `gorm:"type:date;not null;index" json:"date"`
	TotalValue float64   `gorm:"type:decimal(20,8);not null" json:"totalValue"`
	Currency   string    `gorm:"type:varchar(3);not null" json:"currency"`
	CreatedAt  time.Time `gorm:"not null;default:current_timestamp" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"not null;default:current_timestamp" json:"updatedAt"`
}

// TableName specifies the table name for the UserDailyTotalAssetValue model
func (UserDailyTotalAssetValue) TableName() string {
	return "user_daily_total_asset_values"
}
