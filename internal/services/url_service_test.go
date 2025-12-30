package services

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/navaneethk-000/url-shortener-backend/internal/database"
	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"github.com/navaneethk-000/url-shortener-backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var testDB *gorm.DB

// Setup Helper
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

	testDB = database.InitDB(dsn)

	// Create Repos
	urlRepo := repository.NewUrlRepository(testDB)
	clickRepo := repository.NewClickRepository(testDB)

	InitWorker(clickRepo, 10, 1)

	return NewUrlService(urlRepo, clickRepo)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// User Helper
func createDummyUser(email string) *models.User {
	userRepo := repository.NewUserRepository(testDB)
	user := &models.User{
		Email:    email,
		Password: "hashedpassword",
		Name:     "Test User",
	}
	// Clean up if exists
	testDB.Unscoped().Where("email = ?", user.Email).Delete(&models.User{})
	_ = userRepo.Create(user)
	return user
}

func TestShorten_AutoGeneration(t *testing.T) {
	service := setupService()
	// Create a valid user
	user := createDummyUser("auto_gen@test.com")

	// Cleanup potential old data
	testDB.Unscoped().Where("original_url = ?", "https://auto.com").Delete(&models.Url{})

	url, err := service.Shorten("https://auto.com", "", user.ID)

	// Assertion
	assert.NoError(t, err)
	if url != nil {
		assert.NotEmpty(t, url.ShortCode)
		assert.Equal(t, "https://auto.com", url.OriginalURL)
		assert.Equal(t, user.ID, url.UserID)

		// Cleanup URL
		testDB.Unscoped().Delete(url)
	}

	// Cleanup User
	testDB.Unscoped().Delete(user)
}

func TestShorten_CustomAlias(t *testing.T) {
	service := setupService()
	user := createDummyUser("alias_user@test.com")
	alias := "my-alias"

	// Cleanup
	testDB.Unscoped().Where("short_code = ?", alias).Delete(&models.Url{})

	// Success Case
	url, err := service.Shorten("https://google.com", alias, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, alias, url.ShortCode)

	// Duplicate Failure Case
	_, err = service.Shorten("https://yahoo.com", alias, user.ID)
	assert.Error(t, err)
	assert.Equal(t, "alias already in use", err.Error())

	// Cleanup
	testDB.Unscoped().Delete(url)
	testDB.Unscoped().Delete(user)
}

func TestResolve_And_Analytics(t *testing.T) {
	service := setupService()
	user := createDummyUser("resolve@test.com")

	// Cleanup
	testDB.Unscoped().Where("original_url = ?", "https://target.com").Delete(&models.Url{})

	// Create URL
	url, err := service.Shorten("https://target.com", "", user.ID)
	if err != nil {
		t.Fatalf("Setup Failed: %v", err)
	}

	original, err := service.Resolve(url.ShortCode, "google.com", "Firefox", "127.0.0.1")

	// Assert Redirect
	assert.NoError(t, err)
	assert.Equal(t, "https://target.com", original)

	// Wait for Async Goroutine to finish
	time.Sleep(200 * time.Millisecond)

	// Assert Analytics
	var clicks []models.Click
	testDB.Where("url_id = ?", url.ID).Find(&clicks)

	if len(clicks) == 0 {
		t.Error("Expected 1 click to be recorded, found 0. (Is worker pool running in main?)")
	} else {
		assert.Equal(t, "Firefox", clicks[0].UserAgent)
		// Clean up specific click
		testDB.Unscoped().Delete(&clicks)
	}

	// Cleanup
	testDB.Unscoped().Delete(url)
	testDB.Unscoped().Delete(user)
}

func TestDeleteShortLink(t *testing.T) {
	service := setupService()
	owner := createDummyUser("owner@test.com")
	hacker := createDummyUser("hacker@test.com")

	url, err := service.Shorten("https://delete.com", "", owner.ID)
	if err != nil {
		t.Fatalf("Setup Failed: %v", err)
	}

	// Wrong User (Hacker tries to delete)
	err = service.DeleteShortLink(url.ShortCode, hacker.ID)
	assert.Error(t, err)
	assert.Equal(t, "unauthorized: you do not own this link", err.Error())

	// Success: Owner deletes
	err = service.DeleteShortLink(url.ShortCode, owner.ID)
	assert.NoError(t, err)

	// Verify deletion
	found, _ := service.UrlRepo.FindByShortCode(url.ShortCode)
	assert.Nil(t, found)

	// Cleanup
	testDB.Unscoped().Delete(owner)
	testDB.Unscoped().Delete(hacker)
}

func TestGetUserUrls(t *testing.T) {
	service := setupService()
	user := createDummyUser("history@test.com")

	// Create 2 links
	u1, _ := service.Shorten("https://1.com", "", user.ID)
	u2, _ := service.Shorten("https://2.com", "", user.ID)

	// Action
	urls, err := service.GetUserUrls(user.ID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, len(urls))

	// Cleanup
	testDB.Unscoped().Delete(u1)
	testDB.Unscoped().Delete(u2)
	testDB.Unscoped().Delete(user)
}

func TestGenerateQRCode(t *testing.T) {
	service := setupService()
	user := createDummyUser("qr@test.com")
	url, err := service.Shorten("https://qr.com", "", user.ID)
	if err != nil {
		t.Fatalf("Setup Failed: %v", err)
	}

	// Action
	png, err := service.GenerateQRCode(url.ShortCode)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, png)
	assert.Greater(t, len(png), 100)

	// Cleanup
	testDB.Unscoped().Delete(url)
	testDB.Unscoped().Delete(user)
}
