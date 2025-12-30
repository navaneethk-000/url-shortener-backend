package services

import (
	"encoding/json"
	"net/http"
	"time"
)

type GeoData struct {
	Country string `json:"country"`
	City    string `json:"city"`
	Status  string `json:"status"`
}

// Fetches geo-data from an external API
func GetLocation(ip string) (string, string, error) {
	// Handle Localhost
	if ip == "127.0.0.1" || ip == "::1" {
		return "Localhost", "Localhost", nil
	}

	// Call External API
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		return "Unknown", "Unknown", err
	}
	defer resp.Body.Close()

	// Decode Response
	var data GeoData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "Unknown", "Unknown", err
	}

	if data.Status == "fail" {
		return "Unknown", "Unknown", nil
	}

	return data.Country, data.City, nil
}
