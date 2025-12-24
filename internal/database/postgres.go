package database

import (
	"log"

	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connects to Postgres and performs auto migrations
func InitDB(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Migrating database schema...")
	err = db.AutoMigrate(&models.User{}, &models.Url{}, &models.Click{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database connected and migrated successfully!")
	return db
}
