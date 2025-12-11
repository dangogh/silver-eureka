package router

import (
	"net/http"

	"github.com/dangogh/silver-eureka/internal/database"
	"github.com/dangogh/silver-eureka/internal/handler"
	"github.com/dangogh/silver-eureka/internal/stats"
)

// New creates a new HTTP router with all application routes
func New(db *database.DB) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", handleHealth(db))

	// Stats endpoints
	statsHandler := stats.New(db)
	mux.HandleFunc("/stats/endpoints", statsHandler.HandleEndpointStats)
	mux.HandleFunc("/stats/sources", statsHandler.HandleSourceStats)
	mux.HandleFunc("/stats/summary", statsHandler.HandleSummary)
	mux.HandleFunc("/stats/download", statsHandler.HandleDownload)

	// Default handler for all other requests (logs them)
	logHandler := handler.New(db)
	mux.Handle("/", logHandler)

	return mux
}

// handleHealth returns a health check handler
func handleHealth(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check database connectivity
		if err := db.Ping(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","database":"down"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","database":"up"}`))
	}
}
