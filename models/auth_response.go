package models

// AuthResponse represents the response for authentication operations
type AuthResponse struct {
	Token        string `json:"token"`
	User         User   `json:"user"`
	RefreshToken string `json:"refreshToken,omitempty"`
}
