# Bespin Worker Service

The Bespin Worker Service is responsible for processing jobs from the job queue. It is designed to be scalable and can be run as multiple instances to handle increased workload.

## Features

- Processes jobs from Redis-based job queue
- Supports multiple job types:
  - Random text generation
  - Webhook processing
- Scalable design for handling large workloads
- Real-time job status updates via Redis pub/sub

## Architecture

The worker service follows a simple, modular architecture:

- `cmd/worker`: Main entry point
- `internal/queue`: Queue management and job processing
- `internal/jobs`: Job type implementations
- `pkg/models`: Shared data models

## Configuration

The service can be configured using environment variables:

- `REDIS_ADDR`: Redis server address (default: "localhost:6379")

## Development

### Prerequisites

- Go 1.21 or higher
- Redis

### Building

```bash
# Build the worker
go build -o bin/worker ./cmd/worker
```

### Running

```bash
# Run with default configuration
./bin/worker

# Run with custom Redis address
REDIS_ADDR=redis:6379 ./bin/worker
```

### Docker

```bash
# Build Docker image
docker build -t bespin-worker .

# Run Docker container
docker run -e REDIS_ADDR=redis:6379 bespin-worker
```

## Testing

```bash
# Run tests
go test ./...
```

## Deployment

The worker service is designed to be deployed in a containerized environment. It can be scaled horizontally by running multiple instances, which will automatically distribute the workload.

### Scaling Considerations

- Each worker instance processes jobs independently
- Redis handles job distribution and locking
- Workers can be added or removed without disruption
- Consider monitoring queue length for auto-scaling
