package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for incoming requests
type RateLimiter struct {
	// Per-IP rate limiters
	perIP      map[string]*rate.Limiter
	mu         sync.Mutex
	perIPRate  rate.Limit
	perIPBurst int

	// Global rate limiter
	global *rate.Limiter

	// Cleanup ticker
	cleanup *time.Ticker
}

// NewRateLimiter creates a new rate limiter with per-IP and global limits
// perIPReqPerMin: requests per minute per IP (e.g., 100)
// globalReqPerMin: total requests per minute globally (e.g., 10000)
func NewRateLimiter(perIPReqPerMin, globalReqPerMin int) *RateLimiter {
	rl := &RateLimiter{
		perIP:      make(map[string]*rate.Limiter),
		perIPRate:  rate.Limit(float64(perIPReqPerMin) / 60.0), // Convert to per-second rate
		perIPBurst: perIPReqPerMin / 10,                        // Allow bursts of 10% of per-minute rate
		global:     rate.NewLimiter(rate.Limit(float64(globalReqPerMin)/60.0), globalReqPerMin/10),
		cleanup:    time.NewTicker(5 * time.Minute),
	}

	// Start cleanup goroutine to remove inactive IP limiters
	go rl.cleanupRoutine()

	return rl
}

// cleanupRoutine periodically cleans up inactive IP rate limiters
func (rl *RateLimiter) cleanupRoutine() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		// Remove limiters that haven't been used in the last 5 minutes
		for ip, limiter := range rl.perIP {
			if limiter.Tokens() >= float64(rl.perIPBurst) {
				delete(rl.perIP, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	rl.cleanup.Stop()
}

// getLimiter returns the rate limiter for a specific IP address
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.perIP[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.perIPRate, rl.perIPBurst)
		rl.perIP[ip] = limiter
	}

	return limiter
}

// Middleware returns a middleware function that applies rate limiting
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract IP address from request
			ip := getIPAddress(r)

			// Check global rate limit first
			if !rl.global.Allow() {
				slog.Warn("Global rate limit exceeded",
					"ip", ip,
					"path", r.URL.Path,
				)
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// Check per-IP rate limit
			limiter := rl.getLimiter(ip)
			if !limiter.Allow() {
				slog.Warn("Per-IP rate limit exceeded",
					"ip", ip,
					"path", r.URL.Path,
				)
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getIPAddress extracts the real IP address from the request
// Priority: X-Forwarded-For > X-Real-IP > RemoteAddr
func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := indexOf(xff, ','); idx > 0 {
			return xff[:idx]
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr (format: "IP:port")
	if idx := indexOf(r.RemoteAddr, ':'); idx > 0 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

// indexOf returns the index of the first occurrence of c in s, or -1 if not found
func indexOf(s string, c rune) int {
	for i, ch := range s {
		if ch == c {
			return i
		}
	}
	return -1
}
