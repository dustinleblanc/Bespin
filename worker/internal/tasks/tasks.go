package tasks

import (
	"encoding/json"
	"fmt"
)

const (
	// Task types
	TypeRandomText     = "random_text"
	TypeProcessWebhook = "process_webhook"
)

// RandomTextPayload represents the payload for random text generation jobs
type RandomTextPayload struct {
	Length int `json:"length"`
}

// WebhookPayload represents the payload for webhook processing jobs
type WebhookPayload struct {
	WebhookID string `json:"webhook_id"`
	Source    string `json:"source"`
	Event     string `json:"event"`
}

// NewRandomTextPayload creates a new random text payload
func NewRandomTextPayload(length int) (*RandomTextPayload, error) {
	if length <= 0 || length > 1000 {
		return nil, fmt.Errorf("length must be between 1 and 1000")
	}
	return &RandomTextPayload{Length: length}, nil
}

// NewWebhookPayload creates a new webhook payload
func NewWebhookPayload(webhookID, source, event string) (*WebhookPayload, error) {
	if webhookID == "" {
		return nil, fmt.Errorf("webhook ID is required")
	}
	if source == "" {
		return nil, fmt.Errorf("source is required")
	}
	if event == "" {
		return nil, fmt.Errorf("event is required")
	}
	return &WebhookPayload{
		WebhookID: webhookID,
		Source:    source,
		Event:     event,
	}, nil
}

// Serialize converts a payload to bytes
func Serialize(payload interface{}) ([]byte, error) {
	return json.Marshal(payload)
}

// DeserializeRandomText converts bytes to a random text payload
func DeserializeRandomText(data []byte) (*RandomTextPayload, error) {
	var p RandomTextPayload
	err := json.Unmarshal(data, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize random text payload: %w", err)
	}
	return &p, nil
}

// DeserializeWebhook converts bytes to a webhook payload
func DeserializeWebhook(data []byte) (*WebhookPayload, error) {
	var p WebhookPayload
	err := json.Unmarshal(data, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize webhook payload: %w", err)
	}
	return &p, nil
}
