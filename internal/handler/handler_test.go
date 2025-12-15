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
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

	h := New(db)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	// Serve the request
	h.ServeHTTP(w, req)

	// Check response - should return 404 for unmatched routes
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Expected Content-Type text/plain, got %s", w.Header().Get("Content-Type"))
	}

	// Check response body
	expectedBody := "404 page not found\n"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
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
		name       string
		remoteAddr string
		xff        string
		xri        string
		expectedIP string
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
		{
			name:       "RemoteAddr without port",
			remoteAddr: "192.168.1.1",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For with whitespace",
			remoteAddr: "192.168.1.1:12345",
			xff:        " 203.0.113.1 , 198.51.100.1",
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

func TestServeHTTP_DatabaseError(t *testing.T) {
	// Create temporary database
	dbPath := "/tmp/test_handler_error.db"
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Close the database to force an error
	if err := db.Close(); err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}

	h := New(db)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	// Serve the request - should return 503 due to database error
	h.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	// Check response body contains error message
	body := w.Body.String()
	if body != `{"error":"logging temporarily unavailable","status":"degraded"}` {
		t.Errorf("Expected degraded error message, got %q", body)
	}
}

func TestServeHTTP_WithHeaders(t *testing.T) {
	// Create temporary database
	dbPath := "/tmp/test_handler_headers.db"
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

	h := New(db)

	// Create test request with X-Forwarded-For header
	req := httptest.NewRequest(http.MethodPost, "/api/endpoint?param=value", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	w := httptest.NewRecorder()

	// Serve the request
	h.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	// Verify log was created with correct IP from X-Forwarded-For
	logs, err := db.GetLogs(1)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	if logs[0].IPAddress != "203.0.113.1" {
		t.Errorf("Expected IP 203.0.113.1 (from X-Forwarded-For), got %s", logs[0].IPAddress)
	}

	if logs[0].URL != "/api/endpoint?param=value" {
		t.Errorf("Expected URL /api/endpoint?param=value, got %s", logs[0].URL)
	}
}
