package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("GITHUB_WEBHOOK_SECRET", "test-secret")
}

func TestServiceWithMockRepository(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	// Helper function to generate valid signature
	generateSignature := func(payload []byte) string {
		secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
		h := hmac.New(sha256.New, []byte(secret))
		h.Write(payload)
		return hex.EncodeToString(h.Sum(nil))
	}

	t.Run("CreateReceipt", func(t *testing.T) {
		testCases := []struct {
			name       string
			source     string
			event      string
			payload    []byte
			signature  string
			wantErr    bool
			errMessage string
		}{
			{
				name:    "valid receipt",
				source:  "github",
				event:   "push",
				payload: []byte(`{"test": "data"}`),
				wantErr: false,
			},
			{
				name:       "missing event",
				source:     "github",
				event:      "",
				payload:    []byte(`{}`),
				wantErr:    true,
				errMessage: "event is required",
			},
			{
				name:       "invalid source",
				source:     "invalid",
				event:      "push",
				payload:    []byte(`{}`),
				wantErr:    true,
				errMessage: "invalid source: invalid",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Generate signature for valid cases
				signature := ""
				if tc.source == "github" && tc.event != "" {
					signature = generateSignature(tc.payload)
				}

				receipt, err := service.CreateReceipt(ctx, tc.source, tc.event, tc.payload, signature)

				if tc.wantErr {
					assert.Error(t, err)
					if tc.errMessage != "" {
						assert.Contains(t, err.Error(), tc.errMessage)
					}
					assert.Nil(t, receipt)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, receipt)
					assert.Equal(t, tc.source, receipt.Source)
					assert.Equal(t, tc.event, receipt.Event)
					assert.Equal(t, tc.payload, receipt.Payload)
					assert.Equal(t, signature, receipt.Signature)
				}
			})
		}
	})

	t.Run("GetReceipt", func(t *testing.T) {
		// Create a test receipt first
		payload := []byte(`{"test": "data"}`)
		signature := generateSignature(payload)
		receipt, err := service.CreateReceipt(ctx, "github", "push", payload, signature)
		assert.NoError(t, err)

		// Test getting the receipt
		found, err := service.GetReceipt(ctx, receipt.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, receipt.ID, found.ID)

		// Test getting non-existent receipt
		found, err = service.GetReceipt(ctx, "non-existent")
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("ListReceipts", func(t *testing.T) {
		// Create test receipts
		payload := []byte(`{"test": "data"}`)
		signature := generateSignature(payload)
		_, err := service.CreateReceipt(ctx, "github", "push", payload, signature)
		assert.NoError(t, err)
		_, err = service.CreateReceipt(ctx, "github", "pull_request", payload, signature)
		assert.NoError(t, err)

		// Test listing all receipts
		receipts, err := service.ListReceipts(ctx, "github", 10, 0)
		assert.NoError(t, err)
		assert.NotEmpty(t, receipts)

		// Test listing with invalid source
		receipts, err = service.ListReceipts(ctx, "invalid", 10, 0)
		assert.Error(t, err)
		assert.Empty(t, receipts)
	})

	t.Run("CountReceipts", func(t *testing.T) {
		// Test counting receipts for valid source
		count, err := service.CountReceipts(ctx, "github")
		assert.NoError(t, err)
		assert.Greater(t, count, int64(0))

		// Test counting receipts for invalid source
		count, err = service.CountReceipts(ctx, "invalid")
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
	})
}
