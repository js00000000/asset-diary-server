package models

import "time"

type AllowedProject struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"not null;unique" json:"name"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AllowedProject) TableName() string {
	return "allowed_projects"
}
