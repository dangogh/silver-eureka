package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/dangogh/silver-eureka/internal/database"
)

// Handler handles HTTP requests and logs them to the database
type Handler struct {
	db *database.DB
}

// New creates a new Handler
func New(db *database.DB) *Handler {
	return &Handler{db: db}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handler invoked: ServeHTTP (catch-all)", "method", r.Method, "path", r.URL.Path)
	// Limit request body size to 1MB to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// Extract IP address from request
	ipAddress := getIPAddress(r)

	// Get the full URL
	url := r.URL.String()

	// Debug log for each incoming request
	slog.Debug("Incoming request",
		"method", r.Method,
		"url", url,
		"remote_addr", r.RemoteAddr,
		"ip_address", ipAddress,
		"user_agent", r.UserAgent(),
		"headers", r.Header,
	)

	// Log the request to the database
	if err := h.db.LogRequest(ipAddress, url); err != nil {
		slog.Error("Error logging request to database",
			"error", err,
			"ip_address", ipAddress,
			"url", url,
		)
		// Graceful degradation: return error response but don't crash
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"error":"logging temporarily unavailable","status":"degraded"}`)
		return
	}

	slog.Info("Request logged successfully",
		"ip_address", ipAddress,
		"url", url,
	)

	// Return 404 for all unmatched routes
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "404 page not found\n")
}

// getIPAddress extracts the client IP address from the request
// It checks X-Forwarded-For and X-Real-IP headers first, then falls back to RemoteAddr
func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	// RemoteAddr format is "IP:port", we need just the IP
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}
