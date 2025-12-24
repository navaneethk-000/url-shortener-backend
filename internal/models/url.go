package models

import (
	"time"
)

type Url struct {
	ID          uint64  `gorm:"primaryKey;autoIncrement" json:"id"`
	OriginalURL string  `gorm:"type:text;not null" json:"original_url"`
	ShortCode   string  `gorm:"type:varchar(10);uniqueIndex;not null" json:"short_code"`
	TotalClicks int     `gorm:"default:0" json:"total_clicks"`
	Clicks      []Click `gorm:"foreignKey:UrlID" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID uint64 `gorm:"not null;index" json:"user_id"`
	User   User   `gorm:"foreignKey:UserID" json:"-"`
}
