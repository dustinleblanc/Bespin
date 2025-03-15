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
	"github.com/dustinleblanc/go-bespin/internal/database"
	"github.com/dustinleblanc/go-bespin/internal/jobs"
	"github.com/dustinleblanc/go-bespin/internal/queue"
	"github.com/dustinleblanc/go-bespin/internal/webhook"
	"github.com/dustinleblanc/go-bespin/internal/websocket"
	"github.com/dustinleblanc/go-bespin/pkg/models"
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

	// Create webhook repository
	var webhookRepo webhook.Repository

	// Log database connection parameters
	logger.Printf("Connecting to PostgreSQL at %s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))

	// Connect to PostgreSQL using GORM
	db, err := database.NewGormDB()
	if err != nil {
		logger.Printf("Failed to connect to PostgreSQL: %v", err)
		logger.Fatalf("Could not connect to database. Exiting.")
	}

	// Auto migrate models
	if err := db.AutoMigrate(&models.WebhookReceipt{}); err != nil {
		logger.Printf("Failed to run auto migrations: %v", err)
		logger.Fatalf("Could not migrate database schema. Exiting.")
	}

	// Create GORM repository
	logger.Println("Using PostgreSQL with GORM for webhook storage")
	webhookRepo = webhook.NewGormRepository(db)
	logger.Printf("Using GORM repository: %T", webhookRepo)

	// Create webhook service
	webhookService := webhook.NewService(webhookRepo)

	// Create API handlers
	handlers := api.NewHandlers(jobQueue, webhookService)

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

	// Close database connection
	if db != nil {
		if err := db.Close(); err != nil {
			logger.Printf("Error closing database connection: %v", err)
		}
	}

	// Cancel the context to stop all background goroutines
	cancel()

	logger.Println("Server exiting")
}
