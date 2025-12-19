package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/time/rate"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(100, 10000)
	defer rl.Stop()

	if rl == nil {
		t.Fatal("Expected rate limiter to be created")
	}

	if rl.perIPRate != rate.Limit(100.0/60.0) {
		t.Errorf("Expected per-IP rate of %v, got %v", 100.0/60.0, rl.perIPRate)
	}

	if rl.perIPBurst != 10 {
		t.Errorf("Expected per-IP burst of 10, got %d", rl.perIPBurst)
	}
}

func TestRateLimiter_PerIPLimit(t *testing.T) {
	// Create a rate limiter: 10 requests per minute per IP (burst of 1)
	rl := NewRateLimiter(10, 10000)
	defer rl.Stop()

	handler := rl.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request should succeed
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Exhaust the burst (10 requests per minute = ~1 per 6 seconds, burst of 1)
	// Make rapid requests to trigger rate limit
	for i := 0; i < 5; i++ {
		req = httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// This request should be rate limited
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d after burst", w.Code)
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(10, 10000)
	defer rl.Stop()

	handler := rl.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First IP should succeed
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("Expected status 200 for IP 1, got %d", w1.Code)
	}

	// Second IP should also succeed (separate rate limit)
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status 200 for IP 2, got %d", w2.Code)
	}
}

func TestRateLimiter_GlobalLimit(t *testing.T) {
	// Create a rate limiter with very low global limit: 60 requests per minute
	// This gives burst of 6 (60/10), which should allow first few requests
	rl := NewRateLimiter(1000, 60)
	defer rl.Stop()

	handler := rl.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make many requests from different IPs to trigger global limit
	successCount := 0
	rateLimitedCount := 0

	for i := 0; i < 20; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		// Create proper IP address strings
		req.RemoteAddr = "192.168.1." + string(rune('0'+i)) + ":12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			successCount++
		} else if w.Code == http.StatusTooManyRequests {
			rateLimitedCount++
		}
	}

	// With burst of 6, we should get at least a few successful requests
	if successCount < 1 {
		t.Errorf("Expected at least 1 successful request, got %d", successCount)
	}

	// And we should have rate limited many requests (20 - burst)
	if rateLimitedCount < 1 {
		t.Errorf("Expected at least 1 request to be rate limited by global limit, got %d", rateLimitedCount)
	}

	t.Logf("Success: %d, Rate limited: %d (expected ~6 success, ~14 limited)", successCount, rateLimitedCount)
}

func TestRateLimiter_XForwardedFor(t *testing.T) {
	rl := NewRateLimiter(10, 10000)
	defer rl.Stop()

	handler := rl.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request with X-Forwarded-For header
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Same X-Forwarded-For IP should share rate limit
	for i := 0; i < 5; i++ {
		req = httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.2:12345" // Different RemoteAddr
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.2")
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// Should be rate limited (same first IP in X-Forwarded-For)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}
}

func TestRateLimiter_XRealIP(t *testing.T) {
	rl := NewRateLimiter(10, 10000)
	defer rl.Stop()

	handler := rl.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request with X-Real-IP header
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Real-IP", "203.0.113.1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRateLimiter_CleanupRoutine(t *testing.T) {
	rl := NewRateLimiter(100, 10000)
	defer rl.Stop()

	// Add some limiters
	limiter1 := rl.getLimiter("192.168.1.1")
	limiter2 := rl.getLimiter("192.168.1.2")

	if limiter1 == nil || limiter2 == nil {
		t.Fatal("Failed to create limiters")
	}

	// Verify limiters exist
	rl.mu.Lock()
	count := len(rl.perIP)
	rl.mu.Unlock()

	if count != 2 {
		t.Errorf("Expected 2 limiters, got %d", count)
	}

	// Cleanup routine should run, but we won't wait for it in tests
	// Just verify the structure is correct
}

func TestGetIPAddress(t *testing.T) {
	tests := []struct {
		name          string
		remoteAddr    string
		xForwardedFor string
		xRealIP       string
		expectedIP    string
	}{
		{
			name:       "RemoteAddr only",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:          "X-Forwarded-For takes priority",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:          "X-Forwarded-For with multiple IPs",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1, 198.51.100.1, 192.0.2.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "10.0.0.1:12345",
			xRealIP:    "203.0.113.1",
			expectedIP: "203.0.113.1",
		},
		{
			name:          "X-Forwarded-For over X-Real-IP",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			xRealIP:       "198.51.100.1",
			expectedIP:    "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := getIPAddress(req)
			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		c        rune
		expected int
	}{
		{"Found at start", "hello", 'h', 0},
		{"Found in middle", "hello", 'l', 2},
		{"Found at end", "hello", 'o', 4},
		{"Not found", "hello", 'x', -1},
		{"Empty string", "", 'a', -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indexOf(tt.s, tt.c)
			if result != tt.expected {
				t.Errorf("indexOf(%q, %q) = %d, expected %d", tt.s, tt.c, result, tt.expected)
			}
		})
	}
}

func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter(100, 10000)

	// Stop should not panic
	rl.Stop()

	// Calling Stop again should not panic
	rl.Stop()
}
