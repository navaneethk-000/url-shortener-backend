package services

import (
	"errors"
	"time"

	"github.com/navaneethk-000/url-shortener-backend/internal/base62"
	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"github.com/navaneethk-000/url-shortener-backend/internal/repository"
)

type UrlService struct {
	UrlRepo   *repository.UrlRepository
	ClickRepo *repository.ClickRepository
}

// Factory
func NewUrlService(uRepo *repository.UrlRepository, cRepo *repository.ClickRepository) *UrlService {
	return &UrlService{
		UrlRepo:   uRepo,
		ClickRepo: cRepo,
	}
}

// Creates a new short link
func (s *UrlService) Shorten(originalURL, customAlias string) (*models.Url, error) {
	// Custom Alias Logic
	if customAlias != "" {
		existing, _ := s.UrlRepo.FindByShortCode(customAlias)
		if existing != nil {
			return nil, errors.New("alias already in use")
		}
		newUrl := &models.Url{
			OriginalURL: originalURL,
			ShortCode:   customAlias,
			CreatedAt:   time.Now(),
		}
		return newUrl, s.UrlRepo.Create(newUrl)
	}

	// Base62 Logic
	newUrl := &models.Url{
		OriginalURL: originalURL,
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
