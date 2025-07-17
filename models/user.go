package models

import "time"

type User struct {
	ID            string    `gorm:"primaryKey;type:uuid" db:"id" json:"id"`
	Email         string    `gorm:"unique;not null" db:"email" json:"email"`
	Username      string    `gorm:"unique;not null" db:"username" json:"username"`
	Password_Hash *string   `gorm:"db:\"password_hash\" json:\"-\""`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

func (User) TableName() string {
	return "users"
}

type UserSignUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserSignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
