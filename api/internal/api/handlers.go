package api

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/dustinleblanc/go-bespin-api/internal/queue"
	"github.com/dustinleblanc/go-bespin-api/internal/webhook"
	"github.com/dustinleblanc/go-bespin-api/internal/websocket"
	"github.com/dustinleblanc/go-bespin-api/pkg/models"
	"github.com/gin-gonic/gin"
)

// Handlers contains the HTTP handlers for the API
type Handlers struct {
	jobQueue       queue.Queue
	webhookService webhook.WebhookService
	wsServer       *websocket.Server
}

// NewHandlers creates a new Handlers instance
func NewHandlers(jobQueue queue.Queue, webhookService webhook.WebhookService) *Handlers {
	return &Handlers{
		jobQueue:       jobQueue,
		webhookService: webhookService,
		wsServer:       websocket.NewServer(),
	}
}

// HandleRandomText handles requests to generate random text
func (h *Handlers) HandleRandomText(c *gin.Context) {
	// Get the length parameter from the query string
	lengthStr := c.DefaultQuery("length", "100")
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid length parameter"})
		return
	}

	// Create a new job
	job := &models.Job{
		Type: models.JobTypeRandomText,
		Data: models.RandomTextJobData{
			Length: length,
		},
	}

	// Add the job to the queue
	jobID, err := h.jobQueue.AddJob(c.Request.Context(), job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add job to queue"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"job_id": jobID,
		"status": "queued",
	})
}

// HandleWebhook handles incoming webhook requests
func (h *Handlers) HandleWebhook(c *gin.Context) {
	// Get the source from the URL parameter
	source := c.Param("source")
	if source == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "source is required"})
		return
	}

	// Get the event from the header
	event := c.GetHeader("X-Event-Type")
	if event == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Event-Type header is required"})
		return
	}

	// Get the signature from the header
	signature := c.GetHeader("X-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Signature header is required"})
		return
	}

	// Read the request body
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read request body"})
		return
	}

	// Create a webhook receipt
	receipt, err := h.webhookService.CreateReceipt(c.Request.Context(), source, event, payload, signature)
	if err != nil {
		if err.Error() == fmt.Sprintf("invalid source: %s", source) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a new job
	job := &models.Job{
		Type: models.JobTypeProcessWebhook,
		Data: models.WebhookJobData{
			ReceiptID: receipt.ID,
		},
	}

	// Add the job to the queue
	jobID, err := h.jobQueue.AddJob(c.Request.Context(), job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add job to queue"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":     jobID,
		"receipt_id": receipt.ID,
		"status":     "queued",
	})
}

// HandleGetJobResult handles requests to get a job result
func (h *Handlers) HandleGetJobResult(c *gin.Context) {
	// Get the job ID from the URL parameter
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job ID is required"})
		return
	}

	// Get the job result
	result, err := h.jobQueue.GetJobResult(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get job result: %v", err)})
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// HandleWebSocket handles WebSocket connections
func (h *Handlers) HandleWebSocket(c *gin.Context) {
	// Get the job ID from the query string
	jobID := c.Query("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job_id is required"})
		return
	}

	// Let the WebSocket server handle the connection
	h.wsServer.HandleConnection(c.Writer, c.Request, jobID)
}

// NotifyJobStatus notifies clients about job status changes
func (h *Handlers) NotifyJobStatus(jobID string, status string, result interface{}) {
	h.wsServer.NotifyJobStatus(jobID, status, result)
}
