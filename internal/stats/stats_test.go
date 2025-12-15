package stats

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dangogh/silver-eureka/internal/database"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	dbPath := "/tmp/test_stats.db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in cleanup
		}
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in cleanup
		}
	}

	return db, cleanup
}

func TestHandleEndpointStats(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Add some test data
	if err := db.LogRequest("192.168.1.1", "/test/path1"); err != nil {
		t.Fatalf("Failed to log request: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := db.LogRequest("192.168.1.2", "/test/path1"); err != nil {
		t.Fatalf("Failed to log request: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := db.LogRequest("192.168.1.1", "/test/path2"); err != nil {
		t.Fatalf("Failed to log request: %v", err)
	}

	handler := New(db)

	req := httptest.NewRequest(http.MethodGet, "/stats/endpoints", nil)
	w := httptest.NewRecorder()

	handler.HandleEndpointStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}

	var stats []database.EndpointStats
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf("Expected 2 endpoint stats, got %d", len(stats))
	}

	// First endpoint should be /test/path1 with 2 requests
	if stats[0].URL != "/test/path1" {
		t.Errorf("Expected first URL to be /test/path1, got %s", stats[0].URL)
	}
	if stats[0].Count != 2 {
		t.Errorf("Expected count 2 for /test/path1, got %d", stats[0].Count)
	}
	if stats[0].UniqueIPs != 2 {
		t.Errorf("Expected 2 unique IPs for /test/path1, got %d", stats[0].UniqueIPs)
	}
}

func TestHandleSourceStats(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Add some test data
	if err := db.LogRequest("192.168.1.1", "/test/path1"); err != nil {
		t.Fatalf("Failed to log request: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := db.LogRequest("192.168.1.1", "/test/path2"); err != nil {
		t.Fatalf("Failed to log request: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := db.LogRequest("192.168.1.2", "/test/path1"); err != nil {
		t.Fatalf("Failed to log request: %v", err)
	}

	handler := New(db)

	req := httptest.NewRequest(http.MethodGet, "/stats/sources", nil)
	w := httptest.NewRecorder()

	handler.HandleSourceStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}

	var stats []database.SourceStats
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf("Expected 2 source stats, got %d", len(stats))
	}

	// First source should be 192.168.1.1 with 2 requests
	if stats[0].IPAddress != "192.168.1.1" {
		t.Errorf("Expected first IP to be 192.168.1.1, got %s", stats[0].IPAddress)
	}
	if stats[0].Count != 2 {
		t.Errorf("Expected count 2 for 192.168.1.1, got %d", stats[0].Count)
	}
	if stats[0].UniqueURLs != 2 {
		t.Errorf("Expected 2 unique URLs for 192.168.1.1, got %d", stats[0].UniqueURLs)
	}
}

func TestHandleSummary(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Add some test data
	if err := db.LogRequest("192.168.1.1", "/test/path1"); err != nil {
		t.Fatalf("Failed to log request: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := db.LogRequest("192.168.1.2", "/test/path2"); err != nil {
		t.Fatalf("Failed to log request: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := db.LogRequest("192.168.1.1", "/test/path1"); err != nil {
		t.Fatalf("Failed to log request: %v", err)
	}

	handler := New(db)

	req := httptest.NewRequest(http.MethodGet, "/stats/summary", nil)
	w := httptest.NewRecorder()

	handler.HandleSummary(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}

	var summary database.Summary
	if err := json.NewDecoder(w.Body).Decode(&summary); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if summary.TotalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", summary.TotalRequests)
	}
	if summary.UniqueIPs != 2 {
		t.Errorf("Expected 2 unique IPs, got %d", summary.UniqueIPs)
	}
	if summary.UniqueURLs != 2 {
		t.Errorf("Expected 2 unique URLs, got %d", summary.UniqueURLs)
	}
}
