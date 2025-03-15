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
	"testing"

	"github.com/dustinleblanc/go-bespin/internal/queue"
	"github.com/dustinleblanc/go-bespin/internal/webhook"
	"github.com/dustinleblanc/go-bespin/internal/websocket"
	"github.com/dustinleblanc/go-bespin/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Set up test environment
	os.Setenv("GO_ENV", "test")
	os.Setenv("GITHUB_WEBHOOK_SECRET", "test-secret-for-testing")
}

func setupTestServer(t *testing.T) (*webhook.MockRepository, *gin.Engine) {
	// Set up test environment for each test
	os.Setenv("GO_ENV", "test")
	os.Setenv("GITHUB_WEBHOOK_SECRET", "test-secret-for-testing")

	jobQueue := &queue.MockJobQueue{}
	repository := webhook.NewMockRepository()
	webhookService := webhook.NewService(repository)
	wsServer := websocket.NewServer(jobQueue)
	handlers := NewHandlers(jobQueue, webhookService)
	router := SetupRouter(handlers, wsServer)
	return repository, router
}

func TestReceiveWebhook(t *testing.T) {
	testCases := []struct {
		name           string
		source         string
		payload        map[string]interface{}
		signature      string
		expectedStatus int
		expectedValid  bool
		expectJobQueue bool
	}{
		{
			name:   "Valid webhook",
			source: "github",
			payload: map[string]interface{}{
				"event": "test-event",
				"data":  "test-data",
			},
			signature:      "", // Will be generated
			expectedStatus: http.StatusOK,
			expectedValid:  true,
			expectJobQueue: true,
		},
		{
			name:   "Invalid signature",
			source: "github",
			payload: map[string]interface{}{
				"event": "test-event",
				"data":  "test-data",
			},
			signature:      "invalid-signature",
			expectedStatus: http.StatusUnauthorized,
			expectedValid:  false,
			expectJobQueue: false,
		},
		{
			name:   "Missing signature",
			source: "github",
			payload: map[string]interface{}{
				"event": "test-event",
				"data":  "test-data",
			},
			signature:      "",
			expectedStatus: http.StatusBadRequest,
			expectedValid:  false,
			expectJobQueue: false,
		},
		{
			name:   "Unknown source",
			source: "unknown",
			payload: map[string]interface{}{
				"event": "test-event",
				"data":  "test-data",
			},
			signature:      "", // Will be generated
			expectedStatus: http.StatusBadRequest,
			expectedValid:  false,
			expectJobQueue: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository, router := setupTestServer(t)
			factory := webhook.NewFactory()

			payloadBytes, _ := json.Marshal(tc.payload)

			// Generate signature if needed for the test case
			signature := tc.signature
			if signature == "" && tc.name != "Missing signature" {
				signature = factory.GenerateSignature(tc.source, payloadBytes)
			}

			req, _ := http.NewRequest("POST", "/api/webhooks/"+tc.source, bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			if signature != "" {
				req.Header.Set("X-Webhook-Signature", signature)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var response models.WebhookResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tc.expectedValid, response.Verified)

				// Verify storage
				stored, err := repository.GetByID(repository.Context(), response.ID)
				assert.NoError(t, err)
				assert.Equal(t, tc.source, stored.Source)
				assert.Equal(t, tc.payload["event"], stored.Event)

				// Verify job queueing
				if tc.expectJobQueue {
					assert.NotEmpty(t, response.JobID, "Expected a job ID for valid webhook")
					assert.Empty(t, response.Warning, "Expected no warning for successful job queueing")
				} else {
					assert.Empty(t, response.JobID, "Expected no job ID for invalid webhook")
				}
			}
		})
	}
}

// TestReceiveWebhookJobQueueFailure tests the behavior when job queueing fails
func TestReceiveWebhookJobQueueFailure(t *testing.T) {
	// Create a mock job queue that always fails
	failingJobQueue := &queue.MockJobQueue{
		AddJobFunc: func(jobType string, data interface{}) (string, error) {
			return "", fmt.Errorf("simulated job queue failure")
		},
	}

	// Set up the test server with the failing job queue
	repository := webhook.NewMockRepository()
	webhookService := webhook.NewService(repository)
	wsServer := websocket.NewServer(failingJobQueue)
	handlers := NewHandlers(failingJobQueue, webhookService)
	router := SetupRouter(handlers, wsServer)

	// Create a valid webhook request
	payload := map[string]interface{}{
		"event": "test-event",
		"data":  "test-data",
	}
	payloadBytes, _ := json.Marshal(payload)
	factory := webhook.NewFactory()
	signature := factory.GenerateSignature("github", payloadBytes)

	// Send the request
	req, _ := http.NewRequest("POST", "/api/webhooks/github", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.WebhookResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.ID)
	assert.True(t, response.Verified)
	assert.Empty(t, response.JobID)
	assert.NotEmpty(t, response.Warning)
	assert.Contains(t, response.Warning, "processing job could not be queued")

	// Verify webhook was still stored
	stored, err := repository.GetByID(repository.Context(), response.ID)
	assert.NoError(t, err)
	assert.Equal(t, "github", stored.Source)
	assert.Equal(t, "test-event", stored.Event)
}

func TestGetWebhook(t *testing.T) {
	testCases := []struct {
		name           string
		setupWebhook   bool
		expectedStatus int
	}{
		{
			name:           "Existing webhook",
			setupWebhook:   true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-existent webhook",
			setupWebhook:   false,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository, router := setupTestServer(t)
			factory := webhook.NewFactory()

			var webhookID string
			if tc.setupWebhook {
				receipt := factory.CreateWebhookReceipt(
					"github",
					"test-event",
					map[string]interface{}{"data": "test"},
				)
				err := repository.Store(repository.Context(), receipt)
				assert.NoError(t, err)
				webhookID = receipt.ID
			} else {
				webhookID = "non-existent-id"
			}

			req, _ := http.NewRequest("GET", "/api/webhooks/"+webhookID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var retrievedReceipt models.WebhookReceipt
				err := json.Unmarshal(w.Body.Bytes(), &retrievedReceipt)
				assert.NoError(t, err)
				assert.Equal(t, webhookID, retrievedReceipt.ID)
			}
		})
	}
}

func TestListWebhooks(t *testing.T) {
	testCases := []struct {
		name           string
		source         string
		webhookCount   int
		expectedCount  int
		expectedStatus int
	}{
		{
			name:           "List all webhooks",
			source:         "",
			webhookCount:   3,
			expectedCount:  3,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "List source webhooks",
			source:         "github",
			webhookCount:   3,
			expectedCount:  3,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty source",
			source:         "empty",
			webhookCount:   0,
			expectedCount:  0,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository, router := setupTestServer(t)
			factory := webhook.NewFactory()

			// Create test webhooks
			for i := 0; i < tc.webhookCount; i++ {
				receipt := factory.CreateWebhookReceipt(
					"github",
					"test-event",
					map[string]interface{}{"data": "test"},
				)
				err := repository.Store(repository.Context(), receipt)
				assert.NoError(t, err)
			}

			url := "/api/webhooks"
			if tc.source != "" {
				url += "?source=" + tc.source
			}

			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var response struct {
				Webhooks []*models.WebhookReceipt `json:"webhooks"`
				Count    int                      `json:"count"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCount, response.Count)
			assert.Equal(t, tc.expectedCount, len(response.Webhooks))
		})
	}
}

// Helper function to generate a signature
func generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
