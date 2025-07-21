package models

import (
	"strings"
)

type ErrorCode string

const (
	ErrCodeDuplicateEmail    ErrorCode = "DUPLICATE_EMAIL"
	ErrCodeDuplicateUsername ErrorCode = "DUPLICATE_USERNAME"
	ErrCodeInternal          ErrorCode = "INTERNAL_ERROR"
	ErrCodeInvalidRequest    ErrorCode = "INVALID_REQUEST"
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
)

type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// IsDuplicateError checks if the error is a duplicate key error for the given field
func IsDuplicateError(err error, field string) bool {
	if err == nil {
		return false
	}
	// This is a simple check, you might need to adjust based on your database driver
	return strings.Contains(err.Error(), "duplicate key value") &&
		strings.Contains(err.Error(), field)
}
