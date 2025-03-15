package api

import (
	"net/http"
	"time"

	"github.com/dustinleblanc/go-bespin/internal/queue"
	"github.com/dustinleblanc/go-bespin/pkg/models"
	"github.com/gin-gonic/gin"
)

// Handlers contains the API handlers
type Handlers struct {
	jobQueue *queue.JobQueue
}

// NewHandlers creates a new Handlers instance
func NewHandlers(jobQueue *queue.JobQueue) *Handlers {
	return &Handlers{
		jobQueue: jobQueue,
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
