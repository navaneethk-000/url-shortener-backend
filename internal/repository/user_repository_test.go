package repository

import (
	"fmt"
	"testing"

	"github.com/joho/godotenv"
	"github.com/navaneethk-000/url-shortener-backend/internal/database"
	"github.com/navaneethk-000/url-shortener-backend/internal/models"
)

func TestCreateAndFindUser(t *testing.T) {

	// Setup DB
	_ = godotenv.Load("../../.env")
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "myuser"),
		getEnv("DB_PASSWORD", "mypassword"),
		getEnv("DB_NAME", "shortener_db"),
		getEnv("DB_PORT", "5432"),
	)
	db := database.InitDB(dsn)
	repo := NewUserRepository(db)

	email := "test@example.com"

	// Clean up old test data
	db.Unscoped().Where("email = ?", email).Delete(&models.User{})

	// Test Create
	user := &models.User{
		Email:    email,
		Password: "hashed_secret_password",
	}
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test FindByEmail
	foundUser, err := repo.FindByEmail(email)
	if err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}

	if foundUser.Email != email {
		t.Errorf("Expected email %s, got %s", email, foundUser.Email)
	}

	// Cleanup after test
	db.Unscoped().Delete(user)
}
