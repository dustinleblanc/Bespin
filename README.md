# Bespin

A cloud job processing platform.

## Overview

Bespin is a job processing system that allows clients to create and monitor jobs through a REST API and WebSocket connections. The system consists of:

- A Go-based API server for job processing
- A Vue.js web client for interacting with the API
- Webhook support for integrating with external services
- PostgreSQL database for persistent storage using GORM

## Project Structure

```
bespin/
├── api/                  # Go API server
│   ├── cmd/              # Application entry points
│   ├── internal/         # Private application code
│   ├── pkg/              # Public libraries
│   └── bin/              # Compiled binaries
├── web/                  # Vue.js web client (Nuxt)
├── docker-compose.yml    # Docker Compose configuration
├── Makefile              # Build and run commands
└── run.sh                # Shell script for running the application
```

## Prerequisites

- Go 1.21 or higher
- Node.js 16 or higher
- pnpm
- Docker and Docker Compose
- PostgreSQL (included in Docker Compose setup)

## Testing Philosophy

Bespin follows a comprehensive testing strategy across both its API and web frontend:

### API Testing

- Focus on testing public interfaces and service boundaries
- Mock implementations for external dependencies
- Emphasis on behavior verification over implementation details
- Currently tested components:
  - API handlers
  - Webhook system
  - WebSocket server

### Web Frontend Testing

- Component-driven testing using Vitest and Testing Library
- Focus on user interactions and accessibility
- Real-world use case scenarios
- Store testing for state management
- Integration tests for component interactions

## Running Tests

```bash
# Run all tests (API and web)
make test

# Run API tests only
cd api && go test ./... -v

# Run web tests once and exit
cd web && pnpm test

# Run web tests in watch mode (for development)
cd web && pnpm test:watch

# Run web tests with UI
cd web && pnpm test:ui

# Generate web test coverage
cd web && pnpm test:coverage
```

## Running the Application

### Using the Makefile

The easiest way to run Bespin is using the provided Makefile:

```bash
# Run in development mode (local Go and Nuxt servers)
make dev

# Run using Docker Compose
make docker

# Build all components
make build

# Build the API server
make build-api

# Build the web client
make build-web

# Run tests
make test

# Clean up resources
make clean

# Stop and remove all Docker containers
make docker-clean

# Show help
make help
```

### Using the Shell Script

Alternatively, you can use the provided shell script:

```bash
# Make the script executable
chmod +x run.sh

# Run in development mode (local Go and Nuxt servers)
./run.sh dev

# Run using Docker Compose
./run.sh docker

# Clean up resources
./run.sh clean

# Stop and remove all Docker containers
./run.sh docker-clean

# Show help
./run.sh help
```

### Manual Setup

1. Start Redis and PostgreSQL:
   ```bash
   docker run -d -p 6379:6379 redis:alpine
   docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=bespin postgres:14-alpine
   ```

2. Start the API server:
   ```bash
   cd api
   make build
   ./bin/bespin-api
   ```

3. Start the web client:
   ```bash
   cd web
   pnpm dev
   ```

4. The API will be available at http://localhost:3002
5. The web client will be available at http://localhost:8000

## Features

- RESTful API for job creation and management
- Real-time job status updates via WebSockets
- Redis-based job queue for reliable job processing
- PostgreSQL database for persistent storage using GORM ORM
- Webhook support for integrating with external services
- Random text generation job type
- Modern Vue.js web interface

## API Endpoints

- `GET /api/` - Root endpoint, returns API information
- `GET /api/test` - Test endpoint
- `GET /api/jobs/test` - Job service test endpoint
- `POST /api/jobs/random-text` - Create a random text job
  - Query parameters:
    - `length` (optional) - Length of the random text to generate (default: 100)
- `GET /api/ws/jobs` - WebSocket endpoint for job updates
- `POST /api/webhooks/:source` - Receive webhooks from external services
- `GET /api/webhooks/:id` - Get a specific webhook receipt
- `GET /api/webhooks` - List webhook receipts

## WebSocket Events

The WebSocket endpoint sends job update events in the following format:

```json
{
  "type": "job_status",
  "job_id": "string",
  "status": "string", // pending, running, completed, failed
  "result": "any"     // optional result data
}
```

### Connecting to WebSocket

Connect to the WebSocket endpoint with a job ID:

```javascript
const ws = new WebSocket(`ws://localhost:3002/api/ws?job_id=${jobId}`);

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log(`Job ${data.job_id} status: ${data.status}`);
  if (data.result) {
    console.log(`Result: ${data.result}`);
  }
};
```

### WebSocket Features

- **Job-Specific Updates**: Each client subscribes to updates for a specific job
- **Status History**: New clients receive the latest status upon connection
- **Efficient Broadcasting**: Messages are filtered by job ID
- **Connection Management**: Automatic handling of connection lifecycle
- **Future Features**: Support for team and site-wide broadcasts planned

## Webhook Support

Bespin includes a robust webhook system that allows external services to trigger events in the application. Webhooks are received, verified, and stored in PostgreSQL using GORM.

### Webhook Endpoints

- `POST /api/webhooks/:source` - Receive webhooks from external services
  - URL parameters:
    - `source` - The source of the webhook (e.g., "github", "stripe", "test")
  - Headers:
    - `X-Webhook-Signature` - HMAC signature for verification
    - `X-Webhook-Event` (optional) - Event type
  - Body:
    - JSON payload with at least an `event` field

### Webhook Verification

Webhooks are verified using HMAC-SHA256 signatures. The signature is calculated using a secret key specific to each webhook source.

### Webhook Storage

Webhook receipts are stored in PostgreSQL using GORM, providing persistent storage and easy querying capabilities.

## Database

Bespin uses PostgreSQL for persistent storage with GORM as the ORM layer. This provides:

- Type-safe database operations
- Automatic migrations
- Relationship management
- Query building
- Transaction support

## License

MIT
