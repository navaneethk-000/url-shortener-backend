package database

import (
	"log"
	"time"

	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(dsn string) *gorm.DB {
	var db *gorm.DB
	var err error

	// Wait up to 2 seconds for DB to be ready
	for i := 1; i <= 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("Database connected successfully!")
			break
		}
		log.Printf("Attempt %d/10: Database not ready yet... waiting 2s", i)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Migrating database schema...")
	db.AutoMigrate(&models.User{}, &models.Url{}, &models.Click{})
	return db
}
