package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/navaneethk-000/url-shortener-backend/internal/database"
	"github.com/navaneethk-000/url-shortener-backend/internal/handlers"
	"github.com/navaneethk-000/url-shortener-backend/internal/repository"
	"github.com/navaneethk-000/url-shortener-backend/internal/routes"
	"github.com/navaneethk-000/url-shortener-backend/internal/services"
)

func main() {
	// Load Environment Variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using system defaults")
	}

	// Database Connection
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	db := database.InitDB(dsn)

	// Repo Layer
	urlRepo := repository.NewUrlRepository(db)
	clickRepo := repository.NewClickRepository(db)

	// Service Layer (Injects Repos)
	urlService := services.NewUrlService(urlRepo, clickRepo)

	// Handler Layer (Injects Service)
	urlHandler := handlers.NewUrlHandler(urlService)

	// Setup Router
	router := routes.SetupRouter(urlHandler)

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
