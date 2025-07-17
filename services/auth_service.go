package services

import (
	"asset-diary/models"
	"asset-diary/repositories"
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

var (
	ErrInvalidToken = errors.New("Invalid refresh token")
)

type AuthServiceInterface interface {
	SignUp(req *models.UserSignUpRequest) (*models.AuthResponse, error)
	SignIn(email, password string) (*models.AuthResponse, error)
	GoogleLogin(token string) (*models.AuthResponse, error)
	RefreshToken(refreshToken string) (string, string, error)
	ForgotPassword(email string) error
	VerifyResetCode(email, code string) error
	RevokeRefreshToken(token string) error
}

type AuthService struct {
	authRepo    repositories.AuthRepositoryInterface
	userService UserServiceInterface
}

func NewAuthService(authRepo repositories.AuthRepositoryInterface, userService UserServiceInterface) *AuthService {
	return &AuthService{
		authRepo:    authRepo,
		userService: userService,
	}
}

func (s *AuthService) SignUp(req *models.UserSignUpRequest) (*models.AuthResponse, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	hashedPassword := string(hashed)
	user := &models.User{
		Email:         req.Email,
		Username:      req.Username,
		Password_Hash: &hashedPassword,
	}

	err = s.authRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.generateAndStoreRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token:        accessToken,
		User:         *user,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) GoogleLogin(token string) (*models.AuthResponse, error) {
	payload, err := idtoken.Validate(context.Background(), token, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		return nil, errors.New("invalid Google token")
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		return nil, errors.New("failed to extract email from token")
	}
	name, _ := payload.Claims["name"].(string)
	googleID, _ := payload.Claims["sub"].(string)

	user, err := s.authRepo.FindUserByEmail(email)
	if err == gorm.ErrRecordNotFound {
		user = &models.User{
			Email:    email,
			Username: name,
		}
		user.LinkGoogleAccount(googleID, email)
		err = s.authRepo.CreateUser(user)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	if user.IsGoogleLinked() {
		if *user.GoogleID != googleID {
			return nil, errors.New("this Google account is already linked to a different user account")
		}
	} else {
		if user.Password_Hash != nil {
			return nil, errors.New("this email is already registered with a password. Please sign in with your password instead")
		}
	}

	accessToken, refreshToken, err := s.generateAndStoreRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %v", err)
	}

	return &models.AuthResponse{
		Token:        accessToken,
		User:         *user,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) SignIn(email, password string) (*models.AuthResponse, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	user, err := s.authRepo.FindUserByEmail(email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.Password_Hash == nil {
		return nil, errors.New("no password set for this account")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.Password_Hash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	accessToken, refreshToken, err := s.generateAndStoreRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func hashRefreshToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *AuthService) generateAndStoreRefreshToken(userID, email string) (string, string, error) {
	accessToken, refreshToken, err := generateTokens(userID, email)
	if err != nil {
		return "", "", err
	}

	tokenHash := hashRefreshToken(refreshToken)

	refreshExpiry := 168 * time.Hour
	if expiryStr := os.Getenv("REFRESH_TOKEN_EXPIRY"); expiryStr != "" {
		if duration, err := time.ParseDuration(expiryStr); err == nil {
			refreshExpiry = duration
		}
	}

	if err := s.authRepo.StoreRefreshToken(userID, tokenHash, time.Now().Add(refreshExpiry)); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (string, string, error) {
	tokenHash := hashRefreshToken(refreshToken)
	token, err := s.authRepo.FindRefreshToken(tokenHash)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	claims, err := validateRefreshToken(refreshToken)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	userID, ok1 := claims["user_id"].(string)
	email, ok2 := claims["email"].(string)
	if !ok1 || !ok2 || userID != token.UserID {
		return "", "", ErrInvalidToken
	}

	if err := s.authRepo.RevokeRefreshToken(tokenHash); err != nil {
		return "", "", fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return s.generateAndStoreRefreshToken(userID, email)
}

func (s *AuthService) RevokeRefreshToken(token string) error {
	tokenHash := hashRefreshToken(token)
	return s.authRepo.RevokeRefreshToken(tokenHash)
}

func generateAccessToken(userID string, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT secret not set in environment")
	}

	tokenExpiry := 1 * time.Hour
	tokenExpiryStr := os.Getenv("ACCESS_TOKEN_EXPIRY")
	if tokenExpiryStr != "" {
		duration, err := time.ParseDuration(tokenExpiryStr)
		if err == nil {
			tokenExpiry = duration
		}
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(tokenExpiry).Unix(),
		"jti":     uuid.New().String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func generateRefreshToken(userID string, email string) (string, error) {
	secret := os.Getenv("JWT_REFRESH_SECRET")
	if secret == "" {
		return "", errors.New("JWT refresh secret not set in environment")
	}

	refreshExpiry := 168 * time.Hour
	refreshExpiryStr := os.Getenv("REFRESH_TOKEN_EXPIRY")
	if refreshExpiryStr != "" {
		duration, err := time.ParseDuration(refreshExpiryStr)
		if err == nil {
			refreshExpiry = duration
		}
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(refreshExpiry).Unix(),
		"jti":     uuid.New().String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func validateRefreshToken(refreshToken string) (jwt.MapClaims, error) {
	secret := os.Getenv("JWT_REFRESH_SECRET")
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func generateTokens(userID, email string) (string, string, error) {
	accessToken, err := generateAccessToken(userID, email)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := generateRefreshToken(userID, email)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) ForgotPassword(email string) error {
	user, err := s.authRepo.FindUserByEmail(email)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	verificationCode := generateVerificationCode()

	err = s.authRepo.StoreVerificationCode(email, verificationCode, 3*time.Minute)
	if err != nil {
		return errors.New("failed to generate verification code")
	}

	err = sendVerificationEmail(email, verificationCode)
	if err != nil {
		return errors.New("failed to send verification email")
	}

	return nil
}

func (s *AuthService) VerifyResetCode(email, code string) error {
	isValid, err := s.authRepo.ValidateVerificationCode(email, code)
	if err != nil || !isValid {
		return errors.New("invalid or expired verification code")
	}

	newPassword := generateRandomPassword()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to generate new password")
	}

	err = s.authRepo.UpdateUserPassword(email, string(hashedPassword))
	if err != nil {
		return errors.New("failed to update password")
	}
	err = sendPasswordResetEmail(email, newPassword)
	if err != nil {
		return errors.New("failed to send password reset email")
	}

	return nil
}

func sendPasswordResetEmail(email, newPassword string) error {
	from := os.Getenv("EMAIL_FROM")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpHost := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")

	if from == "" || smtpUser == "" || smtpPass == "" || smtpHost == "" || portStr == "" {
		return errors.New("incomplete email configuration")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 {
		return fmt.Errorf("invalid SMTP port: %s", portStr)
	}

	var body bytes.Buffer
	templ := template.Must(template.New("passwordReset").Parse(`Your password has been reset. Please log in with the following temporary password:

{{.Password}}

Please change this password immediately after logging in.

Best regards,
Asset Diary Team`))

	err = templ.Execute(&body, struct {
		Password string
	}{
		Password: newPassword,
	})
	if err != nil {
		return fmt.Errorf("failed to create email template: %v", err)
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Asset Diary Password Reset")
	msg.SetBody("text/plain", body.String())

	dialer := gomail.NewDialer(smtpHost, port, smtpUser, smtpPass)
	dialer.TLSConfig = &tls.Config{ServerName: smtpHost, InsecureSkipVerify: false}

	if err := dialer.DialAndSend(msg); err != nil {
		log.Printf("Failed to send password reset email to %s: %v (SMTP User: %s, Host: %s:%d)", email, err, smtpUser, smtpHost, port)
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("Password reset email sent to %s", email)
	return nil
}

func generateVerificationCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func generateRandomPassword() string {
	passwordLength := 12
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
	password := make([]byte, passwordLength)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}

func sendVerificationEmail(email, code string) error {
	from := os.Getenv("EMAIL_FROM")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpHost := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")

	if from == "" || smtpUser == "" || smtpPass == "" || smtpHost == "" || portStr == "" {
		return errors.New("incomplete email configuration")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 {
		return fmt.Errorf("invalid SMTP port: %s", portStr)
	}

	var body bytes.Buffer
	templ := template.Must(template.New("verificationCode").Parse(`Your verification code is:

{{.Code}}

This code will expire in 3 minutes.

Best regards,
Asset Diary Team`))

	err = templ.Execute(&body, struct {
		From string
		To   string
		Code string
	}{
		From: from,
		To:   email,
		Code: code,
	})
	if err != nil {
		return fmt.Errorf("failed to create email template: %v", err)
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Asset Diary Verification Code")
	msg.SetBody("text/plain", body.String())

	dialer := gomail.NewDialer(smtpHost, port, smtpUser, smtpPass)
	dialer.TLSConfig = &tls.Config{ServerName: smtpHost, InsecureSkipVerify: false}

	if err := dialer.DialAndSend(msg); err != nil {
		log.Printf("Failed to send email to %s: %v (SMTP User: %s, Host: %s:%d)", email, err, smtpUser, smtpHost, port)
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("Verification email sent to %s", email)
	return nil
}
