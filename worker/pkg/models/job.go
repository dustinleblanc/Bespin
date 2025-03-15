package models

import (
	"time"
)

// Job represents a job in the queue
type Job struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

// JobResult represents the result of a job
type JobResult struct {
	JobID     string      `json:"job_id"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error"`
	CreatedAt time.Time   `json:"created_at"`
}

// WebhookJobData represents the data for a webhook processing job
type WebhookJobData struct {
	WebhookID string `json:"webhook_id"`
	Source    string `json:"source"`
	Event     string `json:"event"`
}

// Job status constants
const (
	JobStatusPending    = "pending"
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
		Status:    JobStatusPending,
	}
}

// JobResponse represents the response when creating a job
type JobResponse struct {
	JobID string `json:"job_id"`
}
