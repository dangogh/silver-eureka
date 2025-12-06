package main

import (
	"context"
	"fmt"
	"log"
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
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
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

	log.Println("Database initialized successfully")

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
		log.Printf("Server starting on port %d", cfg.Port)
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
		log.Printf("Shutdown signal received: %v", sig)

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			// Force close if graceful shutdown fails
			server.Close()
			return fmt.Errorf("could not gracefully shutdown server: %w", err)
		}

		log.Println("Server stopped gracefully")
	}

	return nil
}
