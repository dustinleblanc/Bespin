package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/dustinleblanc/go-bespin/pkg/models"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestVerifySignature(t *testing.T) {
	// Create a Redis repository
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	repository := NewRedisRepository(redisClient)

	// Create a webhook service and factory
	service := NewService(repository)
	factory := NewFactory()

	// Test cases
	testCases := []struct {
		name      string
		source    string
		payload   []byte
		signature string
		expected  bool
	}{
		{
			name:      "Valid signature",
			source:    "test",
			payload:   []byte(`{"event":"test","data":"test"}`),
			signature: factory.GenerateSignature("test", []byte(`{"event":"test","data":"test"}`)),
			expected:  true,
		},
		{
			name:      "Invalid signature",
			source:    "test",
			payload:   []byte(`{"event":"test","data":"test"}`),
			signature: "invalid-signature",
			expected:  false,
		},
		{
			name:      "Unknown source",
			source:    "unknown",
			payload:   []byte(`{"event":"test","data":"test"}`),
			signature: "signature",
			expected:  false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.VerifySignature(tc.source, tc.payload, tc.signature)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestStoreAndGetWebhook(t *testing.T) {
	// Skip if Redis is not available
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis is not available")
	}

	// Create a Redis repository
	repository := NewRedisRepository(redisClient)

	// Create a webhook service and factory
	service := NewService(repository)
	factory := NewFactory()

	// Create a webhook receipt
	receipt := factory.CreateWebhookReceipt(
		"test",
		"test-event",
		map[string]interface{}{"data": "test"},
	)

	// Store the webhook receipt
	err = service.StoreWebhook(ctx, receipt)
	assert.NoError(t, err)

	// Get the webhook receipt
	retrievedReceipt, err := service.GetWebhook(ctx, receipt.ID)
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
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis is not available")
	}

	// Create a Redis repository
	repository := NewRedisRepository(redisClient)

	// Create a webhook service and factory
	service := NewService(repository)
	factory := NewFactory()

	// Create and store multiple webhook receipts
	receipts := make([]*models.WebhookReceipt, 3)
	for i := 0; i < 3; i++ {
		receipt := factory.CreateWebhookReceipt(
			"test",
			"test-event",
			map[string]interface{}{"data": "test"},
		)
		err = service.StoreWebhook(ctx, receipt)
		assert.NoError(t, err)
		receipts[i] = receipt
	}

	// List webhooks
	listedReceipts, err := service.ListWebhooks(ctx, "test", 10, 0)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(listedReceipts), 3)

	// Count webhooks
	count, err := service.CountWebhooks(ctx, "test")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, 3)

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
