package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dustinleblanc/go-bespin-api/internal/queue"
	"github.com/dustinleblanc/go-bespin-api/internal/webhook"
	internalws "github.com/dustinleblanc/go-bespin-api/internal/websocket"
	"github.com/dustinleblanc/go-bespin-api/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	// Set up test environment
	os.Setenv("GO_ENV", "test")
	os.Setenv("GITHUB_WEBHOOK_SECRET", "test-secret-for-testing")
}

func TestHandleRandomText(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockQueue := &queue.MockQueue{}
	mockRepo := webhook.NewMockRepository()
	webhookService := webhook.NewService(mockRepo)
	handlers := NewHandlers(mockQueue, webhookService)

	router := gin.New()
	router.GET("/random-text", handlers.HandleRandomText)

	tests := []struct {
		name       string
		length     string
		wantStatus int
	}{
		{
			name:       "valid request",
			length:     "10",
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "invalid length - not a number",
			length:     "invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no length parameter",
			length:     "",
			wantStatus: http.StatusAccepted, // Uses default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock expectations
			if tt.wantStatus == http.StatusAccepted {
				mockQueue.On("AddJob", mock.Anything, mock.MatchedBy(func(job *models.Job) bool {
					return job.Type == models.JobTypeRandomText
				})).Return("test-job-id", nil).Once()
			}

			// Create request
			url := "/random-text"
			if tt.length != "" {
				url += "?length=" + tt.length
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.wantStatus, w.Code)

			// Verify mock expectations
			mockQueue.AssertExpectations(t)
		})
	}
}

