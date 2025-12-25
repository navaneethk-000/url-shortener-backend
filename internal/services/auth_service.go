package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"github.com/navaneethk-000/url-shortener-backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo  *repository.UserRepository
	JwtSecret string
}

func NewAuthService(repo *repository.UserRepository, secret string) *AuthService {
	return &AuthService{
		UserRepo:  repo,
		JwtSecret: secret,
	}
}

// Register creates a user with a hashed password
func (s *AuthService) Register(email, password string) (*models.User, error) {
	// Check if user exists
	existing, _ := s.UserRepo.FindByEmail(email)
	if existing != nil {
		return nil, errors.New("email already in use")
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Save user
	user := &models.User{
		Email:    email,
		Password: string(hashed),
	}

	err = s.UserRepo.Create(user)
	return user, err
}

// Login checks credentials and returns JWT
func (s *AuthService) Login(email, password string) (string, error) {
	// Find User
	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}

	// Compare Password (Input vs Hash)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.JwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
