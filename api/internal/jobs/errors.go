package jobs

import (
	"errors"
)

// Error definitions
var (
	ErrInvalidJobData = errors.New("invalid job data")
	ErrJobNotFound    = errors.New("job not found")
	ErrJobFailed      = errors.New("job failed")
)
