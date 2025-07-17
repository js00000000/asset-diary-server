package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"asset-diary/models"
	"asset-diary/repositories"
)

type GoogleUserInfo struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type UserServiceInterface interface {
	DeleteUser(userID string) error
	GetAllUserIDs() ([]string, error)
	FindUserByEmail(email string) (*models.User, error)
	LinkGoogleAccount(userID, idToken string) error
	UnlinkGoogleAccount(userID string) error
	GetGoogleAccountStatus(userID string) (bool, error)
}

type UserService struct {
	userRepo repositories.UserRepositoryInterface
}

func NewUserService(userRepo repositories.UserRepositoryInterface) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) DeleteUser(userID string) error {
	return s.userRepo.DeleteUser(userID)
}

func (s *UserService) GetAllUserIDs() ([]string, error) {
	return s.userRepo.ListAllUserIDs()
}

func (s *UserService) FindUserByEmail(email string) (*models.User, error) {
	return s.userRepo.FindUserByEmail(email)
}

func (s *UserService) LinkGoogleAccount(userID, idToken string) error {
	userInfo, err := s.verifyGoogleIDToken(idToken)
	if err != nil {
		return fmt.Errorf("failed to verify Google ID token: %w", err)
	}

	existingUser, err := s.userRepo.FindUserByGoogleID(userInfo.ID)
	if err != nil {
		return fmt.Errorf("failed to check for existing Google account: %w", err)
	}
	if existingUser != nil && existingUser.ID != userID {
		return fmt.Errorf("this Google account is already linked to another user")
	}

	if existingUser != nil && existingUser.ID == userID {
		return nil
	}

	return s.userRepo.LinkGoogleAccount(userID, userInfo.ID, userInfo.Email)
}

func (s *UserService) UnlinkGoogleAccount(userID string) error {
	return s.userRepo.UnlinkGoogleAccount(userID)
}

func (s *UserService) GetGoogleAccountStatus(userID string) (bool, error) {
	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		return false, fmt.Errorf("failed to get Google account status: %w", err)
	}

	return user.IsGoogleLinked(), nil
}

func (s *UserService) verifyGoogleIDToken(idToken string) (*GoogleUserInfo, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken))
	if err != nil {
		return nil, fmt.Errorf("failed to verify token with Google: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Google response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google API error: %s", string(body))
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse Google user info: %w", err)
	}

	if userInfo.ID == "" || userInfo.Email == "" {
		return nil, fmt.Errorf("invalid user info from Google")
	}

	// if !userInfo.VerifiedEmail {
	// 	return nil, fmt.Errorf("Google email not verified")
	// }

	return &userInfo, nil
}
