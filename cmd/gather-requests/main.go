package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dangogh/silver-eureka/internal/config"
	"github.com/dangogh/silver-eureka/internal/database"
	"github.com/dangogh/silver-eureka/internal/router"
)

func main() {
	// Initialize structured JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("Application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg := config.Load()

	// Ensure database directory exists
	dbDir := cfg.DBPath
	if idx := strings.LastIndex(dbDir, "/"); idx > 0 {
		dbDir = dbDir[:idx]
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Initialize database
	db, err := database.New(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database", "error", err)
		}
	}()

	slog.Info("Database initialized successfully", "database", cfg.DBPath)

	// Log auth status
	if cfg.AuthUsername != "" && cfg.AuthPassword != "" {
		slog.Info("HTTP Basic Auth enabled for /stats/* endpoints")
	} else {
		slog.Warn("HTTP Basic Auth not configured - stats endpoints are public")
	}

	// Log retention status
	if cfg.LogRetentionDays > 0 {
		slog.Info("Log retention enabled", "retention_days", cfg.LogRetentionDays)
	} else {
		slog.Info("Log retention disabled - logs will be kept indefinitely")
	}

	// Create HTTP router with all endpoints
	h := router.New(db, cfg.AuthUsername, cfg.AuthPassword)

	// Create HTTP server with concurrency-friendly settings
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           h,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the HTTP server in a goroutine
	go func() {
		slog.Info("HTTP server starting", "port", cfg.Port)
		serverErrors <- server.ListenAndServe()
	}()

	// Start background log cleanup goroutine if retention is enabled
	if cfg.LogRetentionDays > 0 {
		go func() {
			// Run cleanup immediately on startup
			if deleted, err := db.CleanupOldLogs(cfg.LogRetentionDays); err != nil {
				slog.Error("Failed to cleanup old logs on startup", "error", err)
			} else if deleted > 0 {
				slog.Info("Cleaned up old logs on startup", "deleted", deleted)
			}

			// Then run daily
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()

			for range ticker.C {
				deleted, err := db.CleanupOldLogs(cfg.LogRetentionDays)
				if err != nil {
					slog.Error("Failed to cleanup old logs", "error", err)
				} else if deleted > 0 {
					slog.Info("Cleaned up old logs", "deleted", deleted, "retention_days", cfg.LogRetentionDays)
				} else {
					slog.Debug("Log cleanup ran, no old logs found")
				}
			}
		}()
	}

	// Channel to listen for interrupt or terminate signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		slog.Info("Shutdown signal received", "signal", sig.String())

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			// Force close if graceful shutdown fails
			if closeErr := server.Close(); closeErr != nil {
				slog.Error("Failed to force close server", "error", closeErr)
			}
			return fmt.Errorf("could not gracefully shutdown server: %w", err)
		}

		slog.Info("Server stopped gracefully")
	}

	return nil
}
