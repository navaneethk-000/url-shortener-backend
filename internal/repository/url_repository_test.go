package repository

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/navaneethk-000/url-shortener-backend/internal/database"
	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupRepoTest() (*UrlRepository, *gorm.DB) {
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
	return NewUrlRepository(db), db
}

// Creates a dummy user because URLs need a Foreign Key
func createTestUser(db *gorm.DB, email string) *models.User {
	user := &models.User{Email: email, Password: "pw"}
	// Force cleanup first
	db.Unscoped().Where("email = ?", email).Delete(&models.User{})
	db.Create(user)
	return user
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func TestUrlRepository_Create(t *testing.T) {
	repo, db := setupRepoTest()
	owner := createTestUser(db, "create_test@example.com")

	// Cleanup URL if exists
	db.Unscoped().Where("short_code = ?", "test1").Delete(&models.Url{})

	url := &models.Url{
		OriginalURL: "https://test.com",
		ShortCode:   "test1",
		UserID:      owner.ID,
	}

	err := repo.Create(url)

	assert.NoError(t, err)
	assert.NotZero(t, url.ID, "ID should be auto-generated")

	// Verify in DB
	var found models.Url
	db.First(&found, url.ID)
	assert.Equal(t, "https://test.com", found.OriginalURL)

	// Cleanup
	db.Unscoped().Delete(url)
	db.Unscoped().Delete(owner)
}

func TestUrlRepository_FindByShortCode(t *testing.T) {
	repo, db := setupRepoTest()
	owner := createTestUser(db, "find_test@example.com")

	// Prepare Data
	db.Unscoped().Where("short_code = ?", "findme").Delete(&models.Url{})
	url := &models.Url{OriginalURL: "https://find.com", ShortCode: "findme", UserID: owner.ID}
	repo.Create(url)

	// Test Success
	found, err := repo.FindByShortCode("findme")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, url.ID, found.ID)

	// Test Not Found
	notFound, err := repo.FindByShortCode("ghost_link")
	assert.NoError(t, err) // Should not error, just return nil
	assert.Nil(t, notFound)

	// Cleanup
	db.Unscoped().Delete(url)
	db.Unscoped().Delete(owner)
}

func TestUrlRepository_IncrementClicks(t *testing.T) {
	repo, db := setupRepoTest()
	owner := createTestUser(db, "inc_test@example.com")

	// Create URL with 0 clicks
	db.Unscoped().Where("short_code = ?", "clickme").Delete(&models.Url{})
	url := &models.Url{OriginalURL: "https://click.com", ShortCode: "clickme", UserID: owner.ID, TotalClicks: 0}
	repo.Create(url)

	// Increment
	err := repo.IncrementClicks(url.ID)
	assert.NoError(t, err)

	// Verify
	var updated models.Url
	db.First(&updated, url.ID)
	assert.Equal(t, 1, updated.TotalClicks)

	// Cleanup
	db.Unscoped().Delete(url)
	db.Unscoped().Delete(owner)
}

func TestUrlRepository_Delete(t *testing.T) {
	repo, db := setupRepoTest()
	owner := createTestUser(db, "del_test@example.com")

	// Create URL
	db.Unscoped().Where("short_code = ?", "delme").Delete(&models.Url{})
	url := &models.Url{OriginalURL: "https://del.com", ShortCode: "delme", UserID: owner.ID}
	repo.Create(url)

	// Delete it
	err := repo.Delete(url.ID)
	assert.NoError(t, err)

	// Verify it's gone
	var count int64
	db.Model(&models.Url{}).Where("id = ?", url.ID).Count(&count)
	assert.Equal(t, int64(0), count)

	// Cleanup user
	db.Unscoped().Delete(owner)
}
