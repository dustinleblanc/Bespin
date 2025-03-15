# Bespin Cloud Platform

A modern cloud development platform with real-time job processing capabilities.

## Features

- NestJS API with Socket.IO integration
- Bull queue for background job processing
- Vue.js frontend with Tailwind CSS
- Docker-based development environment

## Architecture

The application consists of the following components:

- **API Server**: NestJS application that handles HTTP requests and WebSocket connections
- **Job Processing**: Integrated Bull queue processor for background jobs
- **Web Client**: Vue.js application with real-time updates via Socket.IO
- **Infrastructure**: Redis for job queuing and PostgreSQL for data storage

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Node.js 20+
- pnpm

### Development

1. Clone the repository
2. Install dependencies:
   ```
   pnpm install
   ```
3. Start the development environment:
   ```
   pnpm dev
   ```

This will start all services in development mode with hot reloading.

### API Endpoints

- `GET /api`: Root endpoint
- `GET /api/test`: Test endpoint
- `GET /api/jobs/test`: Job service test endpoint
- `POST /api/jobs/random-text`: Create a new random text generation job

### WebSocket Events

- `job-completed:{jobId}`: Emitted when a job is completed

## License

This project is licensed under the MIT License - see the LICENSE file for details.
