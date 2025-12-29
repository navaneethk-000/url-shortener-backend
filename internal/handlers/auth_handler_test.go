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

// This mocks IAuthService so we can test the Handler without a DB.
type mockAuthService struct {
	RegisterFunc func(email, password string) (*models.User, error)
	LoginFunc    func(email, password string) (string, error)
}

func (m *mockAuthService) Register(email, password string) (*models.User, error) {
	return m.RegisterFunc(email, password)
}

func (m *mockAuthService) Login(email, password string) (string, error) {
	return m.LoginFunc(email, password)
}

// Register
func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    string
		setupMock      func() *mockAuthService
		expectedStatus int
	}{
		{
			name:        "Success - User Created",
			requestBody: `{"email": "new@test.com", "password": "password123"}`,
			setupMock: func() *mockAuthService {
				return &mockAuthService{
					RegisterFunc: func(email, password string) (*models.User, error) {
						return &models.User{ID: 1, Email: email}, nil
					},
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:        "Fail - Invalid Email Format",
			requestBody: `{"email": "not-an-email", "password": "password123"}`,
			setupMock: func() *mockAuthService {
				return &mockAuthService{} // Should not be called
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Fail - Password Too Short",
			requestBody: `{"email": "valid@test.com", "password": "123"}`, // Min 6 chars required
			setupMock: func() *mockAuthService {
				return &mockAuthService{}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Fail - Email Already Exists",
			requestBody: `{"email": "existing@test.com", "password": "password123"}`,
			setupMock: func() *mockAuthService {
				return &mockAuthService{
					RegisterFunc: func(email, password string) (*models.User, error) {
						return nil, errors.New("email already in use")
					},
				}
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request, _ = http.NewRequest("POST", "/api/register", bytes.NewBufferString(tt.requestBody))
			c.Request.Header.Set("Content-Type", "application/json")

			// Execute
			handler := NewAuthHandler(tt.setupMock())
			handler.Register(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// Login
func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    string
		setupMock      func() *mockAuthService
		expectedStatus int
	}{
		{
			name:        "Success - Login Returns Token",
			requestBody: `{"email": "user@test.com", "password": "password123"}`,
			setupMock: func() *mockAuthService {
				return &mockAuthService{
					LoginFunc: func(email, password string) (string, error) {
						return "fake-jwt-token", nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Fail - Missing Password",
			requestBody: `{"email": "user@test.com"}`,
			setupMock: func() *mockAuthService {
				return &mockAuthService{}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Fail - Wrong Credentials",
			requestBody: `{"email": "user@test.com", "password": "wrongpassword"}`,
			setupMock: func() *mockAuthService {
				return &mockAuthService{
					LoginFunc: func(email, password string) (string, error) {
						return "", errors.New("invalid credentials")
					},
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request, _ = http.NewRequest("POST", "/api/login", bytes.NewBufferString(tt.requestBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler := NewAuthHandler(tt.setupMock())
			handler.Login(c)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
