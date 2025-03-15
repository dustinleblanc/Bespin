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

	"github.com/dustinleblanc/go-bespin-api/internal/api"
	"github.com/dustinleblanc/go-bespin-api/internal/database"
	"github.com/dustinleblanc/go-bespin-api/internal/queue"
	"github.com/dustinleblanc/go-bespin-api/internal/webhook"
)

func main() {
	// Initialize logger
	logger := log.New(os.Stdout, "[API] ", log.LstdFlags)

	// Get environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// Connect to the database
	db, err := database.NewConnection()
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	// Create webhook repository and service
	webhookRepo := webhook.NewGormRepository(db)
	webhookService := webhook.NewService(webhookRepo)

	// Create job queue
	jobQueue, err := queue.NewAsynqQueue(redisAddr)
	if err != nil {
		logger.Fatalf("Failed to create job queue: %v", err)
	}
	defer jobQueue.Close()

	// Create router
	router := api.NewRouter(jobQueue, webhookService)

	// Create server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Printf("Server is running on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Println("Server exiting")
}
