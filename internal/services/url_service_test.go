package services

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/navaneethk-000/url-shortener-backend/internal/database"
	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"github.com/navaneethk-000/url-shortener-backend/internal/repository"
)

func setupService() *UrlService {

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
	return NewUrlService(repository.NewUrlRepository(db), repository.NewClickRepository(db))
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func TestShorten_ShouldGenerateId_IfAliasMissing(t *testing.T) {
	service := setupService()

	// Clean up old data
	service.UrlRepo.DB.Unscoped().Where("original_url = ?", "https://service-test.com").Delete(&models.Url{})

	// Shorten a Url without alias
	url, err := service.Shorten("https://service-test.com", "")

	if err != nil {
		t.Fatalf("Service failed: %v", err)
	}

	// Verify Base62 was used
	if url.ShortCode == "" {
		t.Error("Expected generated ShortCode, got empty string")
	}

	// Cleanup after test
	service.UrlRepo.DB.Unscoped().Delete(url)
}

func TestShorten_ShouldFail_IfAliasTaken(t *testing.T) {

	service := setupService()
	alias := "custom1"

	// Pre-cleanup
	service.UrlRepo.DB.Unscoped().Where("short_code = ?", alias).Delete(&models.Url{})

	// Create first time
	u1, err := service.Shorten("https://a.com", alias)
	if err != nil {
		t.Fatalf("First create failed: %v", err)
	}

	// Create second time which should fail
	_, err = service.Shorten("https://b.com", alias)

	if err == nil {
		t.Error("Expected error for duplicate alias, got nil")
	}
	if err.Error() != "alias already in use" {
		t.Errorf("Expected 'alias already in use', got '%v'", err)
	}

	// Cleanup after test
	service.UrlRepo.DB.Unscoped().Delete(u1)
}
