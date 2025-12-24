package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/navaneethk-000/url-shortener-backend/internal/handlers"
	"github.com/navaneethk-000/url-shortener-backend/internal/middleware"
)

func SetupRouter(urlHandler *handlers.UrlHandler) *gin.Engine {
	r := gin.Default()

	// Apply Middleware
	r.Use(middleware.Cors())

	// API Group
	api := r.Group("/api")
	{
		api.POST("/shorten", urlHandler.CreateShortUrl)
		api.GET("/stats/:code", urlHandler.GetStats)
	}

	// Root Redirect
	r.GET("/:code", urlHandler.Redirect)

	return r
}
