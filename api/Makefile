.PHONY: build run test clean docker-build docker-run

# Build the application
build:
	go build -o bin/bespin-api ./cmd/api

# Run the application
run:
	go run ./cmd/api

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker-build:
	docker build -t bespin-api .

# Run Docker container
docker-run:
	docker run -p 3002:3002 --env REDIS_ADDR=host.docker.internal:6379 bespin-api

# Run Docker Compose
docker-compose-up:
	cd .. && docker-compose up

# Stop Docker Compose
docker-compose-down:
	cd .. && docker-compose down

# Install dependencies
deps:
	go mod tidy

# Run linter
lint:
	go vet ./...

# Default target
all: clean build
