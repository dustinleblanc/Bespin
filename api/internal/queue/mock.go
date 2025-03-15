package queue

import (
	"context"

	"github.com/dustinleblanc/go-bespin-api/pkg/models"
	"github.com/stretchr/testify/mock"
)

// MockQueue is a mock implementation of the Queue interface
type MockQueue struct {
	mock.Mock
}

// AddJob mocks the AddJob method
func (m *MockQueue) AddJob(ctx context.Context, job *models.Job) (string, error) {
	args := m.Called(ctx, job)
	return args.String(0), args.Error(1)
}

// GetJobResult mocks the GetJobResult method
func (m *MockQueue) GetJobResult(ctx context.Context, jobID string) (*models.JobResult, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.JobResult), args.Error(1)
}

// Close mocks the Close method
func (m *MockQueue) Close() error {
	args := m.Called()
	return args.Error(0)
}
