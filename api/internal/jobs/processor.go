package jobs

import (
	"context"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/dustinleblanc/go-bespin/internal/queue"
	"github.com/dustinleblanc/go-bespin/pkg/models"
)

// Processor handles job processing
type Processor struct {
	jobQueue *queue.JobQueue
	logger   *log.Logger
}

// NewProcessor creates a new job processor
func NewProcessor(jobQueue *queue.JobQueue) *Processor {
	return &Processor{
		jobQueue: jobQueue,
		logger:   log.New(log.Writer(), "[JobProcessor] ", log.LstdFlags),
	}
}

// Start starts the job processor
func (p *Processor) Start(ctx context.Context) {
	p.logger.Println("Starting job processor")

	// Register job handlers
	p.registerJobHandlers(ctx)
}

// registerJobHandlers registers handlers for different job types
func (p *Processor) registerJobHandlers(ctx context.Context) {
	// Register random text job handler
	p.jobQueue.StartProcessing(ctx, "random-text", p.processRandomTextJob)
}

// processRandomTextJob processes a random text job
func (p *Processor) processRandomTextJob(job *models.Job) (interface{}, error) {
	p.logger.Printf("Processing random text job: %s", job.ID)

	// Extract job data
	data, ok := job.Data.(map[string]interface{})
	if !ok {
		p.logger.Printf("Invalid job data format: %v", job.Data)
		return nil, ErrInvalidJobData
	}

	// Get length parameter
	lengthFloat, ok := data["length"].(float64)
	if !ok {
		p.logger.Printf("Invalid length parameter: %v", data["length"])
		return nil, ErrInvalidJobData
	}

	length := int(lengthFloat)

	// Generate random text
	result := p.generateRandomText(length)

	p.logger.Printf("Completed random text job: %s", job.ID)
	return result, nil
}

// generateRandomText generates a random text of the specified length
func (p *Processor) generateRandomText(length int) string {
	p.logger.Printf("Generating random text of length: %d", length)

	// Simulate processing time
	time.Sleep(2 * time.Second)

	words := []string{
		"cloud", "computing", "platform", "service", "data",
		"storage", "network", "server", "virtual", "container",
		"function", "application", "microservice", "kubernetes", "docker",
		"infrastructure", "code", "deployment", "scaling", "monitoring",
	}

	var result strings.Builder

	for i := 0; i < length; i++ {
		randomIndex := rand.Intn(len(words))
		result.WriteString(words[randomIndex])
		result.WriteString(" ")
	}

	return strings.TrimSpace(result.String())
}
