package services

import "testing"

func TestGetLocation_Localhost(t *testing.T) {
	country, _, err := GetLocation("127.0.0.1")
	if err != nil {
		t.Fatalf("Error getting location: %v", err)
	}
	if country != "Localhost" {
		t.Errorf("Expected Localhost, got %s", country)
	}
}
