package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dustinleblanc/go-bespin-worker/pkg/models"
	"github.com/dustinleblanc/go-bespin-worker/pkg/queue"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// JobQueueInterface defines the interface for job queue operations
type JobQueueInterface interface {
	AddJob(jobType string, data interface{}) (string, error)
	GetJob(ctx context.Context, jobType string) (string, interface{}, error)
	GetJobResult(ctx context.Context, jobID string) (interface{}, error)
	GetRedisClient() *redis.Client
	StartProcessing(ctx context.Context, jobType string, handler queue.JobHandler) error
}

// jobQueue handles job queue operations
type jobQueue struct {
	client *redis.Client
}

// NewJobQueue creates a new job queue
func NewJobQueue(redisAddr string) (queue.JobQueueInterface, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &jobQueue{
		client: client,
	}, nil
}

// AddJob adds a job to the queue
func (q *jobQueue) AddJob(jobType string, data interface{}) (string, error) {
	job := &models.Job{
		ID:        uuid.New().String(),
		Type:      jobType,
		Data:      data,
		Status:    models.JobStatusPending,
		CreatedAt: time.Now(),
	}

	// Convert job to JSON
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job: %w", err)
	}

	// Add job to queue
	ctx := context.Background()
	if err := q.client.RPush(ctx, fmt.Sprintf("queue:%s", jobType), jobBytes).Err(); err != nil {
		return "", fmt.Errorf("failed to add job to queue: %w", err)
	}

	return job.ID, nil
}

// GetJob gets a job from the queue
func (q *jobQueue) GetJob(ctx context.Context, jobType string) (string, interface{}, error) {
	// Get job from queue
	jobBytes, err := q.client.LPop(ctx, fmt.Sprintf("queue:%s", jobType)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return "", nil, nil
		}
		return "", nil, fmt.Errorf("failed to get job from queue: %w", err)
	}

	// Parse job
	var job models.Job
	if err := json.Unmarshal(jobBytes, &job); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return job.ID, job.Data, nil
}

// GetJobResult gets a job result by ID
func (q *jobQueue) GetJobResult(ctx context.Context, jobID string) (interface{}, error) {
	// Get job result from Redis
	resultBytes, err := q.client.Get(ctx, fmt.Sprintf("result:%s", jobID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get job result: %w", err)
	}

	// Parse result
	var result models.JobResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job result: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf(result.Error)
	}

	return result.Data, nil
}

// GetRedisClient returns the Redis client
func (q *jobQueue) GetRedisClient() *redis.Client {
	return q.client
}

// StartProcessing starts processing jobs of the given type
func (q *jobQueue) StartProcessing(ctx context.Context, jobType string, handler queue.JobHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Get job from queue
			jobID, data, err := q.GetJob(ctx, jobType)
			if err != nil {
				return fmt.Errorf("failed to get job: %w", err)
			}

			if jobID == "" {
				// No jobs available, wait a bit
				time.Sleep(time.Second)
				continue
			}

			// Process job
			result := handler(ctx, jobID, data)

			// Store result
			jobResult := &models.JobResult{
				JobID:     jobID,
				Data:      result,
				Error:     "",
				CreatedAt: time.Now(),
			}

			// Convert result to JSON
			resultBytes, err := json.Marshal(jobResult)
			if err != nil {
				return fmt.Errorf("failed to marshal job result: %w", err)
			}

			// Store result in Redis
			if err := q.client.Set(ctx, fmt.Sprintf("result:%s", jobID), resultBytes, 24*time.Hour).Err(); err != nil {
				return fmt.Errorf("failed to store job result: %w", err)
			}
		}
	}
}
