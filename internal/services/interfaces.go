package services

import "github.com/navaneethk-000/url-shortener-backend/internal/models"

type IUrlService interface {
	Shorten(originalURL, customAlias string, userID uint64) (*models.Url, error)
	Resolve(shortCode, referrer, userAgent, ip string) (string, error)
	GetUrlStats(shortCode string) (*models.Url, []models.Click, error)
	GenerateQRCode(shortCode string) ([]byte, error)
	GetUserUrls(userID uint64) ([]models.Url, error)
}

type IAuthService interface {
	Register(email, password string) (*models.User, error)
	Login(email, password string) (string, error)
}
