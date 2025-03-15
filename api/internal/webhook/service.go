package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/dustinleblanc/go-bespin-api/pkg/models"
)

// Ensure Service implements WebhookService
var _ WebhookService = (*Service)(nil)

// Service handles webhook operations
type Service struct {
	repo    Repository
	logger  *log.Logger
	secrets map[string]string
}

// NewService creates a new webhook service
func NewService(repo Repository) *Service {
	// Load secrets from environment variables
	secrets := make(map[string]string)

	// Load webhook secrets from environment variables
	if secret := os.Getenv("GITHUB_WEBHOOK_SECRET"); secret != "" {
		secrets["github"] = secret
	}
	if secret := os.Getenv("STRIPE_WEBHOOK_SECRET"); secret != "" {
		secrets["stripe"] = secret
	}
	if secret := os.Getenv("SENDGRID_WEBHOOK_SECRET"); secret != "" {
		secrets["sendgrid"] = secret
	}

	return &Service{
		repo:    repo,
		logger:  log.New(log.Writer(), "[WebhookService] ", log.LstdFlags),
		secrets: secrets,
	}
}

// VerifySignature verifies the webhook signature
func (s *Service) VerifySignature(source string, payload []byte, signature string) bool {
	secret, ok := s.secrets[source]
	if !ok {
		s.logger.Printf("No secret found for source: %s", source)
		return false
	}

	// Create HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// CreateReceipt creates a new webhook receipt
func (s *Service) CreateReceipt(ctx context.Context, source, event string, payload []byte, signature string) (*models.WebhookReceipt, error) {
	if !s.IsValidSource(source) {
		return nil, fmt.Errorf("invalid source: %s", source)
	}

	if event == "" {
		return nil, fmt.Errorf("event is required")
	}

	if len(payload) == 0 {
		return nil, fmt.Errorf("payload is required")
	}

	if signature == "" {
		return nil, fmt.Errorf("signature is required")
	}

	// Verify signature
	if !s.VerifySignature(source, payload, signature) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Create receipt
	receipt := models.NewWebhookReceipt(source, event, payload, signature)

	// Save receipt
	if err := s.repo.Create(ctx, receipt); err != nil {
		return nil, fmt.Errorf("failed to save receipt: %w", err)
	}

	return receipt, nil
}

// GetReceipt gets a webhook receipt by ID
func (s *Service) GetReceipt(ctx context.Context, id string) (*models.WebhookReceipt, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	receipt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt: %w", err)
	}

	return receipt, nil
}

// UpdateReceipt updates a webhook receipt
func (s *Service) UpdateReceipt(ctx context.Context, receipt *models.WebhookReceipt) error {
	if receipt == nil {
		return fmt.Errorf("receipt is required")
	}

	if err := s.repo.Update(ctx, receipt); err != nil {
		return fmt.Errorf("failed to update receipt: %w", err)
	}

	return nil
}

// ListReceipts lists webhook receipts for a source
func (s *Service) ListReceipts(ctx context.Context, source string, limit, offset int) ([]*models.WebhookReceipt, error) {
	if source != "" && !s.IsValidSource(source) {
		return nil, fmt.Errorf("invalid source: %s", source)
	}

	receipts, err := s.repo.List(ctx, source, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list receipts: %w", err)
	}

	return receipts, nil
}

// CountReceipts counts webhook receipts for a source
func (s *Service) CountReceipts(ctx context.Context, source string) (int64, error) {
	if source != "" && !s.IsValidSource(source) {
		return 0, fmt.Errorf("invalid source: %s", source)
	}

	count, err := s.repo.Count(ctx, source)
	if err != nil {
		return 0, fmt.Errorf("failed to count receipts: %w", err)
	}

	return count, nil
}

// IsValidSource checks if a source is valid
func (s *Service) IsValidSource(source string) bool {
	_, ok := s.secrets[source]
	return ok
}
