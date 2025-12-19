package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitingIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

	// Create router with rate limiting enabled
	router := NewWithRateLimiter(db, "", "", true)

	// Make multiple rapid requests from same IP
	successCount := 0
	rateLimitedCount := 0

	for i := 0; i < 15; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
		req.RemoteAddr = "192.0.2.1:12345"
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code == http.StatusNotFound {
			// Request was processed (logged) but path doesn't exist
			successCount++
		} else if rec.Code == http.StatusTooManyRequests {
			rateLimitedCount++
		}
	}

	// Should have some successful and some rate limited
	if successCount == 0 {
		t.Error("Expected at least some successful requests")
	}

	if rateLimitedCount == 0 {
		t.Error("Expected at least some requests to be rate limited")
	}

	t.Logf("Successful: %d, Rate limited: %d", successCount, rateLimitedCount)
}

func TestRateLimitingDisabled(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

	// Create router with rate limiting disabled
	router := NewWithRateLimiter(db, "", "", false)

	// Make many rapid requests - none should be rate limited
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
		req.RemoteAddr = "192.0.2.1:12345"
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code == http.StatusTooManyRequests {
			t.Errorf("Request %d was rate limited when rate limiting is disabled", i)
		}
	}
}
