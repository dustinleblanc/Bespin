package models

import (
	"time"
)

// Job represents a job in the queue
type Job struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	CreatedAt time.Time   `json:"createdAt"`
	Status    string      `json:"status"`
}

// WebhookJobData represents the data for a webhook processing job
type WebhookJobData struct {
	WebhookID string `json:"webhookId"`
	Source    string `json:"source"`
	Event     string `json:"event"`
}

// JobResult represents the result of a job
type JobResult struct {
	JobID       string      `json:"jobId"`
	Result      interface{} `json:"result"`
	Error       string      `json:"error,omitempty"`
	CompletedAt time.Time   `json:"completedAt"`
}

// JobStatus constants
const (
	JobStatusQueued     = "queued"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
)

// NewJob creates a new job with the given type and data
func NewJob(jobType string, data interface{}) *Job {
	return &Job{
		Type:      jobType,
		Data:      data,
		CreatedAt: time.Now(),
		Status:    JobStatusQueued,
	}
}

// JobResponse represents the response when creating a job
type JobResponse struct {
	JobID string `json:"jobId"`
}
