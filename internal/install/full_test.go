package install

import (
	"testing"
)

func TestSSIDFilename(t *testing.T) {
	tests := []struct {
		ssid     string
		expected string
	}{
		{"MyWiFi", "MyWiFi"},
		{"Home Network", "Home Network"},
		{"guest-5g", "guest-5g"},
		{"SomeWifi 12345 - 2.4", "=536f6d6557696669203132333435202d20322e34"},
		{"Café", "=436166c3a9"},
	}

	for _, tt := range tests {
		got := SSIDFilename(tt.ssid)
		if got != tt.expected {
			t.Errorf("SSIDFilename(%q) = %q, want %q", tt.ssid, got, tt.expected)
		}
	}
}
