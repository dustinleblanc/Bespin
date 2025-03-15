package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dustinleblanc/go-bespin/pkg/models"
)

// Service handles webhook operations
type Service struct {
	repository Repository
	logger     *log.Logger
	secrets    map[string]string
}

// NewService creates a new webhook service
func NewService(repository Repository) *Service {
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
		repository: repository,
		logger:     log.New(log.Writer(), "[WebhookService] ", log.LstdFlags),
		secrets:    secrets,
	}
}

// VerifySignature verifies the webhook signature
func (s *Service) VerifySignature(source string, payload []byte, signature string) bool {
	secret, ok := s.secrets[source]
	if !ok {
		s.logger.Printf("Unknown webhook source: %s", source)
		return false
	}

	// If no secret is configured, we can't verify
	if secret == "" {
		s.logger.Printf("No secret configured for source: %s", source)
		return false
	}

	// Calculate expected signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Clean up the provided signature (some services prefix with algorithm)
	cleanSignature := signature
	if strings.Contains(signature, "=") {
		parts := strings.Split(signature, "=")
		if len(parts) == 2 {
			cleanSignature = parts[1]
		}
	}

	// Compare signatures
	return hmac.Equal([]byte(cleanSignature), []byte(expectedSignature))
}

// StoreWebhook stores a webhook receipt
func (s *Service) StoreWebhook(ctx context.Context, receipt *models.WebhookReceipt) error {
	// Store in repository
	if err := s.repository.Store(ctx, receipt); err != nil {
		return fmt.Errorf("failed to store webhook receipt: %w", err)
	}

	s.logger.Printf("Stored webhook receipt: %s from source: %s", receipt.ID, receipt.Source)
	return nil
}

// GetWebhook retrieves a webhook receipt by ID
func (s *Service) GetWebhook(ctx context.Context, id string) (*models.WebhookReceipt, error) {
	// Get from repository
	receipt, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook receipt: %w", err)
	}

	return receipt, nil
}

// ListWebhooks lists webhook receipts by source
func (s *Service) ListWebhooks(ctx context.Context, source string, limit, offset int) ([]*models.WebhookReceipt, error) {
	// Get from repository
	receipts, err := s.repository.List(ctx, source, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhook receipts: %w", err)
	}

	return receipts, nil
}

// CountWebhooks counts webhook receipts by source
func (s *Service) CountWebhooks(ctx context.Context, source string) (int, error) {
	// Get from repository
	count, err := s.repository.Count(ctx, source)
	if err != nil {
		return 0, fmt.Errorf("failed to count webhook receipts: %w", err)
	}

	return count, nil
}

// IsValidSource checks if a webhook source is valid
func (s *Service) IsValidSource(source string) bool {
	_, ok := s.secrets[source]
	return ok
}
