package services

import "github.com/navaneethk-000/url-shortener-backend/internal/models"

type IUrlService interface {

	// Create links
	Shorten(originalURL, customAlias string) (*models.Url, error)

	// Redirect users
	Resolve(shortCode, referrer, userAgent, ip string) (string, error)

	// Show graphs
	GetUrlStats(shortCode string) (*models.Url, []models.Click, error)
}
