package webhook

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dustinleblanc/go-bespin/internal/database"
	"github.com/dustinleblanc/go-bespin/pkg/models"
	"gorm.io/gorm"
)

// GormRepository implements the Repository interface using GORM
type GormRepository struct {
	db     *database.GormDB
	logger *log.Logger
}

// NewGormRepository creates a new GORM repository
func NewGormRepository(db *database.GormDB) *GormRepository {
	return &GormRepository{
		db:     db,
		logger: log.New(os.Stdout, "[WebhookGormRepository] ", log.LstdFlags),
	}
}

// Store stores a webhook receipt in the database
func (r *GormRepository) Store(ctx context.Context, receipt *models.WebhookReceipt) error {
	r.logger.Printf("Storing webhook receipt in database: %s from source: %s", receipt.ID, receipt.Source)

	// Use context with GORM
	tx := r.db.DB.WithContext(ctx)

	// Insert into database
	if err := tx.Create(receipt).Error; err != nil {
		r.logger.Printf("Failed to store webhook receipt in database: %v", err)
		return fmt.Errorf("failed to store webhook receipt: %w", err)
	}

	r.logger.Printf("Successfully stored webhook receipt in database: %s", receipt.ID)
	return nil
}

// GetByID retrieves a webhook receipt by ID from the database
func (r *GormRepository) GetByID(ctx context.Context, id string) (*models.WebhookReceipt, error) {
	var receipt models.WebhookReceipt

	// Use context with GORM
	tx := r.db.DB.WithContext(ctx)

	// Query database
	if err := tx.First(&receipt, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("webhook receipt not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get webhook receipt: %w", err)
	}

	return &receipt, nil
}

// List lists webhook receipts by source from the database
func (r *GormRepository) List(ctx context.Context, source string, limit, offset int) ([]*models.WebhookReceipt, error) {
	var receipts []*models.WebhookReceipt

	// Use context with GORM
	tx := r.db.DB.WithContext(ctx)

	// Build query
	query := tx.Model(&models.WebhookReceipt{}).Order("created_at DESC").Limit(limit).Offset(offset)
	if source != "" {
		query = query.Where("source = ?", source)
	}

	// Execute query
	if err := query.Find(&receipts).Error; err != nil {
		return nil, fmt.Errorf("failed to list webhook receipts: %w", err)
	}

	return receipts, nil
}

// Count counts webhook receipts by source from the database
func (r *GormRepository) Count(ctx context.Context, source string) (int, error) {
	var count int64

	// Use context with GORM
	tx := r.db.DB.WithContext(ctx)

	// Build query
	query := tx.Model(&models.WebhookReceipt{})
	if source != "" {
		query = query.Where("source = ?", source)
	}

	// Execute query
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count webhook receipts: %w", err)
	}

	return int(count), nil
}
