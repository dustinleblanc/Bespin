package webhook

import (
	"context"
	"fmt"

	"github.com/dustinleblanc/go-bespin-api/pkg/models"
	"gorm.io/gorm"
)

// GormRepository implements Repository using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM repository
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// Create creates a new webhook receipt
func (r *GormRepository) Create(ctx context.Context, receipt *models.WebhookReceipt) error {
	result := r.db.WithContext(ctx).Create(receipt)
	if result.Error != nil {
		return fmt.Errorf("failed to create webhook receipt: %w", result.Error)
	}
	return nil
}

// GetByID retrieves a webhook receipt by ID
func (r *GormRepository) GetByID(ctx context.Context, id string) (*models.WebhookReceipt, error) {
	var receipt models.WebhookReceipt
	result := r.db.WithContext(ctx).First(&receipt, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("webhook receipt not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get webhook receipt: %w", result.Error)
	}
	return &receipt, nil
}

// Update updates a webhook receipt
func (r *GormRepository) Update(ctx context.Context, receipt *models.WebhookReceipt) error {
	result := r.db.WithContext(ctx).Save(receipt)
	if result.Error != nil {
		return fmt.Errorf("failed to update webhook receipt: %w", result.Error)
	}
	return nil
}

// List retrieves a list of webhook receipts with optional filtering
func (r *GormRepository) List(ctx context.Context, source string, limit, offset int) ([]*models.WebhookReceipt, error) {
	var receipts []*models.WebhookReceipt
	query := r.db.WithContext(ctx)

	if source != "" {
		query = query.Where("source = ?", source)
	}

	result := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&receipts)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list webhook receipts: %w", result.Error)
	}
	return receipts, nil
}

// Count counts webhook receipts with optional filtering
func (r *GormRepository) Count(ctx context.Context, source string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.WebhookReceipt{})

	if source != "" {
		query = query.Where("source = ?", source)
	}

	result := query.Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count webhook receipts: %w", result.Error)
	}
	return count, nil
}
