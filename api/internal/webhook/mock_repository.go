package webhook

import (
	"context"
	"fmt"
	"sync"

	"github.com/dustinleblanc/go-bespin/pkg/models"
)

// MockRepository implements the Repository interface for testing
type MockRepository struct {
	webhooks map[string]*models.WebhookReceipt
	sources  map[string][]string
	mu       sync.RWMutex
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{
		webhooks: make(map[string]*models.WebhookReceipt),
		sources:  make(map[string][]string),
	}
}

// Store stores a webhook receipt in memory
func (r *MockRepository) Store(ctx context.Context, receipt *models.WebhookReceipt) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store webhook
	r.webhooks[receipt.ID] = receipt

	// Add to source list
	if _, ok := r.sources[receipt.Source]; !ok {
		r.sources[receipt.Source] = make([]string, 0)
	}
	r.sources[receipt.Source] = append(r.sources[receipt.Source], receipt.ID)

	// Add to all list
	if _, ok := r.sources["all"]; !ok {
		r.sources["all"] = make([]string, 0)
	}
	r.sources["all"] = append(r.sources["all"], receipt.ID)

	return nil
}

// GetByID retrieves a webhook receipt by ID from memory
func (r *MockRepository) GetByID(ctx context.Context, id string) (*models.WebhookReceipt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	receipt, ok := r.webhooks[id]
	if !ok {
		return nil, fmt.Errorf("webhook receipt not found: %s", id)
	}

	return receipt, nil
}

// List lists webhook receipts by source from memory
func (r *MockRepository) List(ctx context.Context, source string, limit, offset int) ([]*models.WebhookReceipt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Get IDs
	var ids []string
	if source == "" {
		ids = r.sources["all"]
	} else {
		ids = r.sources[source]
	}

	// Apply pagination
	if offset >= len(ids) {
		return []*models.WebhookReceipt{}, nil
	}

	end := offset + limit
	if end > len(ids) {
		end = len(ids)
	}
	ids = ids[offset:end]

	// Get webhooks
	receipts := make([]*models.WebhookReceipt, 0, len(ids))
	for _, id := range ids {
		receipt, ok := r.webhooks[id]
		if ok {
			receipts = append(receipts, receipt)
		}
	}

	return receipts, nil
}

// Count counts webhook receipts by source from memory
func (r *MockRepository) Count(ctx context.Context, source string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if source == "" {
		return len(r.sources["all"]), nil
	}

	return len(r.sources[source]), nil
}
