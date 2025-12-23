package repository

import (
	"errors"

	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"gorm.io/gorm"
)

type UrlRepository struct {
	DB *gorm.DB
}

func NewUrlRepository(db *gorm.DB) *UrlRepository {
	return &UrlRepository{DB: db}
}

// Saves a new URL to the database
func (r *UrlRepository) Create(url *models.Url) error {
	return r.DB.Create(url).Error
}

// Looks up a URL by its alias
func (r *UrlRepository) FindByShortCode(code string) (*models.Url, error) {
	var url models.Url
	// SELECT * FROM urls WHERE short_code = 'code' LIMIT 1;
	result := r.DB.Where("short_code = ?", code).First(&url)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &url, nil
}

func (r *UrlRepository) IncrementClicks(urlID uint64) error {
	return r.DB.Model(&models.Url{}).
		Where("id = ?", urlID).
		UpdateColumn("total_clicks", gorm.Expr("total_clicks + ?", 1)).
		Error
}
