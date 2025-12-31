package middleware

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	config := cors.DefaultConfig()

	// Default Origins
	origins := []string{
		"http://localhost:3000", // Nuxt Dev
		"http://localhost",      // Local Docker Production
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL != "" {
		origins = append(origins, frontendURL)
	}

	config.AllowOrigins = origins
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}

	return cors.New(config)
}
