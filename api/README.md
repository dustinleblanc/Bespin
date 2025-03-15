# Bespin API

The Bespin API is a Go-based server that provides job processing, webhook handling, and WebSocket communication capabilities.

## Architecture

The API is structured using a clean architecture approach:

- `cmd/` - Application entry points
- `internal/` - Private application code
  - `api/` - API handlers and routing
  - `database/` - Database connections and migrations
  - `jobs/` - Job processing logic
  - `queue/` - Job queue implementation
  - `webhook/` - Webhook handling
  - `websocket/` - WebSocket server
- `pkg/` - Public libraries and models
- `bin/` - Compiled binaries

## Technologies

- **Go** - Programming language
- **Gin** - Web framework
- **GORM** - Object-Relational Mapping for PostgreSQL
- **Redis** - For job queue and caching
- **WebSockets** - For real-time communication
- **PostgreSQL** - For persistent storage

## Database

The API uses PostgreSQL for persistent storage with GORM as the ORM layer. This provides:

- Type-safe database operations
- Automatic migrations
- Relationship management
- Query building
- Transaction support

### Database Configuration

Database connection parameters are configured through environment variables:

- `DB_HOST` - PostgreSQL host (default: "localhost")
- `DB_PORT` - PostgreSQL port (default: "5432")
- `DB_USER` - PostgreSQL user (default: "postgres")
- `DB_PASSWORD` - PostgreSQL password (default: "postgres")
- `DB_NAME` - PostgreSQL database name (default: "bespin")

### Models

The main database models include:

- `WebhookReceipt` - Stores received webhooks

## Webhook System

The webhook system allows external services to trigger events in the application. Webhooks are received, verified, and stored in PostgreSQL using GORM.

### Webhook Flow

1. External service sends a webhook to `/api/webhooks/:source`
2. API verifies the webhook signature
3. Webhook receipt is created and stored in PostgreSQL
4. API returns a response with the webhook receipt ID

### Webhook Sources

The system supports multiple webhook sources, each with its own secret key for signature verification:

- `github` - GitHub webhooks
- `stripe` - Stripe webhooks
- `sendgrid` - SendGrid webhooks
- `test` - Test webhooks (for development and testing)

### Webhook Verification

Webhooks are verified using HMAC-SHA256 signatures. The signature is calculated using a secret key specific to each webhook source.

Secret keys are configured through environment variables:

- `GITHUB_WEBHOOK_SECRET` - GitHub webhook secret
- `STRIPE_WEBHOOK_SECRET` - Stripe webhook secret
- `SENDGRID_WEBHOOK_SECRET` - SendGrid webhook secret

For testing, a default secret of "test-secret" is used for the "test" source.

### Webhook Storage

Webhook receipts are stored in PostgreSQL using GORM. The `WebhookReceipt` model includes:

- `ID` - Unique identifier (UUID)
- `Source` - Webhook source (e.g., "github", "stripe")
- `Event` - Event type
- `Payload` - JSON payload
- `Headers` - HTTP headers
- `Signature` - HMAC signature
- `Verified` - Whether the signature was verified
- `CreatedAt` - Timestamp

### Webhook Endpoints

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

## Job System

The job system allows clients to create and monitor jobs through a REST API and WebSocket connections.

### Job Types

- `random-text` - Generates random text

### Job Endpoints

- `POST /api/jobs/random-text` - Create a random text job
  - Query parameters:
    - `length` (optional) - Length of the random text to generate (default: 100)

- `GET /api/ws/jobs` - WebSocket endpoint for job updates

## Development

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL
- Redis

### Running the API

```bash
# Build the API
make build

# Run the API
./bin/bespin-api
```

### Environment Variables

- `PORT` - API port (default: "3002")
- `REDIS_ADDR` - Redis address (default: "localhost:6379")
- `DB_HOST` - PostgreSQL host (default: "localhost")
- `DB_PORT` - PostgreSQL port (default: "5432")
- `DB_USER` - PostgreSQL user (default: "postgres")
- `DB_PASSWORD` - PostgreSQL password (default: "postgres")
- `DB_NAME` - PostgreSQL database name (default: "bespin")
- `GITHUB_WEBHOOK_SECRET` - GitHub webhook secret
- `STRIPE_WEBHOOK_SECRET` - Stripe webhook secret
- `SENDGRID_WEBHOOK_SECRET` - SendGrid webhook secret

### Testing

```bash
# Run tests
make test
```

### Docker

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run
```
