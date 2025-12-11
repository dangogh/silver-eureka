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
	defer db.Close()

	slog.Info("Database initialized successfully", "database", cfg.DBPath)

	// Create HTTP router with all endpoints
	h := router.New(db)

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
			server.Close()
			return fmt.Errorf("could not gracefully shutdown server: %w", err)
		}

		slog.Info("Server stopped gracefully")
	}

	return nil
}
