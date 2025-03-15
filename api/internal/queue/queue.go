package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dustinleblanc/go-bespin/pkg/models"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// JobQueue handles job queue operations
type JobQueue struct {
	redisClient *redis.Client
	logger      *log.Logger
}

// JobHandler is a function that processes a job
type JobHandler func(job *models.Job) (interface{}, error)

// NewJobQueue creates a new job queue
func NewJobQueue(redisAddr string) *JobQueue {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	logger := log.New(log.Writer(), "[JobQueue] ", log.LstdFlags)

	return &JobQueue{
		redisClient: client,
		logger:      logger,
	}
}

// AddJob adds a job to the queue
func (q *JobQueue) AddJob(jobType string, data interface{}) (string, error) {
	ctx := context.Background()
	jobID := uuid.New().String()

	job := &models.Job{
		ID:        jobID,
		Type:      jobType,
		Data:      data,
		CreatedAt: time.Now(),
		Status:    models.JobStatusQueued,
	}

	jobJSON, err := json.Marshal(job)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job: %w", err)
	}

	// Store job data
	err = q.redisClient.Set(ctx, fmt.Sprintf("job:%s", jobID), jobJSON, 0).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store job: %w", err)
	}

	// Add job to queue
	err = q.redisClient.LPush(ctx, fmt.Sprintf("queue:%s", jobType), jobID).Err()
	if err != nil {
		return "", fmt.Errorf("failed to add job to queue: %w", err)
	}

	q.logger.Printf("Added job %s of type %s to queue", jobID, jobType)
	return jobID, nil
}

// StartProcessing starts processing jobs of the given type
func (q *JobQueue) StartProcessing(ctx context.Context, jobType string, handler JobHandler) {
	q.logger.Printf("Starting job processor for type: %s", jobType)

	go func() {
		for {
			select {
			case <-ctx.Done():
				q.logger.Printf("Stopping job processor for type: %s", jobType)
				return
			default:
				// Try to get a job from the queue
				result, err := q.redisClient.BRPop(ctx, 5*time.Second, fmt.Sprintf("queue:%s", jobType)).Result()
				if err != nil {
					if err != redis.Nil {
						q.logger.Printf("Error getting job from queue: %v", err)
					}
					continue
				}

				if len(result) < 2 {
					continue
				}

				jobID := result[1]
				q.processJob(ctx, jobID, handler)
			}
		}
	}()
}

// processJob processes a job
func (q *JobQueue) processJob(ctx context.Context, jobID string, handler JobHandler) {
	q.logger.Printf("Processing job: %s", jobID)

	// Get job data
	jobJSON, err := q.redisClient.Get(ctx, fmt.Sprintf("job:%s", jobID)).Result()
	if err != nil {
		q.logger.Printf("Error getting job data: %v", err)
		return
	}

	var job models.Job
	if err := json.Unmarshal([]byte(jobJSON), &job); err != nil {
		q.logger.Printf("Error unmarshaling job data: %v", err)
		return
	}

	// Update job status to processing
	job.Status = models.JobStatusProcessing
	updatedJobJSON, _ := json.Marshal(job)
	q.redisClient.Set(ctx, fmt.Sprintf("job:%s", jobID), updatedJobJSON, 0)

	// Process the job
	result, err := handler(&job)

	jobResult := models.JobResult{
		JobID:       jobID,
		CompletedAt: time.Now(),
	}

	if err != nil {
		q.logger.Printf("Error processing job %s: %v", jobID, err)
		// Update job status to failed
		job.Status = models.JobStatusFailed
		updatedJobJSON, _ := json.Marshal(job)
		q.redisClient.Set(ctx, fmt.Sprintf("job:%s", jobID), updatedJobJSON, 0)

		// Store error
		jobResult.Error = err.Error()
	} else {
		// Update job status to completed
		job.Status = models.JobStatusCompleted
		updatedJobJSON, _ := json.Marshal(job)
		q.redisClient.Set(ctx, fmt.Sprintf("job:%s", jobID), updatedJobJSON, 0)

		// Store result
		jobResult.Result = result
	}

	// Store job result
	resultJSON, _ := json.Marshal(jobResult)
	q.redisClient.Set(ctx, fmt.Sprintf("job:%s:result", jobID), resultJSON, 0)

	// Publish completion event
	q.redisClient.Publish(ctx, fmt.Sprintf("job-completed:%s", jobID), resultJSON)

	q.logger.Printf("Completed job: %s", jobID)
}

// GetJob gets a job by ID
func (q *JobQueue) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	jobJSON, err := q.redisClient.Get(ctx, fmt.Sprintf("job:%s", jobID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	var job models.Job
	if err := json.Unmarshal([]byte(jobJSON), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// GetJobResult gets a job result by ID
func (q *JobQueue) GetJobResult(ctx context.Context, jobID string) (*models.JobResult, error) {
	resultJSON, err := q.redisClient.Get(ctx, fmt.Sprintf("job:%s:result", jobID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get job result: %w", err)
	}

	var result models.JobResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job result: %w", err)
	}

	return &result, nil
}

// GetRedisClient returns the Redis client
func (q *JobQueue) GetRedisClient() *redis.Client {
	return q.redisClient
}
