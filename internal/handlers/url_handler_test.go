package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreateShortUrl_InvalidJSON(t *testing.T) {

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	// Creates handler with nil service
	handler := NewUrlHandler(nil)

	// Creates Request with the missing "original_url"
	jsonBody := `{"custom_alias": "fail"}`
	req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	ctx.Request = req

	handler.CreateShortUrl(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}
