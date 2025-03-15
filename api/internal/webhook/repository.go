package webhook

import (
	"context"

	"github.com/dustinleblanc/go-bespin-api/pkg/models"
)

// Repository defines the interface for webhook storage
type Repository interface {
	// Create creates a new webhook receipt
	Create(ctx context.Context, receipt *models.WebhookReceipt) error

	// GetByID retrieves a webhook receipt by ID
	GetByID(ctx context.Context, id string) (*models.WebhookReceipt, error)

	// Update updates a webhook receipt
	Update(ctx context.Context, receipt *models.WebhookReceipt) error

	// List retrieves a list of webhook receipts with optional filtering
	List(ctx context.Context, source string, limit, offset int) ([]*models.WebhookReceipt, error)

	// Count counts webhook receipts with optional filtering
	Count(ctx context.Context, source string) (int64, error)
}
