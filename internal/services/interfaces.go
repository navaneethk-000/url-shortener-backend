package services

import "github.com/navaneethk-000/url-shortener-backend/internal/models"

type IUrlService interface {
	Shorten(originalURL, customAlias string, userID uint64, qrColor, qrBgColor string) (*models.Url, error)
	UpdateStyles(shortCode string, userID uint64, qrColor, qrBgColor string) error
	Resolve(shortCode, referrer, userAgent, ip string) (string, error)
	GetUrlStats(shortCode string) (*models.Url, []models.Click, error)
	GenerateQRCode(shortCode, fgColor, bgColor string) ([]byte, error)
	GetUserUrls(userID uint64) ([]models.Url, error)
	DeleteShortLink(shortCode string, userID uint64) error
}

type IAuthService interface {
	Register(name, email, password string) (*models.User, error)
	Login(email, password string) (string, error)
}
