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
	userRepo := repository.NewUserRepository(db)

	// Service Layer (Injects Repos)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set in .env")
	}

	urlService := services.NewUrlService(urlRepo, clickRepo)
	authService := services.NewAuthService(userRepo, jwtSecret)

	// Handler Layer (Injects Service)
	urlHandler := handlers.NewUrlHandler(urlService)
	authHandler := handlers.NewAuthHandler(authService)

	// Setup Router
	router := routes.SetupRouter(urlHandler, authHandler)

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
