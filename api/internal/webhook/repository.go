package webhook

import (
	"context"

	"github.com/dustinleblanc/go-bespin/pkg/models"
)

// Repository defines the interface for webhook storage operations
type Repository interface {
	Store(ctx context.Context, receipt *models.WebhookReceipt) error
	GetByID(ctx context.Context, id string) (*models.WebhookReceipt, error)
	List(ctx context.Context, source string, limit, offset int) ([]*models.WebhookReceipt, error)
	Count(ctx context.Context, source string) (int, error)
}
