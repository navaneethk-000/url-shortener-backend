package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/navaneethk-000/url-shortener-backend/internal/handlers"
	"github.com/navaneethk-000/url-shortener-backend/internal/middleware"
)

func SetupRouter(urlHandler *handlers.UrlHandler, authHandler *handlers.AuthHandler) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Cors())

	api := r.Group("/api")
	{
		// Public Routes
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		api.GET("/stats/:code", urlHandler.GetStats)
		api.GET("/qr/:code", urlHandler.GetQRCode)

		// Protected Routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/shorten", urlHandler.CreateShortUrl)
			protected.GET("/user/urls", urlHandler.GetUserUrls)
		}

	}

	r.GET("/:code", urlHandler.Redirect)

	return r
}
