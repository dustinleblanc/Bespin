package jobs

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/dustinleblanc/go-bespin-worker/pkg/tasks"
	"github.com/hibiken/asynq"
)

// Processor handles job processing
type Processor struct {
	logger *log.Logger
}

// NewProcessor creates a new job processor
func NewProcessor() *Processor {
	return &Processor{
		logger: log.New(log.Writer(), "[JobProcessor] ", log.LstdFlags),
	}
}

// HandleRandomTextTask processes a random text job
func (p *Processor) HandleRandomTextTask(ctx context.Context, t *asynq.Task) error {
	payload, err := tasks.DeserializeRandomText(t.Payload())
	if err != nil {
		return fmt.Errorf("failed to deserialize random text payload: %w", err)
	}

	p.logger.Printf("Processing random text job with length: %d", payload.Length)

	// Generate random text
	result := p.generateRandomText(payload.Length)

	// In a real application, you might want to store the result somewhere
	// or send it back through a channel/webhook
	p.logger.Printf("Generated random text: %s", result)

	return nil
}

// HandleWebhookTask processes a webhook job
func (p *Processor) HandleWebhookTask(ctx context.Context, t *asynq.Task) error {
	payload, err := tasks.DeserializeWebhook(t.Payload())
	if err != nil {
		return fmt.Errorf("failed to deserialize webhook payload: %w", err)
	}

	p.logger.Printf("Processing webhook job: ID=%s, Source=%s, Event=%s",
		payload.WebhookID, payload.Source, payload.Event)

	// Here you would typically:
	// 1. Fetch the webhook data from the database
	// 2. Process it according to the source and event type
	// 3. Update the webhook status in the database
	// 4. Send any necessary notifications

	return nil
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
