package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"authservice/internal/api"
	"authservice/internal/auth"
	"authservice/internal/config"
	"authservice/internal/database"
	"authservice/internal/repository"
)

func main() {
	log.Println("Starting auth service...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	pool, err := database.NewPostgresPool(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// --- Dependency Injection ---
	userRepo := repository.NewPostgresUserRepository(pool)
	authSvc := auth.NewAuthService(userRepo, cfg) // Pass cfg here
	authHandler := api.NewAuthHandler(authSvc)

	// Setup router
	router := api.NewRouter(authHandler)

	// Setup HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
		// Add timeouts for production hardening
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// --- Graceful Shutdown ---
	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Channel to listen for server errors
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		log.Printf("Auth service listening on :%s", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)
	case sig := <-stop:
		log.Printf("Received signal: %s. Starting graceful shutdown...", sig)
	}

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Graceful shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}
