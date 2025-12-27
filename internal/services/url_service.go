package services

import (
	"errors"
	"fmt"
	"os"
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

// Fetch all URLs created by a specific user
func (s *UrlService) GetUserUrls(userID uint64) ([]models.Url, error) {
	var urls []models.Url
	result := s.UrlRepo.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&urls)
	return urls, result.Error
}

func (s *UrlService) GenerateQRCode(shortCode string) ([]byte, error) {
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

	// Generate QR code (256x256 pixels, Medium error correction)
	png, err := qrcode.Encode(fullURL, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}

	return png, nil
}

// Factory
func NewUrlService(uRepo *repository.UrlRepository, cRepo *repository.ClickRepository) *UrlService {
	return &UrlService{
		UrlRepo:   uRepo,
		ClickRepo: cRepo,
	}
}

// Creates a new short link
func (s *UrlService) Shorten(originalURL, customAlias string, userID uint64) (*models.Url, error) {
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
			CreatedAt:   time.Now(),
		}
		return newUrl, s.UrlRepo.Create(newUrl)
	}

	// Base62 Logic
	newUrl := &models.Url{
		OriginalURL: originalURL,
		UserID:      userID,
		CreatedAt:   time.Now(),
	}
	// Save first to generate ID
	err := s.UrlRepo.Create(newUrl)
	if err != nil {
		return nil, err
	}

	// Update with generated code
	newUrl.ShortCode = base62.Encode(newUrl.ID)
	return newUrl, s.UrlRepo.DB.Save(newUrl).Error
}

// Resolve finds URL and logs click (Async)
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
		_ = s.UrlRepo.IncrementClicks(url.ID)
		click := &models.Click{
			UrlID:     url.ID,
			Referrer:  referrer,
			UserAgent: userAgent,
			IPAddress: ip,
		}
		_ = s.ClickRepo.SaveClick(click)
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
