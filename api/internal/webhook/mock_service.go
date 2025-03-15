package webhook

import (
	"context"

	"github.com/dustinleblanc/go-bespin-api/pkg/models"
	"github.com/stretchr/testify/mock"
)

// WebhookService defines the interface for webhook operations
type WebhookService interface {
	VerifySignature(source string, payload []byte, signature string) bool
	CreateReceipt(ctx context.Context, source, event string, payload []byte, signature string) (*models.WebhookReceipt, error)
	GetReceipt(ctx context.Context, id string) (*models.WebhookReceipt, error)
	UpdateReceipt(ctx context.Context, receipt *models.WebhookReceipt) error
	ListReceipts(ctx context.Context, source string, limit, offset int) ([]*models.WebhookReceipt, error)
	CountReceipts(ctx context.Context, source string) (int64, error)
	IsValidSource(source string) bool
}

// Ensure MockService implements WebhookService
var _ WebhookService = (*MockService)(nil)

// MockService is a mock implementation of the webhook service
type MockService struct {
	mock.Mock
}

// NewMockService creates a new mock service
func NewMockService() *MockService {
	return &MockService{}
}

// VerifySignature verifies the webhook signature
func (s *MockService) VerifySignature(source string, payload []byte, signature string) bool {
	args := s.Called(source, payload, signature)
	return args.Bool(0)
}

// CreateReceipt creates a new webhook receipt
func (s *MockService) CreateReceipt(ctx context.Context, source, event string, payload []byte, signature string) (*models.WebhookReceipt, error) {
	args := s.Called(ctx, source, event, payload, signature)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WebhookReceipt), args.Error(1)
}

// GetReceipt gets a webhook receipt by ID
func (s *MockService) GetReceipt(ctx context.Context, id string) (*models.WebhookReceipt, error) {
	args := s.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WebhookReceipt), args.Error(1)
}

// UpdateReceipt updates a webhook receipt
func (s *MockService) UpdateReceipt(ctx context.Context, receipt *models.WebhookReceipt) error {
	args := s.Called(ctx, receipt)
	return args.Error(0)
}

// ListReceipts lists webhook receipts for a source
func (s *MockService) ListReceipts(ctx context.Context, source string, limit, offset int) ([]*models.WebhookReceipt, error) {
	args := s.Called(ctx, source, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.WebhookReceipt), args.Error(1)
}

// CountReceipts counts webhook receipts for a source
func (s *MockService) CountReceipts(ctx context.Context, source string) (int64, error) {
	args := s.Called(ctx, source)
	return args.Get(0).(int64), args.Error(1)
}

// IsValidSource checks if a source is valid
func (s *MockService) IsValidSource(source string) bool {
	args := s.Called(source)
	return args.Bool(0)
}
