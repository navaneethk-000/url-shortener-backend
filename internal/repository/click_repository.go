package repository

import (
	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"gorm.io/gorm"
)

type ClickRepository struct {
	DB *gorm.DB
}

func NewClickRepository(db *gorm.DB) *ClickRepository {
	return &ClickRepository{DB: db}
}

// Records a new visit
func (r *ClickRepository) SaveClick(click *models.Click) error {
	return r.DB.Create(click).Error
}

// Fetches the history of clicks for a specific link
func (r *ClickRepository) GetStatsByUrlID(urlID uint64) ([]models.Click, error) {
	var clicks []models.Click

	// SELECT * FROM clicks WHERE url_id = ? ORDER BY clicked_at DESC;
	result := r.DB.Where("url_id = ?", urlID).
		Order("clicked_at desc").
		Find(&clicks)

	return clicks, result.Error
}
