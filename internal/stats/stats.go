package stats

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/dangogh/silver-eureka/internal/database"
)

// Handler handles statistics requests
type Handler struct {
	db *database.DB
}

// New creates a new stats Handler
func New(db *database.DB) *Handler {
	return &Handler{db: db}
}

// HandleEndpointStats returns statistics grouped by endpoint
func (h *Handler) HandleEndpointStats(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Endpoint stats requested",
		"method", r.Method,
		"remote_addr", r.RemoteAddr,
	)

	stats, err := h.db.GetEndpointStats()
	if err != nil {
		slog.Error("Failed to get endpoint stats", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "failed to retrieve endpoint statistics", "details": err.Error()}); encodeErr != nil {
			// Response already started
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		slog.Error("Failed to encode endpoint stats", "error", err)
	}

	slog.Info("Endpoint stats retrieved", "count", len(stats))
}

// HandleSourceStats returns statistics grouped by IP address
func (h *Handler) HandleSourceStats(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Source stats requested",
		"method", r.Method,
		"remote_addr", r.RemoteAddr,
	)

	stats, err := h.db.GetSourceStats()
	if err != nil {
		slog.Error("Failed to get source stats", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "failed to retrieve source statistics", "details": err.Error()}); encodeErr != nil {
			// Response already started
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		slog.Error("Failed to encode source stats", "error", err)
	}

	slog.Info("Source stats retrieved", "count", len(stats))
}

// HandleSummary returns overall statistics
func (h *Handler) HandleSummary(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Summary stats requested",
		"method", r.Method,
		"remote_addr", r.RemoteAddr,
	)

	summary, err := h.db.GetSummary()
	if err != nil {
		slog.Error("Failed to get summary stats", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "failed to retrieve summary statistics", "details": err.Error()}); encodeErr != nil {
			// Response already started
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(summary); err != nil {
		slog.Error("Failed to encode summary stats", "error", err)
	}

	slog.Info("Summary stats retrieved",
		"total_requests", summary.TotalRequests,
		"unique_ips", summary.UniqueIPs,
		"unique_urls", summary.UniqueURLs,
	)
}

// HandleDownload returns all request logs as JSON
func (h *Handler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Download requested",
		"method", r.Method,
		"remote_addr", r.RemoteAddr,
	)

	logs, err := h.db.GetAllLogs()
	if err != nil {
		slog.Error("Failed to get all logs", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "failed to retrieve logs", "details": err.Error()}); encodeErr != nil {
			// Response already started
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=\"request_logs.json\"")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(logs); err != nil {
		slog.Error("Failed to encode logs", "error", err)
	}

	slog.Info("Download completed", "count", len(logs))
}
