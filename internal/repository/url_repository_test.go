package repository

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/navaneethk-000/url-shortener-backend/internal/database"
	"github.com/navaneethk-000/url-shortener-backend/internal/models"
)

func TestCreateUrl(t *testing.T) {

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

	repo := NewUrlRepository(db)

	url := &models.Url{OriginalURL: "https://test.com", ShortCode: "test1"}
	err := repo.Create(url)

	if err != nil {
		t.Errorf("Failed to create URL: %v", err)
	}
	if url.ID == 0 {
		t.Error("ID should be auto-generated")
	}

	// Cleanup database after test
	db.Delete(&models.Url{}, url.ID)

}

// Helper function to get env var with a default value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
