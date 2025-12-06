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

// createRedirectHandler creates an HTTP handler that redirects all requests to HTTPS
func createRedirectHandler(httpsPort int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the redirect
		slog.Debug("HTTP to HTTPS redirect",
			"method", r.Method,
			"url", r.URL.String(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)

		// Construct the HTTPS URL
		host := r.Host
		// If host doesn't include a port, or includes the HTTP port, update it
		if colonIdx := len(host) - 1; colonIdx > 0 {
			for i := len(host) - 1; i >= 0; i-- {
				if host[i] == ':' {
					host = host[:i]
					break
				}
			}
		}

		var httpsURL string
		if httpsPort == 443 {
			httpsURL = fmt.Sprintf("https://%s%s", host, r.RequestURI)
		} else {
			httpsURL = fmt.Sprintf("https://%s:%d%s", host, httpsPort, r.RequestURI)
		}

		slog.Info("Redirecting HTTP to HTTPS",
			"from", fmt.Sprintf("http://%s%s", r.Host, r.RequestURI),
			"to", httpsURL,
		)

		http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
	})
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

	// HTTP redirect server (if TLS is enabled and redirect is enabled)
	var httpRedirectServer *http.Server

	// Start the server in a goroutine
	go func() {
		if cfg.TLSEnabled {
			slog.Info("HTTPS server starting",
				"port", cfg.Port,
				"tls_cert", cfg.TLSCert,
				"tls_key", cfg.TLSKey,
			)

			// Start HTTP redirect server if enabled
			if cfg.HTTPRedirect {
				httpRedirectServer = &http.Server{
					Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
					Handler:      createRedirectHandler(cfg.Port),
					ReadTimeout:  15 * time.Second,
					WriteTimeout: 15 * time.Second,
					IdleTimeout:  60 * time.Second,
				}

				go func() {
					slog.Info("HTTP redirect server starting",
						"http_port", cfg.HTTPPort,
						"https_port", cfg.Port,
					)
					if err := httpRedirectServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						slog.Error("HTTP redirect server error", "error", err)
					}
				}()
			}

			serverErrors <- server.ListenAndServeTLS(cfg.TLSCert, cfg.TLSKey)
		} else {
			slog.Info("HTTP server starting", "port", cfg.Port, "tls", false)
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

		// Shutdown HTTP redirect server first if it exists
		if httpRedirectServer != nil {
			if err := httpRedirectServer.Shutdown(ctx); err != nil {
				slog.Warn("HTTP redirect server shutdown error", "error", err)
				httpRedirectServer.Close()
			} else {
				slog.Info("HTTP redirect server stopped gracefully")
			}
		}

		// Attempt graceful shutdown of main server
		if err := server.Shutdown(ctx); err != nil {
			// Force close if graceful shutdown fails
			server.Close()
			return fmt.Errorf("could not gracefully shutdown server: %w", err)
		}

		slog.Info("Server stopped gracefully")
	}

	return nil
}
