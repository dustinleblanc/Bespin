package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dustinleblanc/go-bespin/internal/api"
	"github.com/dustinleblanc/go-bespin/internal/jobs"
	"github.com/dustinleblanc/go-bespin/internal/queue"
	"github.com/dustinleblanc/go-bespin/internal/websocket"
)

func main() {
	// Create logger
	logger := log.New(os.Stdout, "[Main] ", log.LstdFlags)
	logger.Println("Starting Bespin API server")

	// Get Redis address from environment or use default
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}

	// Create job queue
	jobQueue := queue.NewJobQueue(redisAddr)

	// Create job processor
	processor := jobs.NewProcessor(jobQueue)

	// Create WebSocket server
	wsServer := websocket.NewServer(jobQueue)

	// Create API handlers
	handlers := api.NewHandlers(jobQueue)

	// Set up router
	router := api.SetupRouter(handlers, wsServer)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Create context that listens for signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start job processor
	processor.Start(ctx)

	// Start WebSocket server
	wsServer.Start(ctx)

	// Start HTTP server in a goroutine
	go func() {
		logger.Printf("HTTP server listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down server...")

	// Create a deadline for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	// Cancel the context to stop all background goroutines
	cancel()

	logger.Println("Server exiting")
}
