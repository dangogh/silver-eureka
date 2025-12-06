package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dangogh/silver-eureka/internal/config"
	"github.com/dangogh/silver-eureka/internal/database"
	"github.com/dangogh/silver-eureka/internal/handler"
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

	// Initialize database
	db, err := database.New("requests.db")
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	slog.Info("Database initialized successfully", "database", "requests.db")

	// Create HTTP handler
	h := handler.New(db)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		if cfg.TLSEnabled {
			slog.Info("Server starting with TLS",
				"port", cfg.Port,
				"tls_cert", cfg.TLSCert,
				"tls_key", cfg.TLSKey,
			)
			serverErrors <- server.ListenAndServeTLS(cfg.TLSCert, cfg.TLSKey)
		} else {
			slog.Info("Server starting", "port", cfg.Port, "tls", false)
			serverErrors <- server.ListenAndServe()
		}
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
