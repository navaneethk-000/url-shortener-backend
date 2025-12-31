package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"github.com/stretchr/testify/assert"
)

// --- MOCK SERVICE ---
type mockUrlService struct {
	// Updated: Added two string parameters for QR colors
	ShortenFunc         func(string, string, uint64, string, string) (*models.Url, error)
	ResolveFunc         func(string, string, string, string) (string, error)
	GetUrlStatsFunc     func(string) (*models.Url, []models.Click, error)
	GenerateQRCodeFunc  func(string, string, string) ([]byte, error)
	GetUserUrlsFunc     func(uint64) ([]models.Url, error)
	DeleteShortLinkFunc func(string, uint64) error
	// New: Added UpdateStylesFunc
	UpdateStylesFunc func(string, uint64, string, string) error
}

func (m *mockUrlService) Shorten(o, c string, u uint64, qrc, qrbg string) (*models.Url, error) {
	return m.ShortenFunc(o, c, u, qrc, qrbg)
}
func (m *mockUrlService) Resolve(c, r, ua, ip string) (string, error) {
	return m.ResolveFunc(c, r, ua, ip)
}
func (m *mockUrlService) GetUrlStats(c string) (*models.Url, []models.Click, error) {
	return m.GetUrlStatsFunc(c)
}
func (m *mockUrlService) GenerateQRCode(c, fg, bg string) ([]byte, error) {
	return m.GenerateQRCodeFunc(c, fg, bg)
}
func (m *mockUrlService) GetUserUrls(u uint64) ([]models.Url, error) {
	return m.GetUserUrlsFunc(u)
}
func (m *mockUrlService) DeleteShortLink(c string, u uint64) error {
	return m.DeleteShortLinkFunc(c, u)
}
func (m *mockUrlService) UpdateStyles(c string, u uint64, qrc, qrbg string) error {
	return m.UpdateStylesFunc(c, u, qrc, qrbg)
}

func TestCreateShortUrl(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		body           string
		setupCtx       func(*gin.Context)
		mock           *mockUrlService
		expectedStatus int
	}{
		{
			name:     "Success",
			body:     `{"original_url": "http://google.com"}`,
			setupCtx: func(c *gin.Context) { c.Set("userID", uint64(1)) },
			mock: &mockUrlService{
				// Updated signature
				ShortenFunc: func(o, c string, u uint64, qrc, qrbg string) (*models.Url, error) {
					return &models.Url{ShortCode: "abc"}, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Fail - Unauthorized",
			body:           `{"original_url": "http://google.com"}`,
			setupCtx:       func(c *gin.Context) {},
			mock:           &mockUrlService{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Fail - Invalid JSON",
			body:           `{"bad": }`,
			setupCtx:       func(c *gin.Context) { c.Set("userID", uint64(1)) },
			mock:           &mockUrlService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Fail - Duplicate Alias",
			body:     `{"original_url": "http://google.com", "custom_alias": "taken"}`,
			setupCtx: func(c *gin.Context) { c.Set("userID", uint64(1)) },
			mock: &mockUrlService{
				// Updated signature
				ShortenFunc: func(o, c string, u uint64, qrc, qrbg string) (*models.Url, error) {
					return nil, errors.New("alias already in use")
				},
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request, _ = http.NewRequest("POST", "/api/shorten", bytes.NewBufferString(tt.body))
			tt.setupCtx(c)
			NewUrlHandler(tt.mock).CreateShortUrl(c)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// New Test for UpdateURLStyles
func TestUpdateURLStyles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest("PATCH", "/api/shorten/abc/styles", bytes.NewBufferString(`{"qr_color":"#000","qr_bg_color":"#fff"}`))
	c.Params = []gin.Param{{Key: "code", Value: "abc"}}
	c.Set("userID", uint64(1))

	mock := &mockUrlService{
		UpdateStylesFunc: func(c string, u uint64, qrc, qrbg string) error { return nil },
	}
	NewUrlHandler(mock).UpdateURLStyles(c)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockUrlService{
		ResolveFunc: func(c, r, ua, ip string) (string, error) {
			if c == "fail" {
				return "", errors.New("not found")
			}
			return "http://google.com", nil
		},
	}
	h := NewUrlHandler(mock)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = []gin.Param{{Key: "code", Value: "abc"}}
	c.Request, _ = http.NewRequest("GET", "/abc", nil)
	h.Redirect(c)
	assert.Equal(t, http.StatusFound, rec.Code)
}

func TestGetStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockUrlService{
		GetUrlStatsFunc: func(c string) (*models.Url, []models.Click, error) {
			return &models.Url{}, []models.Click{}, nil
		},
	}
	h := NewUrlHandler(mock)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = []gin.Param{{Key: "code", Value: "abc"}}
	h.GetStats(c)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDeleteShortUrl(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = []gin.Param{{Key: "code", Value: "abc"}}
	c.Set("userID", uint64(1))

	mock := &mockUrlService{
		DeleteShortLinkFunc: func(s string, u uint64) error { return nil },
	}
	NewUrlHandler(mock).DeleteShortUrl(c)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetUserUrls(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set("userID", uint64(1))

	mock := &mockUrlService{
		GetUserUrlsFunc: func(u uint64) ([]models.Url, error) { return []models.Url{}, nil },
	}
	NewUrlHandler(mock).GetUserUrls(c)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetQRCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockUrlService{
		// Updated signature
		GenerateQRCodeFunc: func(c, fg, bg string) ([]byte, error) {
			return []byte("fake-png"), nil
		},
	}
	h := NewUrlHandler(mock)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = []gin.Param{{Key: "code", Value: "abc"}}
	h.GetQRCode(c)
	assert.Equal(t, http.StatusOK, rec.Code)
}
