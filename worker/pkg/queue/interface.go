package queue

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// JobHandler is a function that processes a job
type JobHandler func(ctx context.Context, jobID string, data interface{}) error

// JobQueueInterface defines the interface for job queues
type JobQueueInterface interface {
	// AddJob adds a job to the queue
	AddJob(jobType string, data interface{}) (string, error)

	// GetJob gets a job from the queue
	GetJob(ctx context.Context, jobType string) (string, interface{}, error)

	// GetJobResult gets the result of a job
	GetJobResult(ctx context.Context, jobID string) (interface{}, error)

	// GetRedisClient gets the Redis client
	GetRedisClient() *redis.Client

	// StartProcessing starts processing jobs of the given type
	StartProcessing(ctx context.Context, jobType string, handler JobHandler) error
}

// NewJobQueue creates a new job queue
func NewJobQueue(redisAddr string) (JobQueueInterface, error) {
	// This is just a placeholder that returns nil
	// The actual implementation is in the internal package
	return nil, nil
}
