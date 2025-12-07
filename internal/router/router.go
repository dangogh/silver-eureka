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

	// Stats endpoints
	statsHandler := stats.New(db)
	mux.HandleFunc("/stats/endpoints", statsHandler.HandleEndpointStats)
	mux.HandleFunc("/stats/sources", statsHandler.HandleSourceStats)
	mux.HandleFunc("/stats/summary", statsHandler.HandleSummary)

	// Default handler for all other requests (logs them)
	logHandler := handler.New(db)
	mux.Handle("/", logHandler)

	return mux
}
