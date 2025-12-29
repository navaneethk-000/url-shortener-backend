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
	ShortenFunc         func(string, string, uint64) (*models.Url, error)
	ResolveFunc         func(string, string, string, string) (string, error)
	GetUrlStatsFunc     func(string) (*models.Url, []models.Click, error)
	GenerateQRCodeFunc  func(string) ([]byte, error)
	GetUserUrlsFunc     func(uint64) ([]models.Url, error)
	DeleteShortLinkFunc func(string, uint64) error
}

func (m *mockUrlService) Shorten(o, c string, u uint64) (*models.Url, error) {
	return m.ShortenFunc(o, c, u)
}
func (m *mockUrlService) Resolve(c, r, ua, ip string) (string, error) {
	return m.ResolveFunc(c, r, ua, ip)
}
func (m *mockUrlService) GetUrlStats(c string) (*models.Url, []models.Click, error) {
	return m.GetUrlStatsFunc(c)
}
func (m *mockUrlService) GenerateQRCode(c string) ([]byte, error) {
	return m.GenerateQRCodeFunc(c)
}
func (m *mockUrlService) GetUserUrls(u uint64) ([]models.Url, error) {
	return m.GetUserUrlsFunc(u)
}
func (m *mockUrlService) DeleteShortLink(c string, u uint64) error {
	return m.DeleteShortLinkFunc(c, u)
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
				ShortenFunc: func(o, c string, u uint64) (*models.Url, error) {
					return &models.Url{ShortCode: "abc"}, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Fail - Unauthorized",
			body:           `{"original_url": "http://google.com"}`,
			setupCtx:       func(c *gin.Context) {}, // No UserID
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
				ShortenFunc: func(o, c string, u uint64) (*models.Url, error) {
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

	// Case 1: Success
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = []gin.Param{{Key: "code", Value: "abc"}}
	c.Request, _ = http.NewRequest("GET", "/abc", nil)
	h.Redirect(c)
	assert.Equal(t, http.StatusFound, rec.Code)

	// Case 2: Fail
	rec2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(rec2)
	c2.Params = []gin.Param{{Key: "code", Value: "fail"}}
	c2.Request, _ = http.NewRequest("GET", "/fail", nil)
	h.Redirect(c2)
	assert.Equal(t, http.StatusNotFound, rec2.Code)
}

func TestGetStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockUrlService{
		GetUrlStatsFunc: func(c string) (*models.Url, []models.Click, error) {
			if c == "fail" {
				return nil, nil, errors.New("not found")
			}
			return &models.Url{}, []models.Click{}, nil
		},
	}
	h := NewUrlHandler(mock)

	// Case 1: Success
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = []gin.Param{{Key: "code", Value: "abc"}}
	h.GetStats(c)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Case 2: Fail
	rec2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(rec2)
	c2.Params = []gin.Param{{Key: "code", Value: "fail"}}
	h.GetStats(c2)
	assert.Equal(t, http.StatusNotFound, rec2.Code)
}

func TestDeleteShortUrl(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name     string
		setupCtx func(*gin.Context)
		mock     *mockUrlService
		status   int
	}{
		{
			name:     "Success",
			setupCtx: func(c *gin.Context) { c.Set("userID", uint64(1)) },
			mock:     &mockUrlService{DeleteShortLinkFunc: func(s string, u uint64) error { return nil }},
			status:   http.StatusOK,
		},
		{
			name:     "Unauthorized",
			setupCtx: func(c *gin.Context) {},
			mock:     &mockUrlService{},
			status:   http.StatusUnauthorized,
		},
		{
			name:     "Forbidden (Not Owner)",
			setupCtx: func(c *gin.Context) { c.Set("userID", uint64(1)) },
			mock: &mockUrlService{DeleteShortLinkFunc: func(s string, u uint64) error {
				return errors.New("unauthorized: you do not own this link")
			}},
			status: http.StatusForbidden,
		},
		{
			name:     "Not Found",
			setupCtx: func(c *gin.Context) { c.Set("userID", uint64(1)) },
			mock: &mockUrlService{DeleteShortLinkFunc: func(s string, u uint64) error {
				return errors.New("not found")
			}},
			status: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Params = []gin.Param{{Key: "code", Value: "abc"}}
			tt.setupCtx(c)
			NewUrlHandler(tt.mock).DeleteShortUrl(c)
			assert.Equal(t, tt.status, rec.Code)
		})
	}
}

func TestGetUserUrls(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Case 1: Success
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set("userID", uint64(1))

	mock := &mockUrlService{
		GetUserUrlsFunc: func(u uint64) ([]models.Url, error) { return []models.Url{}, nil },
	}
	NewUrlHandler(mock).GetUserUrls(c)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Case 2: Unauthorized
	rec2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(rec2)
	NewUrlHandler(mock).GetUserUrls(c2)
	assert.Equal(t, http.StatusUnauthorized, rec2.Code)
}

func TestGetQRCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockUrlService{
		GenerateQRCodeFunc: func(c string) ([]byte, error) {
			if c == "fail" {
				return nil, errors.New("err")
			}
			return []byte("fake-png"), nil
		},
	}
	h := NewUrlHandler(mock)

	// Case 1: Success
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = []gin.Param{{Key: "code", Value: "abc"}}
	h.GetQRCode(c)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Case 2: Fail
	rec2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(rec2)
	c2.Params = []gin.Param{{Key: "code", Value: "fail"}}
	h.GetQRCode(c2)
	assert.Equal(t, http.StatusNotFound, rec2.Code)
}
