package models

import (
	"time"
)

// JobType represents the type of job
type JobType string

const (
	// JobTypeRandomText represents a random text generation job
	JobTypeRandomText JobType = "random_text"
	// JobTypeProcessWebhook represents a webhook processing job
	JobTypeProcessWebhook JobType = "process_webhook"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	// JobStatusPending indicates the job is pending execution
	JobStatusPending JobStatus = "pending"
	// JobStatusProcessing indicates the job is being processed
	JobStatusProcessing JobStatus = "processing"
	// JobStatusCompleted indicates the job has completed successfully
	JobStatusCompleted JobStatus = "completed"
	// JobStatusFailed indicates the job has failed
	JobStatusFailed JobStatus = "failed"
	// JobStatusRetrying indicates the job is being retried
	JobStatusRetrying JobStatus = "retrying"
)

// Job represents a job to be processed
type Job struct {
	Type JobType     `json:"type"`
	Data interface{} `json:"data"`
}

// JobResult represents the result of a job
type JobResult struct {
	ID          string     `json:"id"`
	Status      JobStatus  `json:"status"`
	Result      string     `json:"result,omitempty"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// RandomTextJobData represents the data for a random text generation job
type RandomTextJobData struct {
	Length int `json:"length"`
}

// WebhookJobData represents the data for a webhook processing job
type WebhookJobData struct {
	ReceiptID string `json:"receipt_id"`
}

// JobResponse represents the response when creating a job
type JobResponse struct {
	JobID string `json:"jobId"`
}
