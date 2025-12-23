package base62

import "testing"

func TestEncode(t *testing.T) {
	result := Encode(100)
	if result != "1C" {
		t.Errorf("Expected 1C, got %s", result)
	}
}