func TestHandleWebhook(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := webhook.NewMockService()

	testCases := []struct {
		name           string
		source         string
		event          string
		payload        map[string]interface{}
		wantStatus     int
		expectJobQueue bool
		setupMocks     func(*webhook.MockService, *queue.MockQueue, map[string]interface{})
	}{
		{
			name:   "valid request",
			source: "github",
			event:  "push",
			payload: map[string]interface{}{
				"test": "data",
			},
			wantStatus:     http.StatusAccepted,
			expectJobQueue: true,
			setupMocks: func(s *webhook.MockService, q *queue.MockQueue, payload map[string]interface{}) {
				payloadBytes, _ := json.Marshal(payload)
				signature := generateSignature(payloadBytes)

				s.On("CreateReceipt", mock.Anything, "github", "push", payloadBytes, signature).Return(&models.WebhookReceipt{
					ID:        "test-receipt-id",
					Source:    "github",
					Event:     "push",
					Payload:   payloadBytes,
					Signature: signature,
				}, nil).Once()

				q.On("AddJob", mock.Anything, mock.MatchedBy(func(job *models.Job) bool {
					return job.Type == models.JobTypeProcessWebhook && job.Data.(models.WebhookJobData).ReceiptID == "test-receipt-id"
				})).Return("test-job-id", nil).Once()
			},
		},
		{
			name:       "missing source",
			source:     "",
			event:      "push",
			payload:    map[string]interface{}{},
			wantStatus: http.StatusNotFound,
			setupMocks: func(s *webhook.MockService, q *queue.MockQueue, payload map[string]interface{}) {},
		},
		{
			name:       "missing event",
			source:     "github",
			event:      "",
			payload:    map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
			setupMocks: func(s *webhook.MockService, q *queue.MockQueue, payload map[string]interface{}) {},
		},
		{
			name:       "missing signature",
			source:     "github",
			event:      "push",
			payload:    map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
			setupMocks: func(s *webhook.MockService, q *queue.MockQueue, payload map[string]interface{}) {},
		},
		{
			name:   "invalid signature",
			source: "github",
			event:  "push",
			payload: map[string]interface{}{
				"test": "data",
			},
			wantStatus: http.StatusBadRequest,
			setupMocks: func(s *webhook.MockService, q *queue.MockQueue, payload map[string]interface{}) {
				payloadBytes, _ := json.Marshal(payload)
				signature := generateSignature(payloadBytes)
				s.On("CreateReceipt", mock.Anything, "github", "push", payloadBytes, signature).Return(nil, fmt.Errorf("invalid signature")).Once()
			},
		},
		{
			name:   "invalid source",
			source: "invalid",
			event:  "push",
			payload: map[string]interface{}{
				"test": "data",
			},
			wantStatus: http.StatusNotFound,
			setupMocks: func(s *webhook.MockService, q *queue.MockQueue, payload map[string]interface{}) {
				payloadBytes, _ := json.Marshal(payload)
				signature := generateSignature(payloadBytes)
				s.On("CreateReceipt", mock.Anything, "invalid", "push", payloadBytes, signature).Return(nil, fmt.Errorf("invalid source: invalid")).Once()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new router and queue for each test case
			mockQueue := &queue.MockQueue{}
			handlers := NewHandlers(mockQueue, mockService)
			router := gin.New()
			router.POST("/api/webhooks/:source", handlers.HandleWebhook)

			// Setup mock expectations
			tc.setupMocks(mockService, mockQueue, tc.payload)

			// Create request
			payloadBytes, _ := json.Marshal(tc.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/webhooks/"+tc.source, bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Event-Type", tc.event)

			// Generate signature for valid requests
			if tc.name == "valid request" || tc.name == "invalid source" || tc.name == "invalid signature" {
				signature := generateSignature(payloadBytes)
				req.Header.Set("X-Signature", signature)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tc.wantStatus, w.Code)

			// Verify mock expectations
			mockQueue.AssertExpectations(t)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandleGetJobResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockQueue := &queue.MockQueue{}
	mockRepo := webhook.NewMockRepository()
	webhookService := webhook.NewService(mockRepo)
	handlers := NewHandlers(mockQueue, webhookService)

	router := gin.New()
	router.GET("/jobs/:id", handlers.HandleGetJobResult)

	tests := []struct {
		name       string
		jobID      string
		result     *models.JobResult
		err        error
		wantStatus int
	}{
		{
			name:  "existing job",
			jobID: "test-job-id",
			result: &models.JobResult{
				ID:     "test-job-id",
				Status: models.JobStatusCompleted,
				Result: "test result",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existent job",
			jobID:      "non-existent",
			result:     nil,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock expectations
			mockQueue.On("GetJobResult", mock.Anything, tt.jobID).Return(tt.result, tt.err).Once()

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/jobs/"+tt.jobID, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.wantStatus, w.Code)

			// Verify mock expectations
			mockQueue.AssertExpectations(t)
		})
	}
}

func TestHandleWebSocket(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockQueue := &queue.MockQueue{}
	mockRepo := webhook.NewMockRepository()
	webhookService := webhook.NewService(mockRepo)
	handlers := NewHandlers(mockQueue, webhookService)

	// Start the WebSocket server
	go handlers.wsServer.Start()
	defer handlers.wsServer.Stop()

	router := gin.New()
	router.GET("/ws", handlers.HandleWebSocket)

	// Create a test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?job_id=test-job-id"

	// Connect to WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer ws.Close()

	// Create a channel to receive WebSocket messages
	messages := make(chan internalws.JobStatus)
	go func() {
		for {
			var message internalws.JobStatus
			err := ws.ReadJSON(&message)
			if err != nil {
				close(messages)
				return
			}
			messages <- message
		}
	}()

	// Test job status updates
	statusUpdates := []struct {
		status string
		result interface{}
	}{
		{status: "pending", result: nil},
		{status: "running", result: nil},
		{status: "completed", result: "test result"},
	}

	for _, update := range statusUpdates {
		// Send status update
		handlers.NotifyJobStatus("test-job-id", update.status, update.result)

		// Wait for message
		select {
		case msg := <-messages:
			assert.Equal(t, "job_status", msg.Type)
			assert.Equal(t, "test-job-id", msg.JobID)
			assert.Equal(t, update.status, msg.Status)
			if update.result != nil {
				assert.Equal(t, update.result, msg.Result)
			}
		case <-time.After(time.Second):
			t.Fatalf("Timeout waiting for status update: %s", update.status)
		}
	}
}

// Helper function to generate a signature
func generateSignature(payload []byte) string {
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
