# Bespin

A cloud job processing platform.

## Overview

Bespin is a job processing system that allows clients to create and monitor jobs through a REST API and WebSocket connections. The system consists of:

- A Go-based API server for job processing
- A Vue.js web client for interacting with the API

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

1. Start Redis:
   ```bash
   docker run -d -p 6379:6379 redis:alpine
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

## WebSocket Events

The WebSocket endpoint sends job update events in the following format:

```json
{
  "id": "20230101120000.000000000",
  "status": "completed",
  "result": "Random text result...",
  "error": "",
  "updated_at": "2023-01-01T12:00:00Z"
}
```

## License

MIT
