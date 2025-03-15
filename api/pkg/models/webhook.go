package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// WebhookStatus represents the status of a webhook receipt
type WebhookStatus string

const (
	// WebhookStatusPending indicates the webhook is pending processing
	WebhookStatusPending WebhookStatus = "pending"
	// WebhookStatusProcessing indicates the webhook is being processed
	WebhookStatusProcessing WebhookStatus = "processing"
	// WebhookStatusCompleted indicates the webhook has been processed successfully
	WebhookStatusCompleted WebhookStatus = "completed"
	// WebhookStatusFailed indicates the webhook processing failed
	WebhookStatusFailed WebhookStatus = "failed"
)

// WebhookReceipt represents a webhook receipt
type WebhookReceipt struct {
	ID        string        `json:"id" gorm:"primaryKey"`
	Source    string        `json:"source" gorm:"index"`
	Event     string        `json:"event" gorm:"index"`
	Payload   []byte        `json:"payload"`
	Signature string        `json:"signature"`
	Status    WebhookStatus `json:"status" gorm:"index"`
	Error     string        `json:"error,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// WebhookRequest represents the request to create a webhook receipt
type WebhookRequest struct {
	Source    string                 `json:"source" binding:"required"`
	Event     string                 `json:"event" binding:"required"`
	Payload   map[string]interface{} `json:"payload" binding:"required"`
	Signature string                 `json:"signature" binding:"required"`
}

// WebhookResponse represents the response to a webhook request
type WebhookResponse struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message,omitempty"`
}

// NewWebhookReceipt creates a new webhook receipt
func NewWebhookReceipt(source, event string, payload []byte, signature string) *WebhookReceipt {
	return &WebhookReceipt{
		ID:        uuid.New().String(),
		Source:    source,
		Event:     event,
		Payload:   payload,
		Signature: signature,
		Status:    WebhookStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// SetStatus sets the status of the webhook receipt
func (r *WebhookReceipt) SetStatus(status WebhookStatus, err error) {
	r.Status = status
	if err != nil {
		r.Error = err.Error()
	} else {
		r.Error = ""
	}
	r.UpdatedAt = time.Now()
}

// IsComplete returns true if the webhook receipt has been processed
func (r *WebhookReceipt) IsComplete() bool {
	return r.Status == WebhookStatusCompleted || r.Status == WebhookStatusFailed
}

// IsPending returns true if the webhook receipt is pending processing
func (r *WebhookReceipt) IsPending() bool {
	return r.Status == WebhookStatusPending
}

// IsProcessing returns true if the webhook receipt is being processed
func (r *WebhookReceipt) IsProcessing() bool {
	return r.Status == WebhookStatusProcessing
}

// IsFailed returns true if the webhook receipt processing failed
func (r *WebhookReceipt) IsFailed() bool {
	return r.Status == WebhookStatusFailed
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (w *WebhookReceipt) MarshalBinary() ([]byte, error) {
	return json.Marshal(w)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (w *WebhookReceipt) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, w)
}
