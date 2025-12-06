package handler

import (
	"fmt"
	"log"
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
	// Extract IP address from request
	ipAddress := getIPAddress(r)

	// Get the full URL
	url := r.URL.String()

	// Log the request to the database
	if err := h.db.LogRequest(ipAddress, url); err != nil {
		log.Printf("Error logging request: %v", err)
		// Continue serving the request even if logging fails
	}

	// Respond with a simple message
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Request logged: %s from %s\n", url, ipAddress)
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
