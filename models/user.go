package models

import "time"

type User struct {
	ID            string    `gorm:"primaryKey;type:uuid" db:"id" json:"id"`
	Email         string    `gorm:"unique;not null" db:"email" json:"email"`
	Username      string    `gorm:"unique;not null" db:"username" json:"username"`
	Password_Hash *string   `gorm:"db:\"password_hash\" json:\"-\""`
	GoogleID      *string   `gorm:"column:google_id;uniqueIndex" db:"google_id" json:"-"`
	GoogleEmail   *string   `gorm:"column:google_email" db:"google_email" json:"google_email,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

func (User) TableName() string {
	return "users"
}

type UserSignUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789"`
}

type UserSignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type GoogleLinkRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

func (u *User) IsGoogleLinked() bool {
	return u.GoogleID != nil && *u.GoogleID != ""
}

func (u *User) LinkGoogleAccount(googleID, googleEmail string) {
	u.GoogleID = &googleID
	u.GoogleEmail = &googleEmail
}

func (u *User) UnlinkGoogleAccount() {
	u.GoogleID = nil
	u.GoogleEmail = nil
}
