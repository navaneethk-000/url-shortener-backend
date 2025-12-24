package repository

import (
	"errors"

	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// saves a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.DB.Create(user).Error
}

// looks up a user for login
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.DB.Where("email = ?", email).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if user not found instead of returning error
		}
		return nil, result.Error
	}

	return &user, nil
}
