package utils

import (
	"net/http/httptest"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string // X-Forwarded-For header ("" = not sent)
		trustProxy string // value of TRUST_PROXY env var
		want       string
	}{
		{"direct connection, port stripped", "203.0.113.5:5555", "", "", "203.0.113.5"},
		{"header ignored when proxy not trusted", "203.0.113.5:5555", "198.51.100.7", "", "203.0.113.5"},
		{"header ignored when TRUST_PROXY is not exactly true", "203.0.113.5:5555", "198.51.100.7", "yes", "203.0.113.5"},
		{"trusted proxy, single address", "10.0.0.2:4000", "198.51.100.7", "true", "198.51.100.7"},
		{"trusted proxy, last entry wins", "10.0.0.2:4000", "1.2.3.4, 198.51.100.7", "true", "198.51.100.7"},
		{"trusted proxy, garbage header falls back", "10.0.0.2:4000", "not-an-ip", "true", "10.0.0.2"},
		{"trusted proxy, no header", "10.0.0.2:4000", "", "true", "10.0.0.2"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TRUST_PROXY", tc.trustProxy)

			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tc.remoteAddr
			if tc.xff != "" {
				req.Header.Set("X-Forwarded-For", tc.xff)
			}

			got := GetClientIP(req)

			if got != tc.want {
				t.Errorf("GetClientIP() = %q, want %q", got, tc.want)
			}
		})
	}
}
