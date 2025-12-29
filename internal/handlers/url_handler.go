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

// CreateSortUrl handles POST /api/shorten
func (h *UrlHandler) CreateShortUrl(c *gin.Context) {
	var req CreateUrlRequest

	// JSON validation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get UserID from middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Call Service
	url, err := h.Service.Shorten(req.OriginalURL, req.CustomAlias, userID.(uint64))
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

// Redirect handles GET /:code
func (h *UrlHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	// Grab analytics data from request headers
	referrer := c.Request.Referer()
	userAgent := c.Request.UserAgent()
	ip := c.ClientIP()

	originalUrl, err := h.Service.Resolve(code, referrer, userAgent, ip)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.Redirect(http.StatusFound, originalUrl)
}

// GetStats handles GET /api/stats/:code
func (h *UrlHandler) GetStats(c *gin.Context) {
	code := c.Param("code")

	url, clicks, err := h.Service.GetUrlStats(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url_data":  url,
		"analytics": clicks,
	})
}

// DeleteShortUrl handles DELETE /api/shorten/:code
func (h *UrlHandler) DeleteShortUrl(c *gin.Context) {
	code := c.Param("code")

	// Get UserID from middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Call Service
	err := h.Service.DeleteShortLink(code, userID.(uint64))
	if err != nil {
		if err.Error() == "unauthorized: you do not own this link" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

// GetUserUrls handles GET /api/user/urls
func (h *UrlHandler) GetUserUrls(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	urls, err := h.Service.GetUserUrls(userID.(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch URLs"})
		return
	}
	c.JSON(http.StatusOK, urls)
}

// GetQRCode handles GET /api/qr/:code
func (h *UrlHandler) GetQRCode(c *gin.Context) {
	code := c.Param("code")

	pngBytes, err := h.Service.GenerateQRCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.Data(http.StatusOK, "image/png", pngBytes)
}
