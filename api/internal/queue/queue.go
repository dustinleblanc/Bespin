package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dustinleblanc/go-bespin-api/pkg/models"
	"github.com/hibiken/asynq"
)

// Queue represents a job queue
type Queue interface {
	// AddJob adds a job to the queue
	AddJob(ctx context.Context, job *models.Job) (string, error)
	// GetJobResult gets a job result
	GetJobResult(ctx context.Context, jobID string) (*models.JobResult, error)
}

// AsynqQueue implements Queue using Asynq
type AsynqQueue struct {
	client    *asynq.Client
	inspector *asynq.Inspector
}

// NewAsynqQueue creates a new AsynqQueue
func NewAsynqQueue(redisAddr string) (*AsynqQueue, error) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: redisAddr})

	return &AsynqQueue{
		client:    client,
		inspector: inspector,
	}, nil
}

// AddJob adds a job to the queue
func (q *AsynqQueue) AddJob(ctx context.Context, job *models.Job) (string, error) {
	// Serialize the job data
	payload, err := json.Marshal(job.Data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job data: %w", err)
	}

	// Create the task
	task := asynq.NewTask(string(job.Type), payload)

	// Enqueue the task
	info, err := q.client.EnqueueContext(ctx, task)
	if err != nil {
		return "", fmt.Errorf("failed to enqueue task: %w", err)
	}

	return info.ID, nil
}

// GetJobResult gets a job result
func (q *AsynqQueue) GetJobResult(ctx context.Context, jobID string) (*models.JobResult, error) {
	// Get the task info
	info, err := q.inspector.GetTaskInfo("default", jobID)
	if err != nil {
		if err == asynq.ErrTaskNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task info: %w", err)
	}

	// Create the job result
	result := &models.JobResult{
		ID:        jobID,
		Status:    models.JobStatusPending,
		CreatedAt: time.Now(), // Asynq doesn't expose task creation time
	}

	// Update the status based on the task state
	switch info.State.String() {
	case "active":
		result.Status = models.JobStatusProcessing
	case "completed":
		result.Status = models.JobStatusCompleted
		result.CompletedAt = &info.CompletedAt
		result.Result = string(info.Result)
	case "failed":
		result.Status = models.JobStatusFailed
		result.Error = info.LastErr
	case "retry":
		result.Status = models.JobStatusRetrying
		result.Error = info.LastErr
	}

	return result, nil
}

// Close closes the queue
func (q *AsynqQueue) Close() error {
	if err := q.client.Close(); err != nil {
		return fmt.Errorf("failed to close client: %w", err)
	}
	return nil
}
