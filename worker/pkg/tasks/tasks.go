package tasks

import (
	"encoding/json"
	"fmt"
)

// Task types
const (
	TypeRandomText = "random-text"
	TypeWebhook    = "webhook"
)

// RandomTextPayload represents the payload for a random text task
type RandomTextPayload struct {
	Length int `json:"length"`
}

// WebhookPayload represents the payload for a webhook task
type WebhookPayload struct {
	WebhookID string `json:"webhook_id"`
	Source    string `json:"source"`
	Event     string `json:"event"`
}

// SerializeRandomText serializes a random text payload
func SerializeRandomText(p *RandomTextPayload) ([]byte, error) {
	return json.Marshal(p)
}

// DeserializeRandomText deserializes a random text payload
func DeserializeRandomText(data []byte) (*RandomTextPayload, error) {
	var p RandomTextPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to deserialize random text payload: %w", err)
	}
	return &p, nil
}

// SerializeWebhook serializes a webhook payload
func SerializeWebhook(p *WebhookPayload) ([]byte, error) {
	return json.Marshal(p)
}

// DeserializeWebhook deserializes a webhook payload
func DeserializeWebhook(data []byte) (*WebhookPayload, error) {
	var p WebhookPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to deserialize webhook payload: %w", err)
	}
	return &p, nil
}
