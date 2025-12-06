package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dangogh/silver-eureka/internal/database"
)

func TestServeHTTP(t *testing.T) {
	// Create temporary database
	dbPath := "/tmp/test_handler.db"
	defer os.Remove(dbPath)

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	h := New(db)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	// Serve the request
	h.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Expected Content-Type text/plain, got %s", w.Header().Get("Content-Type"))
	}

	// Verify log was created
	logs, err := db.GetLogs(1)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	if logs[0].IPAddress != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", logs[0].IPAddress)
	}

	if logs[0].URL != "/test/path" {
		t.Errorf("Expected URL /test/path, got %s", logs[0].URL)
	}
}

func TestGetIPAddress(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		xff            string
		xri            string
		expectedIP     string
	}{
		{
			name:       "RemoteAddr only",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For single IP",
			remoteAddr: "192.168.1.1:12345",
			xff:        "203.0.113.1",
			expectedIP: "203.0.113.1",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			remoteAddr: "192.168.1.1:12345",
			xff:        "203.0.113.1, 198.51.100.1, 192.0.2.1",
			expectedIP: "203.0.113.1",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "192.168.1.1:12345",
			xri:        "203.0.113.2",
			expectedIP: "203.0.113.2",
		},
		{
			name:       "X-Forwarded-For takes precedence over X-Real-IP",
			remoteAddr: "192.168.1.1:12345",
			xff:        "203.0.113.1",
			xri:        "203.0.113.2",
			expectedIP: "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}

			ip := getIPAddress(req)
			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}
