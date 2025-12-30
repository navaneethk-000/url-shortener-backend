package services

import (
	"testing"
	"time"

	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"github.com/navaneethk-000/url-shortener-backend/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestWorkerPool(t *testing.T) {
	// setupService to ensure the global 'testDB' is initialized correctly
	setupService()
	db := testDB

	clickRepo := repository.NewClickRepository(db)
	userRepo := repository.NewUserRepository(db)
	urlRepo := repository.NewUrlRepository(db)

	// Create User
	email := "worker_test@example.com"
	// Cleanup User
	db.Unscoped().Where("email = ?", email).Delete(&models.User{})

	user := &models.User{Email: email, Password: "pw", Name: "Worker Tester"}
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Setup Failed: Could not create user: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("Setup Failed: User ID is 0")
	}

	// Create URL
	shortCode := "work1"
	// Find the old URL ID to delete its clicks first (Foreign Key Constraint)
	var oldUrl models.Url
	if err := db.Unscoped().Where("short_code = ?", shortCode).First(&oldUrl).Error; err == nil {
		// Delete associated clicks first
		db.Unscoped().Where("url_id = ?", oldUrl.ID).Delete(&models.Click{})
		// Then delete the URL
		db.Unscoped().Delete(&oldUrl)
	}

	url := &models.Url{
		OriginalURL: "http://worker-test.com",
		ShortCode:   shortCode,
		UserID:      user.ID,
	}
	err = urlRepo.Create(url)
	if err != nil {
		t.Fatalf("Setup Failed: Could not create url: %v", err)
	}
	if url.ID == 0 {
		t.Fatal("Setup Failed: URL ID is 0")
	}

	// Initialize Worker (Buffer 10, 1 Worker)
	InitWorker(clickRepo, 10, 1)

	// Push a Job
	click := &models.Click{
		UrlID:     url.ID,
		IPAddress: "127.0.0.1",
	}
	WorkerInstance.Push(click)

	// Wait for async processing
	time.Sleep(500 * time.Millisecond)

	// Assert
	stats, err := clickRepo.GetStatsByUrlID(url.ID)
	if err != nil {
		t.Fatalf("Failed to fetch stats: %v", err)
	}

	// Prevent Panic: Check length first
	if len(stats) == 0 {
		t.Fatal("Expected 1 click, got 0. Worker failed to save to DB.")
	}

	assert.Equal(t, "Localhost", stats[0].Country)
	assert.Equal(t, "Localhost", stats[0].City)

	// --- Cleanup ---
	db.Unscoped().Delete(&models.Click{}, stats[0].ID)
	db.Unscoped().Delete(url)
	db.Unscoped().Delete(user)
}
