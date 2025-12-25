package services

import (
	"testing"

	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"github.com/navaneethk-000/url-shortener-backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_HashesPassword(t *testing.T) {

	_ = setupService()

	userRepo := repository.NewUserRepository(testDB)
	authService := NewAuthService(userRepo, "my-secret-key")

	email := "auth_test@example.com"
	rawPassword := "secret123"

	// Cleanup old data
	testDB.Unscoped().Where("email = ?", email).Delete(&models.User{})

	// Register user
	user, err := authService.Register(email, rawPassword)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Ensure the password is hashed and not stored as plain text
	if user.Password == rawPassword {
		t.Error("Security Flaw: Password was saved in plain text!")
	}

	// Password should be a valid Bcrypt hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(rawPassword))
	if err != nil {
		t.Error("Password hash is invalid")
	}
}

func TestLogin_ReturnsToken(t *testing.T) {
	_ = setupService()
	userRepo := repository.NewUserRepository(testDB)
	authService := NewAuthService(userRepo, "my-secret-key")

	email := "login_test@example.com"
	rawPassword := "login123"

	// Register user
	testDB.Unscoped().Where("email = ?", email).Delete(&models.User{})
	_, _ = authService.Register(email, rawPassword)

	// Login with correct password
	token, err := authService.Login(email, rawPassword)
	if err != nil {
		t.Errorf("Login failed with correct password: %v", err)
	}
	if token == "" {
		t.Error("Expected JWT token, got empty string")
	}

	// Login with wrong password
	_, err = authService.Login(email, "wrongpass")
	if err == nil {
		t.Error("Expected error for wrong password, got nil")
	}
}
