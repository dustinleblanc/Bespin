package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dustinleblanc/go-bespin-worker/internal/jobs"
	"github.com/dustinleblanc/go-bespin-worker/pkg/tasks"
	"github.com/hibiken/asynq"
)

func main() {
	// Get Redis address from environment variable
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// Create Redis connection for Asynq
	redisOpt := asynq.RedisClientOpt{
		Addr: redisAddr,
	}

	// Create and configure the Asynq server
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally specify multiple queues with different priorities
			Queues: map[string]int{
				"critical": 6, // processed 60% of the time
				"default":  3, // processed 30% of the time
				"low":      1, // processed 10% of the time
			},
		},
	)

	// Create a new processor
	processor := jobs.NewProcessor()

	// Configure the mux server to handle different task types
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeRandomText, processor.HandleRandomTextTask)
	mux.HandleFunc(tasks.TypeWebhook, processor.HandleWebhookTask)

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down...", sig)
		srv.Stop()
	}()

	// Start the server
	log.Printf("Starting worker server at %s", redisAddr)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}

	fmt.Println("Worker server stopped")
}
