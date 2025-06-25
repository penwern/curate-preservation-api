package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/penwern/curate-preservation-api/pkg/config"
	"github.com/penwern/curate-preservation-api/pkg/logger"
)

func TestTrustedIPAuthentication(t *testing.T) {
	logger.Initialize("debug", "/tmp/curate-preservation-api.log")

	// Test configuration with trusted IPs
	cfg := config.Config{
		DBType:       "sqlite3",
		DBConnection: ":memory:",
		Port:         8080,
		TrustedIPs:   []string{"127.0.0.1", "192.168.1.0/24", "10.0.0.0/8"},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer server.Shutdown()

	tests := []struct {
		name           string
		remoteAddr     string
		expectAuth     bool
		expectedStatus int
	}{
		{
			name:           "Trusted IP - localhost",
			remoteAddr:     "127.0.0.1:12345",
			expectAuth:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Trusted IP - private network",
			remoteAddr:     "192.168.1.50:54321",
			expectAuth:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Trusted IP - 10.x network",
			remoteAddr:     "10.1.2.3:44444",
			expectAuth:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Untrusted IP - public",
			remoteAddr:     "8.8.8.8:80",
			expectAuth:     false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Untrusted IP - different private",
			remoteAddr:     "172.16.1.1:443",
			expectAuth:     false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/v1/preservation-configs", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Set the remote address to simulate the client IP
			req.RemoteAddr = tt.remoteAddr

			rr := httptest.NewRecorder()
			server.router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectAuth && rr.Code == http.StatusOK {
				t.Logf("✓ Trusted IP %s successfully bypassed authentication", tt.remoteAddr)
			} else if !tt.expectAuth && rr.Code == http.StatusUnauthorized {
				t.Logf("✓ Untrusted IP %s correctly required authentication", tt.remoteAddr)
			}
		})
	}
}

func TestIPParsingFunctions(t *testing.T) {
	tests := []struct {
		name     string
		ipStr    string
		clientIP string
		expected bool
	}{
		{
			name:     "IPv4 exact match",
			ipStr:    "192.168.1.1",
			clientIP: "192.168.1.1",
			expected: true,
		},
		{
			name:     "IPv4 CIDR match",
			ipStr:    "192.168.1.0/24",
			clientIP: "192.168.1.100",
			expected: true,
		},
		{
			name:     "IPv4 CIDR no match",
			ipStr:    "192.168.1.0/24",
			clientIP: "192.168.2.100",
			expected: false,
		},
		{
			name:     "IPv6 localhost",
			ipStr:    "::1",
			clientIP: "::1",
			expected: true,
		},
		{
			name:     "Invalid client IP",
			ipStr:    "192.168.1.0/24",
			clientIP: "invalid-ip",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isIPTrusted(tt.clientIP, []string{tt.ipStr})
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for IP %s against %s", tt.expected, result, tt.clientIP, tt.ipStr)
			}
		})
	}
}
