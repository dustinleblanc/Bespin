package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// WebhookReceipt represents a received webhook
type WebhookReceipt struct {
	ID        string         `json:"id" gorm:"primaryKey;type:uuid"`
	Source    string         `json:"source" gorm:"index;type:varchar(255)"`
	Event     string         `json:"event" gorm:"index;type:varchar(255)"`
	Payload   datatypes.JSON `json:"payload" gorm:"type:jsonb"`
	Headers   datatypes.JSON `json:"headers" gorm:"type:jsonb"`
	Signature string         `json:"signature" gorm:"type:text"`
	Verified  bool           `json:"verified" gorm:"index"`
	CreatedAt time.Time      `json:"created_at" gorm:"index;autoCreateTime"`
}

// WebhookRequest represents the request to create a webhook receipt
type WebhookRequest struct {
	Source    string                 `json:"source" binding:"required"`
	Event     string                 `json:"event" binding:"required"`
	Payload   map[string]interface{} `json:"payload" binding:"required"`
	Signature string                 `json:"signature" binding:"required"`
}

// WebhookResponse represents the response when creating a webhook receipt
type WebhookResponse struct {
	ID        string    `json:"id"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
	JobID     string    `json:"job_id,omitempty"`
	Warning   string    `json:"warning,omitempty"`
}

// NewWebhookReceipt creates a new webhook receipt
func NewWebhookReceipt(source, event string, payload map[string]interface{}, headers map[string]string, signature string, verified bool) *WebhookReceipt {
	// Convert headers map to JSON-compatible map
	jsonHeaders := make(map[string]interface{})
	for k, v := range headers {
		jsonHeaders[k] = v
	}

	// Convert payload and headers to JSON
	payloadBytes, _ := json.Marshal(payload)
	headersBytes, _ := json.Marshal(jsonHeaders)

	return &WebhookReceipt{
		ID:        uuid.New().String(),
		Source:    source,
		Event:     event,
		Payload:   datatypes.JSON(payloadBytes),
		Headers:   datatypes.JSON(headersBytes),
		Signature: signature,
		Verified:  verified,
		CreatedAt: time.Now(),
	}
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (w *WebhookReceipt) MarshalBinary() ([]byte, error) {
	return json.Marshal(w)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (w *WebhookReceipt) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, w)
}
