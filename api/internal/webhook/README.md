# Webhook System

The webhook system allows external services to trigger events in the Bespin application. Webhooks are received, verified, and stored in PostgreSQL using GORM.

## Architecture

The webhook system is structured using a clean architecture approach:

- `service.go` - Webhook service for business logic
- `repository.go` - Repository interface for storage operations
- `gorm_repository.go` - GORM implementation of the repository interface
- `factory.go` - Factory for creating test webhook receipts

## Components

### Repository Interface

The `Repository` interface defines the contract for webhook storage operations:

```go
type Repository interface {
    Store(ctx context.Context, receipt *models.WebhookReceipt) error
    GetByID(ctx context.Context, id string) (*models.WebhookReceipt, error)
    List(ctx context.Context, source string, limit, offset int) ([]*models.WebhookReceipt, error)
    Count(ctx context.Context, source string) (int, error)
}
```

### GORM Repository

The `GormRepository` implements the `Repository` interface using GORM:

```go
type GormRepository struct {
    db     *database.GormDB
    logger *log.Logger
}
```

It provides methods for storing, retrieving, listing, and counting webhook receipts in PostgreSQL.

### Service

The `Service` handles webhook operations:

```go
type Service struct {
    repository Repository
    logger     *log.Logger
    secrets    map[string]string
}
```

It provides methods for verifying webhook signatures and managing webhook receipts.

### Factory

The `Factory` provides methods for creating test webhook receipts:

```go
type Factory struct {
    secrets map[string]string
}
```

It's useful for testing and development.

## Webhook Flow

1. External service sends a webhook to `/api/webhooks/:source`
2. API handler reads the request body and headers
3. Webhook service verifies the signature using the source-specific secret
4. Webhook receipt is created with the payload, headers, signature, and verification status
5. Webhook receipt is stored in PostgreSQL using GORM
6. API returns a response with the webhook receipt ID, verification status, and timestamp

## Webhook Sources

The system supports multiple webhook sources, each with its own secret key for signature verification:

- `github` - GitHub webhooks
- `stripe` - Stripe webhooks
- `sendgrid` - SendGrid webhooks
- `test` - Test webhooks (for development and testing)

## Webhook Verification

Webhooks are verified using HMAC-SHA256 signatures. The signature is calculated using a secret key specific to each webhook source.

Secret keys are configured through environment variables:

- `GITHUB_WEBHOOK_SECRET` - GitHub webhook secret
- `STRIPE_WEBHOOK_SECRET` - Stripe webhook secret
- `SENDGRID_WEBHOOK_SECRET` - SendGrid webhook secret

For testing, a default secret of "test-secret" is used for the "test" source.

## Webhook Storage

Webhook receipts are stored in PostgreSQL using GORM. The `WebhookReceipt` model includes:

- `ID` - Unique identifier (UUID)
- `Source` - Webhook source (e.g., "github", "stripe")
- `Event` - Event type
- `Payload` - JSON payload (stored as JSONB in PostgreSQL)
- `Headers` - HTTP headers (stored as JSONB in PostgreSQL)
- `Signature` - HMAC signature
- `Verified` - Whether the signature was verified
- `CreatedAt` - Timestamp

## Usage

### Creating a Webhook Service

```go
// Create a GORM database connection
db, err := database.NewGormDB()
if err != nil {
    log.Fatalf("Failed to connect to database: %v", err)
}

// Auto migrate the webhook receipt model
if err := db.AutoMigrate(&models.WebhookReceipt{}); err != nil {
    log.Fatalf("Failed to run auto migrations: %v", err)
}

// Create a GORM repository
repo := webhook.NewGormRepository(db)

// Create a webhook service
service := webhook.NewService(repo)
```

### Verifying a Webhook

```go
// Verify a webhook signature
verified := service.VerifySignature(source, payload, signature)
```

### Storing a Webhook Receipt

```go
// Create a webhook receipt
receipt := models.NewWebhookReceipt(source, event, payload, headers, signature, verified)

// Store the webhook receipt
err := service.StoreWebhook(ctx, receipt)
```

### Retrieving a Webhook Receipt

```go
// Get a webhook receipt by ID
receipt, err := service.GetWebhook(ctx, id)
```

### Listing Webhook Receipts

```go
// List webhook receipts by source
receipts, err := service.ListWebhooks(ctx, source, limit, offset)
```

### Counting Webhook Receipts

```go
// Count webhook receipts by source
count, err := service.CountWebhooks(ctx, source)
```

## Testing

The webhook system includes a factory for creating test webhook receipts:

```go
// Create a webhook factory
factory := webhook.NewFactory()

// Create a test webhook receipt
receipt := factory.CreateWebhookReceipt(source, event, payload)

// Create a GitHub webhook receipt
githubReceipt := factory.CreateGithubWebhook(event)

// Create a Stripe webhook receipt
stripeReceipt := factory.CreateStripeWebhook(event)
```

## API Endpoints

- `POST /api/webhooks/:source` - Receive webhooks from external services
  - URL parameters:
    - `source` - The source of the webhook (e.g., "github", "stripe", "test")
  - Headers:
    - `X-Webhook-Signature` - HMAC signature for verification
    - `X-Webhook-Event` (optional) - Event type
  - Body:
    - JSON payload with at least an `event` field

- `GET /api/webhooks/:id` - Get a specific webhook receipt
  - URL parameters:
    - `id` - Webhook receipt ID

- `GET /api/webhooks` - List webhook receipts
  - Query parameters:
    - `source` (optional) - Filter by source
