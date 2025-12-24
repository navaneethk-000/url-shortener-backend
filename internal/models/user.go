package models

import "time"

type User struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Email    string `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password string `gorm:"type:text;not null" json:"-"`
	Urls     []Url  `gorm:"foreignKey:UserID" json:"urls,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
