package webhook

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceWithMockRepository(t *testing.T) {
	// Create a mock repository
	repository := NewMockRepository()

	// Create a webhook service and factory
	service := NewService(repository)
	factory := NewFactory()

	// Create a webhook receipt
	receipt := factory.CreateWebhookReceipt(
		"test",
		"test-event",
		map[string]interface{}{"data": "test"},
	)

	// Test storing a webhook
	ctx := context.Background()
	err := service.StoreWebhook(ctx, receipt)
	assert.NoError(t, err)

	// Test getting a webhook
	retrievedReceipt, err := service.GetWebhook(ctx, receipt.ID)
	assert.NoError(t, err)
	assert.Equal(t, receipt.ID, retrievedReceipt.ID)
	assert.Equal(t, receipt.Source, retrievedReceipt.Source)
	assert.Equal(t, receipt.Event, retrievedReceipt.Event)
	assert.Equal(t, receipt.Signature, retrievedReceipt.Signature)
	assert.Equal(t, receipt.Verified, retrievedReceipt.Verified)

	// Test listing webhooks
	receipts, err := service.ListWebhooks(ctx, "test", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(receipts))
	assert.Equal(t, receipt.ID, receipts[0].ID)

	// Test counting webhooks
	count, err := service.CountWebhooks(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Test listing all webhooks
	allReceipts, err := service.ListWebhooks(ctx, "", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(allReceipts))

	// Test counting all webhooks
	allCount, err := service.CountWebhooks(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, 1, allCount)

	// Test listing with pagination
	paginatedReceipts, err := service.ListWebhooks(ctx, "test", 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(paginatedReceipts))

	// Test listing with non-existent source
	nonExistentReceipts, err := service.ListWebhooks(ctx, "non-existent", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(nonExistentReceipts))

	// Test counting with non-existent source
	nonExistentCount, err := service.CountWebhooks(ctx, "non-existent")
	assert.NoError(t, err)
	assert.Equal(t, 0, nonExistentCount)

	// Test getting a non-existent webhook
	_, err = service.GetWebhook(ctx, "non-existent")
	assert.Error(t, err)
}
