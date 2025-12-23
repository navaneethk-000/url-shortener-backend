package repository

import (
	"fmt"
	"testing"

	"github.com/joho/godotenv"
	"github.com/navaneethk-000/url-shortener-backend/internal/database"
	"github.com/navaneethk-000/url-shortener-backend/internal/models"
)

func TestSaveAndGetClickStats(t *testing.T) {
	// DB Connection
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

	urlRepo := NewUrlRepository(db)
	clickRepo := NewClickRepository(db)

	// We can't save a click for a URL that doesn't exist (Foreign Key Constraint) so creating a parent url
	parentUrl := &models.Url{OriginalURL: "https://analytics-test.com", ShortCode: "stats1"}

	// Clean up any old data to avoid unique constraint error
	db.Unscoped().Where("short_code = ?", "stats1").Delete(&models.Url{})

	err := urlRepo.Create(parentUrl)
	if err != nil {
		t.Fatalf("Failed to create parent URL: %v", err)
	}

	// Test: Save a Click
	click := &models.Click{
		UrlID:     parentUrl.ID,
		Referrer:  "google.com",
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}

	err = clickRepo.SaveClick(click)
	if err != nil {
		t.Errorf("Failed to save click: %v", err)
	}

	// Test: Retrieve Stats
	stats, err := clickRepo.GetStatsByUrlID(parentUrl.ID)
	if err != nil {
		t.Errorf("Failed to get stats: %v", err)
	}

	if len(stats) != 1 {
		t.Errorf("Expected 1 click log, got %d", len(stats))
	}
	if stats[0].Referrer != "google.com" {
		t.Errorf("Expected referrer google.com, got %s", stats[0].Referrer)
	}

	// Cleanup database after test
	db.Unscoped().Delete(click)     // Delete click first
	db.Unscoped().Delete(parentUrl) // Delete parent URL second
}
