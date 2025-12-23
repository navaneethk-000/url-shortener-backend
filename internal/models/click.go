package models

import (
	"time"
)

type Click struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UrlID     uint64 `gorm:"index;not null" json:"url_id"`
	Referrer  string `gorm:"type:varchar(255)" json:"referrer"`
	UserAgent string `gorm:"type:text" json:"user_agent"`
	IPAddress string `gorm:"type:varchar(45)" json:"ip_address"`

	ClickedAt time.Time `gorm:"autoCreateTime" json:"clicked_at"`
}
