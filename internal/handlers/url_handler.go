package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/navaneethk-000/url-shortener-backend/internal/services"
)

type UrlHandler struct {
	Service services.IUrlService
}

func NewUrlHandler(s services.IUrlService) *UrlHandler {
	return &UrlHandler{Service: s}
}

// Request Payload
type CreateUrlRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
	CustomAlias string `json:"custom_alias"`
}

func (h *UrlHandler) CreateShortUrl(c *gin.Context) {
	var req CreateUrlRequest

	// JSON validation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Call Service
	url, err := h.Service.Shorten(req.OriginalURL, req.CustomAlias)
	if err != nil {
		if err.Error() == "alias already in use" {
			c.JSON(http.StatusConflict, gin.H{"error": "Alias already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to shorten URL"})
		return
	}

	c.JSON(http.StatusOK, url)
}
