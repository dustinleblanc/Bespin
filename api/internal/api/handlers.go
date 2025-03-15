package api

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/dustinleblanc/go-bespin/internal/queue"
	"github.com/dustinleblanc/go-bespin/internal/webhook"
	"github.com/dustinleblanc/go-bespin/pkg/models"
	"github.com/gin-gonic/gin"
)

// Handlers contains the API handlers
type Handlers struct {
	jobQueue       *queue.JobQueue
	webhookService *webhook.Service
}

// NewHandlers creates a new Handlers instance
func NewHandlers(jobQueue *queue.JobQueue, webhookService *webhook.Service) *Handlers {
	return &Handlers{
		jobQueue:       jobQueue,
		webhookService: webhookService,
	}
}

// GetRoot handles the root endpoint
func (h *Handlers) GetRoot(c *gin.Context) {
	c.String(http.StatusOK, "API server is running")
}

// GetTest handles the test endpoint
func (h *Handlers) GetTest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":   "API is working!",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// GetJobsTest handles the jobs test endpoint
func (h *Handlers) GetJobsTest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":   "Jobs API is working!",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// CreateRandomTextJob handles the random text job creation endpoint
func (h *Handlers) CreateRandomTextJob(c *gin.Context) {
	var request models.RandomTextJobRequest

	// Set default value
	request.Length = 100

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate length
	if request.Length < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Length must be at least 1"})
		return
	}

	// Create job
	jobID, err := h.jobQueue.AddJob("random-text", request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job"})
		return
	}

	c.JSON(http.StatusOK, models.JobResponse{JobID: jobID})
}

// ReceiveWebhook handles incoming webhooks
func (h *Handlers) ReceiveWebhook(c *gin.Context) {
	// Get source from URL parameter
	source := c.Param("source")
	if source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Source is required"})
		return
	}

	// Get signature from header
	signature := c.GetHeader("X-Webhook-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Signature header is required"})
		return
	}

	// Check if source is valid before proceeding
	if !h.webhookService.IsValidSource(source) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown webhook source"})
		return
	}

	// Read the raw body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	// Restore the request body for binding
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Parse the JSON body
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Get event type from payload or header
	event := ""
	if eventVal, ok := payload["event"]; ok {
		if eventStr, ok := eventVal.(string); ok {
			event = eventStr
		}
	}
	if event == "" {
		event = c.GetHeader("X-Webhook-Event")
	}
	if event == "" {
		event = "unknown"
	}

	// Collect headers
	headers := make(map[string]string)
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// Verify signature before proceeding
	verified := h.webhookService.VerifySignature(source, bodyBytes, signature)
	if !verified {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Create webhook receipt
	receipt := models.NewWebhookReceipt(source, event, payload, headers, signature, verified)

	// Store webhook receipt
	if err := h.webhookService.StoreWebhook(c, receipt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store webhook"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, models.WebhookResponse{
		ID:        receipt.ID,
		Verified:  receipt.Verified,
		CreatedAt: receipt.CreatedAt,
	})
}

// GetWebhook handles retrieving a webhook receipt
func (h *Handlers) GetWebhook(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	receipt, err := h.webhookService.GetWebhook(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// ListWebhooks handles listing webhook receipts
func (h *Handlers) ListWebhooks(c *gin.Context) {
	source := c.Query("source")
	limit := 10
	offset := 0

	// Get total count
	count, err := h.webhookService.CountWebhooks(c, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get webhooks
	receipts, err := h.webhookService.ListWebhooks(c, source, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"webhooks": receipts,
		"count":    count,
	})
}
