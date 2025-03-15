package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dustinleblanc/go-bespin/internal/queue"
	"github.com/dustinleblanc/go-bespin/internal/webhook"
	"github.com/dustinleblanc/go-bespin/internal/websocket"
	"github.com/dustinleblanc/go-bespin/pkg/models"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestReceiveWebhook(t *testing.T) {
	// Skip if Redis is not available
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	_, err := redisClient.Ping(redisClient.Context()).Result()
	if err != nil {
		t.Skip("Redis is not available")
	}

	// Create dependencies
	jobQueue := queue.NewJobQueue("localhost:6379")
	repository := webhook.NewRedisRepository(redisClient)
	webhookService := webhook.NewService(repository)
	wsServer := websocket.NewServer(jobQueue)
	handlers := NewHandlers(jobQueue, webhookService)
	router := SetupRouter(handlers, wsServer)
	factory := webhook.NewFactory()

	// Create a test payload
	payload := map[string]interface{}{
		"event": "test-event",
		"data":  "test-data",
	}
	payloadBytes, _ := json.Marshal(payload)

	// Generate a signature
	signature := factory.GenerateSignature("test", payloadBytes)

	// Create a test request
	req, _ := http.NewRequest("POST", "/api/webhooks/test", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var response models.WebhookResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.ID)
	assert.True(t, response.Verified)

	// Clean up
	redisClient.Del(redisClient.Context(), "webhook:"+response.ID)
	redisClient.LRem(redisClient.Context(), "webhooks:test", 0, response.ID)
	redisClient.LRem(redisClient.Context(), "webhooks:all", 0, response.ID)
}

func TestGetWebhook(t *testing.T) {
	// Skip if Redis is not available
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := redisClient.Context()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis is not available")
	}

	// Create dependencies
	jobQueue := queue.NewJobQueue("localhost:6379")
	repository := webhook.NewRedisRepository(redisClient)
	webhookService := webhook.NewService(repository)
	wsServer := websocket.NewServer(jobQueue)
	handlers := NewHandlers(jobQueue, webhookService)
	router := SetupRouter(handlers, wsServer)
	factory := webhook.NewFactory()

	// Create and store a webhook receipt
	receipt := factory.CreateWebhookReceipt(
		"test",
		"test-event",
		map[string]interface{}{"data": "test"},
	)
	err = webhookService.StoreWebhook(ctx, receipt)
	assert.NoError(t, err)

	// Create a test request
	req, _ := http.NewRequest("GET", "/api/webhooks/"+receipt.ID, nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var retrievedReceipt models.WebhookReceipt
	err = json.Unmarshal(w.Body.Bytes(), &retrievedReceipt)
	assert.NoError(t, err)
	assert.Equal(t, receipt.ID, retrievedReceipt.ID)
	assert.Equal(t, receipt.Source, retrievedReceipt.Source)
	assert.Equal(t, receipt.Event, retrievedReceipt.Event)
	assert.Equal(t, receipt.Signature, retrievedReceipt.Signature)
	assert.Equal(t, receipt.Verified, retrievedReceipt.Verified)

	// Clean up
	redisClient.Del(ctx, "webhook:"+receipt.ID)
	redisClient.LRem(ctx, "webhooks:test", 0, receipt.ID)
	redisClient.LRem(ctx, "webhooks:all", 0, receipt.ID)
}

func TestListWebhooks(t *testing.T) {
	// Skip if Redis is not available
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := redisClient.Context()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis is not available")
	}

	// Create dependencies
	jobQueue := queue.NewJobQueue("localhost:6379")
	repository := webhook.NewRedisRepository(redisClient)
	webhookService := webhook.NewService(repository)
	wsServer := websocket.NewServer(jobQueue)
	handlers := NewHandlers(jobQueue, webhookService)
	router := SetupRouter(handlers, wsServer)
	factory := webhook.NewFactory()

	// Create and store multiple webhook receipts
	receipts := make([]*models.WebhookReceipt, 3)
	for i := 0; i < 3; i++ {
		receipt := factory.CreateWebhookReceipt(
			"test",
			"test-event",
			map[string]interface{}{"data": "test"},
		)
		err = webhookService.StoreWebhook(ctx, receipt)
		assert.NoError(t, err)
		receipts[i] = receipt
	}

	// Create a test request
	req, _ := http.NewRequest("GET", "/api/webhooks?source=test", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var response struct {
		Webhooks []*models.WebhookReceipt `json:"webhooks"`
		Count    int                      `json:"count"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, response.Count, 3)

	// Clean up
	for _, receipt := range receipts {
		redisClient.Del(ctx, "webhook:"+receipt.ID)
		redisClient.LRem(ctx, "webhooks:test", 0, receipt.ID)
		redisClient.LRem(ctx, "webhooks:all", 0, receipt.ID)
	}
}

// Helper function to generate a signature
func generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
