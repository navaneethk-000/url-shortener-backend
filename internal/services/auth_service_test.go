package services

import (
	"fmt"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/navaneethk-000/url-shortener-backend/internal/database"
	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"github.com/navaneethk-000/url-shortener-backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Helper to setup Auth Service with DB connection
func setupAuthService() (*AuthService, *gorm.DB) {
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
	userRepo := repository.NewUserRepository(db)

	return NewAuthService(userRepo, "test-secret-key"), db
}

func TestRegister(t *testing.T) {
	service, db := setupAuthService()
	name := "Test User"
	email := "register_test@example.com"
	password := "password123"

	// Cleanup before test
	db.Unscoped().Where("email = ?", email).Delete(&models.User{})

	t.Run("Success - Registers New User", func(t *testing.T) {
		user, err := service.Register(name, email, password)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, name, user.Name)
		assert.NotEqual(t, password, user.Password, "Password should be hashed")

		// Verify Hash
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		assert.NoError(t, err, "Hash should match original password")
	})

	t.Run("Fail - Duplicate Email", func(t *testing.T) {
		// Try to register same email again
		user, err := service.Register(name, email, "newpassword")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "email already in use", err.Error())
	})

	// Cleanup
	db.Unscoped().Where("email = ?", email).Delete(&models.User{})
}

func TestLogin(t *testing.T) {
	service, db := setupAuthService()
	email := "login_test@example.com"
	password := "secret123"

	// Create the user first
	db.Unscoped().Where("email = ?", email).Delete(&models.User{})

	_, err := service.Register("Login User", email, password)
	assert.NoError(t, err)

	t.Run("Success - Login with Correct Credentials", func(t *testing.T) {
		token, err := service.Login(email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify Token claims
		parsedToken, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key"), nil
		})

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, email, claims["email"])
	})

	t.Run("Fail - Wrong Password", func(t *testing.T) {
		token, err := service.Login(email, "wrongpass")

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("Fail - User Not Found", func(t *testing.T) {
		token, err := service.Login("ghost@example.com", "anypass")

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Error(t, err)
	})

	// Cleanup
	db.Unscoped().Where("email = ?", email).Delete(&models.User{})
}
