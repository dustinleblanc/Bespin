package queue

import (
	"context"

	"github.com/dustinleblanc/go-bespin/pkg/models"
	"github.com/go-redis/redis/v8"
)

// MockJobQueue is a mock implementation of the job queue for testing
type MockJobQueue struct {
	AddJobFunc       func(jobType string, data interface{}) (string, error)
	GetJobFunc       func(ctx context.Context, jobID string) (*models.Job, error)
	GetJobResultFunc func(ctx context.Context, jobID string) (*models.JobResult, error)
	RedisClient      *redis.Client
}

// AddJob implements the JobQueue interface
func (m *MockJobQueue) AddJob(jobType string, data interface{}) (string, error) {
	if m.AddJobFunc != nil {
		return m.AddJobFunc(jobType, data)
	}
	return "mock-job-id", nil
}

// GetJob implements the JobQueue interface
func (m *MockJobQueue) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	if m.GetJobFunc != nil {
		return m.GetJobFunc(ctx, jobID)
	}
	return &models.Job{ID: jobID}, nil
}

// GetJobResult implements the JobQueue interface
func (m *MockJobQueue) GetJobResult(ctx context.Context, jobID string) (*models.JobResult, error) {
	if m.GetJobResultFunc != nil {
		return m.GetJobResultFunc(ctx, jobID)
	}
	return &models.JobResult{JobID: jobID}, nil
}

// GetRedisClient implements the JobQueue interface
func (m *MockJobQueue) GetRedisClient() *redis.Client {
	return m.RedisClient
}

// StartProcessing implements the JobQueue interface
func (m *MockJobQueue) StartProcessing(ctx context.Context, jobType string, handler JobHandler) {
	// Do nothing in the mock implementation
}
