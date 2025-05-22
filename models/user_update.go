package models

// UserUpdateRequest matches the OpenAPI schema for updating a user profile
// Add more fields as needed from openapi.json

type UserUpdateRequest struct {
	Username          string             `json:"username"`
	InvestmentProfile *InvestmentProfile `json:"investmentProfile,omitempty"`
}
