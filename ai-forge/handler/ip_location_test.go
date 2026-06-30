package handler

import (
	"testing"
)

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		// Private ranges
		{"private 10.0.0.1", "10.0.0.1", true},
		{"private 10.255.255.255", "10.255.255.255", true},
		{"private 172.16.0.1", "172.16.0.1", true},
		{"private 172.31.255.255", "172.31.255.255", true},
		{"private 192.168.0.1", "192.168.0.1", true},
		{"private 192.168.255.255", "192.168.255.255", true},
		{"private 127.0.0.1 loopback", "127.0.0.1", true},
		{"private 127.255.255.255", "127.255.255.255", true},

		// Public IPs - should return false
		{"public 8.8.8.8", "8.8.8.8", false},
		{"public 1.1.1.1", "1.1.1.1", false},
		{"public 203.0.113.1", "203.0.113.1", false},
		{"public 172.15.255.255 just below private range", "172.15.255.255", false},
		{"public 172.32.0.0 just above private range", "172.32.0.0", false},

		// Invalid formats - should return false
		{"empty string", "", false},
		{"not an IP", "abc.def.ghi.jkl", false},
		{"just numbers", "12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPrivateIP(tt.ip)
			if result != tt.expected {
				t.Errorf("isPrivateIP(%q) = %v; want %v", tt.ip, result, tt.expected)
			}
		})
	}
}
