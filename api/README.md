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

### Database Migrations

Currently, the application uses GORM's AutoMigrate feature for database schema management. This automatically creates and updates database tables based on Go struct definitions. While this approach is convenient for development, it has some limitations:

- Limited control over migration process
- No explicit version control of schema changes
- May not handle complex migrations well

For production deployment, we plan to transition to a more robust migration system using tools like `golang-migrate` or `goose`. This will provide:

- Version-controlled database schema
- Explicit up/down migrations
- Better handling of complex schema changes
- SQL-first approach for precise control
- Integration with deployment processes

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

## WebSocket Server

The WebSocket server provides real-time job status updates to clients. It is built using the `melody` WebSocket framework and supports:

- Job-specific status notifications
- Multiple clients per job
- Status history for new connections
- Future support for team/site-wide broadcasts

### Features

- **Job-Specific Updates**: Each client subscribes to updates for a specific job
- **Status History**: New clients receive the latest status upon connection
- **Efficient Broadcasting**: Uses melody's filtered broadcasting for targeted updates
- **Connection Management**: Automatic handling of connection lifecycle and ping/pong
- **Origin Control**: Configurable origin checking for security

### WebSocket Endpoint

- `GET /api/ws` - WebSocket endpoint for job updates
  - Query parameters:
    - `job_id` - ID of the job to subscribe to
  - Messages:
    ```json
    {
      "type": "job_status",
      "job_id": "string",
      "status": "string", // pending, running, completed, failed
      "result": "any"     // optional result data
    }
    ```

### Example Usage

```javascript
// Connect to WebSocket
const ws = new WebSocket(`ws://localhost:3002/api/ws?job_id=${jobId}`);

// Handle messages
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log(`Job ${data.job_id} status: ${data.status}`);
  if (data.result) {
    console.log(`Result: ${data.result}`);
  }
};
```

### Implementation Details

The WebSocket server uses the `melody` framework for efficient WebSocket handling:

- Connection management is handled automatically
- Messages are filtered by job ID using `BroadcastFilter`
- Status history is maintained for each job
- Connection lifecycle events (connect, disconnect) are logged
- Future support for team/site-wide broadcasts is planned

### Security

- Origin checking is configurable via `CheckOrigin` function
- Job IDs are validated before establishing connections
- Messages are filtered to ensure clients only receive updates for their subscribed job

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

### Testing Philosophy

Our testing strategy focuses on ensuring the reliability of our public interfaces and service boundaries while avoiding over-testing of internal implementation details. This approach:

- **Tests Public Interfaces**: We prioritize testing the APIs and endpoints that other services depend on
- **Tests Service Boundaries**: We verify the correct interaction between our service and external dependencies (databases, message queues, etc.)
- **Avoids Internal Implementation Testing**: We deliberately avoid testing internal packages that don't directly interface with external services
- **Uses Mocks Appropriately**: We use mock implementations for testing to avoid external dependencies and focus on behavior verification

This philosophy helps us:
- Maintain test suite efficiency
- Reduce test maintenance burden
- Focus coverage on critical paths
- Allow for internal refactoring without breaking tests
- Ensure stability of our public contracts

Currently tested components:
- API handlers (public HTTP endpoints)
- Webhook system (external service integration)
- WebSocket server (real-time client communication)

Internal implementation details in packages like `internal/database`, `internal/queue`, and `internal/jobs` are intentionally not covered by tests as they are implementation details that may change.

### Docker

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run
```
