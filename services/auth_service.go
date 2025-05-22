package services

import (
	"asset-diary/models"
	"asset-diary/repositories"
	"bytes"
	"crypto/tls"
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
	"gopkg.in/gomail.v2"
)

var (
	ErrInvalidToken = errors.New("Invalid refresh token")
)

type AuthServiceInterface interface {
	SignUp(req *models.UserSignUpRequest) (*models.AuthResponse, error)
	SignIn(email, password string) (*models.AuthResponse, error)
	RefreshToken(refreshToken string) (string, string, error)
	ForgotPassword(email string) error
	VerifyResetCode(email, code string) error
}

type AuthService struct {
	authRepo repositories.AuthRepositoryInterface
}

func NewAuthService(authRepo repositories.AuthRepositoryInterface) *AuthService {
	return &AuthService{authRepo: authRepo}
}

func (s *AuthService) SignUp(req *models.UserSignUpRequest) (*models.AuthResponse, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &models.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Username:  req.Username,
		CreatedAt: time.Now(),
	}

	err = s.authRepo.CreateUser(user, string(hashed))
	if err != nil {
		return nil, errors.New("email may already be registered")
	}

	token, refreshToken, err := generateTokens(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token:        token,
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

	err = bcrypt.CompareHashAndPassword([]byte(user.Password_Hash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, refreshToken, err := generateTokens(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (string, string, error) {
	claims, err := validateRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	userID, ok1 := claims["user_id"].(string)
	email, ok2 := claims["email"].(string)
	if !ok1 || !ok2 {
		return "", "", ErrInvalidToken
	}

	return generateTokens(userID, email)
}

func generateAccessToken(userID string, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT secret not set in environment")
	}
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func generateRefreshToken(userID string, email string) (string, error) {
	secret := os.Getenv("JWT_REFRESH_SECRET")
	if secret == "" {
		return "", errors.New("JWT refresh secret not set in environment")
	}
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
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
	// Check if user exists
	user, err := s.authRepo.FindUserByEmail(email)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// Generate 6-digit verification code
	verificationCode := generateVerificationCode()

	// Store verification code with expiry
	err = s.authRepo.StoreVerificationCode(email, verificationCode, 3*time.Minute)
	if err != nil {
		return errors.New("failed to generate verification code")
	}

	// Send verification code via email
	err = sendVerificationEmail(email, verificationCode)
	if err != nil {
		return errors.New("failed to send verification email")
	}

	return nil
}

func (s *AuthService) VerifyResetCode(email, code string) error {
	// Verify the code
	isValid, err := s.authRepo.ValidateVerificationCode(email, code)
	if err != nil || !isValid {
		return errors.New("invalid or expired verification code")
	}

	// Generate a new random password
	newPassword := generateRandomPassword()

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to generate new password")
	}

	// Update user's password
	err = s.authRepo.UpdateUserPassword(email, string(hashedPassword))
	if err != nil {
		return errors.New("failed to update password")
	}

	// Send password via email
	err = sendPasswordResetEmail(email, newPassword)
	if err != nil {
		return errors.New("failed to send password reset email")
	}

	return nil
}

func sendPasswordResetEmail(email, newPassword string) error {
	// Email configuration
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
