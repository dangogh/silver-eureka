package router

import (
	"log/slog"
	"net/http"

	"github.com/dangogh/silver-eureka/internal/database"
	"github.com/dangogh/silver-eureka/internal/handler"
	"github.com/dangogh/silver-eureka/internal/middleware"
	"github.com/dangogh/silver-eureka/internal/stats"
	"github.com/dangogh/silver-eureka/internal/web"
)

// New creates a new HTTP router with all application routes
func New(db *database.DB, authUsername, authPassword string) http.Handler {
	return NewWithRateLimiter(db, authUsername, authPassword, true)
}

// NewWithRateLimiter creates a new HTTP router with optional rate limiting
func NewWithRateLimiter(db *database.DB, authUsername, authPassword string, enableRateLimit bool) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint (public, no auth)
	mux.HandleFunc("/health", handleHealth(db))

	// Web interface routes (session-based auth)
	if authUsername != "" && authPassword != "" {
		webHandler := web.NewHandler(db, authUsername, authPassword)
		mux.HandleFunc("GET /login", webHandler.HandleLoginPage)
		mux.HandleFunc("POST /login", webHandler.HandleLoginSubmit)
		mux.HandleFunc("POST /logout", webHandler.RequireAuth(webHandler.HandleLogout))
		mux.HandleFunc("GET /dashboard", webHandler.RequireAuth(webHandler.HandleDashboard))
		mux.HandleFunc("GET /stats-view/{type}", webHandler.RequireAuth(webHandler.HandleStatsView))
	}

	// API stats endpoints (protected with basic auth if configured)
	authMiddleware := middleware.BasicAuth(authUsername, authPassword)
	statsHandler := stats.New(db)
	mux.Handle("/stats/endpoints", authMiddleware(http.HandlerFunc(statsHandler.HandleEndpointStats)))
	mux.Handle("/stats/sources", authMiddleware(http.HandlerFunc(statsHandler.HandleSourceStats)))
	mux.Handle("/stats/summary", authMiddleware(http.HandlerFunc(statsHandler.HandleSummary)))
	mux.Handle("/stats/download", authMiddleware(http.HandlerFunc(statsHandler.HandleDownload)))

	// Default handler for all other requests (logs them, returns 404)
	logHandler := handler.New(db)
	mux.Handle("/", logHandler)

	// Apply rate limiting to all routes if enabled
	if enableRateLimit {
		// Initialize rate limiter: 100 req/min per IP, 10,000 req/min global
		rateLimiter := middleware.NewRateLimiter(100, 10000)
		return rateLimiter.Middleware()(mux)
	}

	return mux
}

// handleHealth returns a health check handler
func handleHealth(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("Handler invoked: handleHealth", "method", r.Method, "path", r.URL.Path)
		// Check database connectivity
		if err := db.Ping(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			if _, writeErr := w.Write([]byte(`{"status":"unhealthy","database":"down"}`)); writeErr != nil {
				// Response already started
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"healthy","database":"up"}`)); err != nil {
			// Response already started
		}
	}
}
