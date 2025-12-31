package services

import (
	"errors"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"time"

	"github.com/navaneethk-000/url-shortener-backend/internal/base62"
	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"github.com/navaneethk-000/url-shortener-backend/internal/repository"

	"github.com/skip2/go-qrcode"
)

type UrlService struct {
	UrlRepo   *repository.UrlRepository
	ClickRepo *repository.ClickRepository
}

func (s *UrlService) DeleteShortLink(shortCode string, userID uint64) error {
	// Find the URL
	url, err := s.UrlRepo.FindByShortCode(shortCode)
	if err != nil {
		return err
	}
	if url == nil {
		return errors.New("URL not found")
	}

	// Security Check
	if url.UserID != userID {
		return errors.New("unauthorized: you do not own this link")
	}

	// Delete
	return s.UrlRepo.Delete(url.ID)
}

// Fetch all URLs created by a specific user
func (s *UrlService) GetUserUrls(userID uint64) ([]models.Url, error) {
	var urls []models.Url
	result := s.UrlRepo.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&urls)
	return urls, result.Error
}

func parseHexColor(s string) color.RGBA {
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}
	// Default to Black if invalid
	if len(s) != 6 {
		return color.RGBA{0, 0, 0, 255}
	}

	hex, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return color.RGBA{0, 0, 0, 255}
	}

	return color.RGBA{
		R: uint8(hex >> 16),
		G: uint8((hex >> 8) & 0xFF),
		B: uint8(hex & 0xFF),
		A: 255,
	}
}

func (s *UrlService) GenerateQRCode(shortCode, fgColor, bgColor string) ([]byte, error) {
	// Check if URL exists
	url, err := s.UrlRepo.FindByShortCode(shortCode)
	if err != nil {
		return nil, err
	}
	if url == nil {
		return nil, errors.New("URL not found")
	}

	// Create the full URL
	baseUrl := os.Getenv("BASE_URL")
	if baseUrl == "" {
		baseUrl = "http://localhost:8080"
	}
	fullURL := fmt.Sprintf("%s/%s", baseUrl, shortCode)

	qr, err := qrcode.New(fullURL, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	qr.ForegroundColor = color.RGBA{0, 0, 0, 255}
	qr.BackgroundColor = color.RGBA{255, 255, 255, 255}

	if fgColor != "" {
		qr.ForegroundColor = parseHexColor(fgColor)
	}
	if bgColor != "" {
		qr.BackgroundColor = parseHexColor(bgColor)
	}

	return qr.PNG(256)
}

func NewUrlService(uRepo *repository.UrlRepository, cRepo *repository.ClickRepository) *UrlService {
	return &UrlService{
		UrlRepo:   uRepo,
		ClickRepo: cRepo,
	}
}

// Creates a new short link
func (s *UrlService) Shorten(originalURL, customAlias string, userID uint64, qrColor, qrBgColor string) (*models.Url, error) {
	// Custom Alias Logic
	if customAlias != "" {
		existing, _ := s.UrlRepo.FindByShortCode(customAlias)
		if existing != nil {
			return nil, errors.New("alias already in use")
		}
		newUrl := &models.Url{
			OriginalURL: originalURL,
			ShortCode:   customAlias,
			UserID:      userID,
			QRColor:     qrColor,
			QRBgColor:   qrBgColor,
			CreatedAt:   time.Now(),
		}
		return newUrl, s.UrlRepo.Create(newUrl)
	}

	// Base62 Logic
	newUrl := &models.Url{
		OriginalURL: originalURL,
		UserID:      userID,
		QRColor:     qrColor,
		QRBgColor:   qrBgColor,
		CreatedAt:   time.Now(),
	}
	// Save first to generate ID
	err := s.UrlRepo.Create(newUrl)
	if err != nil {
		return nil, err
	}

	newUrl.ShortCode = base62.Encode(newUrl.ID)
	return newUrl, s.UrlRepo.DB.Save(newUrl).Error
}

func (s *UrlService) UpdateStyles(shortCode string, userID uint64, qrColor, qrBgColor string) error {
	// Find the URL
	url, err := s.UrlRepo.FindByShortCode(shortCode)
	if err != nil {
		return err
	}
	if url == nil {
		return errors.New("URL not found")
	}

	// Ensure the user owns this link
	if url.UserID != userID {
		return errors.New("unauthorized: you do not own this link")
	}

	// Update only the color fields
	return s.UrlRepo.DB.Model(url).Updates(map[string]interface{}{
		"qr_color":    qrColor,
		"qr_bg_color": qrBgColor,
	}).Error
}

// Find URL and logs click (Async)
func (s *UrlService) Resolve(shortCode string, referrer string, userAgent string, ip string) (string, error) {
	url, err := s.UrlRepo.FindByShortCode(shortCode)
	if err != nil {
		return "", err
	}
	if url == nil {
		return "", errors.New("URL not found")
	}

	// Async Analytics
	go func() {
		fmt.Println(" Resolve: Incrementing clicks...")
		_ = s.UrlRepo.IncrementClicks(url.ID)
		click := &models.Click{
			UrlID:     url.ID,
			Referrer:  referrer,
			UserAgent: userAgent,
			IPAddress: ip,
		}

		if WorkerInstance != nil {
			fmt.Println("🔍 Resolve: Pushing to worker queue...")
			WorkerInstance.Push(click)
		} else {
			fmt.Println("ERROR: WorkerInstance is NIL!")
		}

	}()

	return url.OriginalURL, nil
}

// Fetches data for the dashboard
func (s *UrlService) GetUrlStats(shortCode string) (*models.Url, []models.Click, error) {
	url, err := s.UrlRepo.FindByShortCode(shortCode)
	if err != nil {
		return nil, nil, err
	}
	if url == nil {
		return nil, nil, errors.New("URL not found")
	}

	clicks, err := s.ClickRepo.GetStatsByUrlID(url.ID)
	if err != nil {
		return nil, nil, err
	}

	return url, clicks, nil
}
