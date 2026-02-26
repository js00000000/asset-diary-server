package models

import (
	"encoding/json"
	"time"
)

type WaitingList struct {
	ID          string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Email       string          `gorm:"not null" json:"email" binding:"required,email"`
	Project     string          `gorm:"not null" json:"project" binding:"required"`
	Information json.RawMessage `gorm:"type:jsonb" json:"information"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func (WaitingList) TableName() string {
	return "waiting_lists"
}
