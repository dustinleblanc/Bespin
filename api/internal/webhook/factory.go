package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dustinleblanc/go-bespin/pkg/models"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Factory provides methods for creating webhook receipts
type Factory struct {
	secrets map[string]string
}

// NewFactory creates a new webhook factory
func NewFactory() *Factory {
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

	return &Factory{
		secrets: secrets,
	}
}

// CreateWebhookReceipt creates a new webhook receipt
func (f *Factory) CreateWebhookReceipt(source, event string, payload map[string]interface{}) *models.WebhookReceipt {
	// Generate a signature
	payloadBytes, _ := json.Marshal(payload)
	signature := f.GenerateSignature(source, payloadBytes)

	// Create headers
	headers := map[string]string{
		"Content-Type":        "application/json",
		"X-Webhook-Signature": signature,
		"X-Webhook-Event":     event,
	}

	// Convert headers to JSON-compatible map
	jsonHeaders := make(map[string]interface{})
	for k, v := range headers {
		jsonHeaders[k] = v
	}

	// Convert payload and headers to JSON
	headersBytes, _ := json.Marshal(jsonHeaders)

	// Create receipt
	return &models.WebhookReceipt{
		ID:        uuid.New().String(),
		Source:    source,
		Event:     event,
		Payload:   datatypes.JSON(payloadBytes),
		Headers:   datatypes.JSON(headersBytes),
		Signature: signature,
		Verified:  true,
		CreatedAt: time.Now(),
	}
}

// CreateGithubWebhook creates a GitHub webhook receipt
func (f *Factory) CreateGithubWebhook(event string) *models.WebhookReceipt {
	payload := map[string]interface{}{
		"event":       event,
		"repository":  "bespin",
		"sender":      "user",
		"action":      "opened",
		"timestamp":   time.Now().Format(time.RFC3339),
		"description": fmt.Sprintf("GitHub %s event", event),
	}

	return f.CreateWebhookReceipt("github", event, payload)
}

// CreateStripeWebhook creates a Stripe webhook receipt
func (f *Factory) CreateStripeWebhook(event string) *models.WebhookReceipt {
	payload := map[string]interface{}{
		"event":       event,
		"object":      "event",
		"api_version": "2020-08-27",
		"created":     time.Now().Unix(),
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":      uuid.New().String(),
				"object":  "payment_intent",
				"amount":  1000,
				"status":  "succeeded",
				"created": time.Now().Unix(),
			},
		},
	}

	return f.CreateWebhookReceipt("stripe", event, payload)
}

// GenerateSignature generates a signature for a webhook payload
func (f *Factory) GenerateSignature(source string, payload []byte) string {
	secret, ok := f.secrets[source]
	if !ok {
		return "invalid-signature"
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
