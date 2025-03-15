.PHONY: all dev docker clean build build-api build-web test help deps lint

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m

# Default target
all: help

# Help target
help:
	@echo "Bespin - Cloud Job Processing Platform"
	@echo ""
	@echo "Usage:"
	@echo "  make dev         - Start Bespin in development mode (local)"
	@echo "  make docker      - Start Bespin using Docker Compose"
	@echo "  make build       - Build all components"
	@echo "  make build-api   - Build the API server"
	@echo "  make build-web   - Build the web client"
	@echo "  make test        - Run tests for all components"
	@echo "  make test-api    - Run API tests"
	@echo "  make test-web    - Run web tests"
	@echo "  make deps        - Install dependencies for all components"
	@echo "  make lint        - Run linters for all components"
	@echo "  make clean       - Clean up resources"
	@echo "  make docker-clean - Stop and remove all Docker containers"
	@echo ""

# Check for required tools
check-tools:
	@echo "${YELLOW}Checking for required tools...${NC}"
	@which docker > /dev/null || (echo "${RED}Docker is not installed. Please install Docker first.${NC}" && exit 1)
	@which docker-compose > /dev/null || (echo "${RED}Docker Compose is not installed. Please install Docker Compose first.${NC}" && exit 1)
	@which go > /dev/null || (echo "${RED}Go is not installed. Please install Go first.${NC}" && exit 1)
	@which pnpm > /dev/null || (echo "${RED}pnpm is not installed. Please install pnpm first.${NC}" && exit 1)

# Start the application in development mode (local)
dev: check-tools build-api
	@echo "${GREEN}Starting Bespin in development mode...${NC}"
	@echo "${YELLOW}Starting Redis...${NC}"
	@docker run -d --name bespin-redis -p 6379:6379 redis:alpine || (echo "${RED}Failed to start Redis. It might be already running.${NC}" && docker start bespin-redis || true)
	@echo "${YELLOW}Starting API server...${NC}"
	@cd api && ./bin/bespin-api & echo $$! > ../.api.pid
	@echo "${YELLOW}Starting web client...${NC}"
	@cd web && pnpm run dev & echo $$! > ../.web.pid
	@echo "${GREEN}Bespin is running! Press Ctrl+C to stop.${NC}"
	@echo "${GREEN}API: http://localhost:3002${NC}"
	@echo "${GREEN}Web: http://localhost:8000${NC}"
	@trap 'make clean' INT; while [ -f .api.pid ] && [ -f .web.pid ]; do sleep 1; done

# Start the application using Docker Compose
docker: check-tools
	@echo "${GREEN}Starting Bespin using Docker Compose...${NC}"
	@docker-compose up -d
	@echo "${GREEN}Bespin is running in Docker!${NC}"
	@echo "${GREEN}API: http://localhost:3002${NC}"
	@echo "${GREEN}Web: http://localhost:8000${NC}"

# Build all components
build: build-api build-web

# Build the API server
build-api:
	@echo "${YELLOW}Building API server...${NC}"
	@cd api && make build
	@echo "${GREEN}API server built successfully: api/bin/bespin-api${NC}"

# Build the web client
build-web:
	@echo "${YELLOW}Building web client...${NC}"
	@cd web && pnpm run build
	@echo "${GREEN}Web client built successfully${NC}"

# Run tests for all components
test: test-api test-web

# Run API tests
test-api:
	@echo "${YELLOW}Running API tests...${NC}"
	@cd api && make test

# Run web tests
test-web:
	@echo "${YELLOW}Running web tests...${NC}"
	@cd web && pnpm run test

# Install dependencies for all components
deps: deps-api deps-web

# Install API dependencies
deps-api:
	@echo "${YELLOW}Installing API dependencies...${NC}"
	@cd api && make deps

# Install web dependencies
deps-web:
	@echo "${YELLOW}Installing web dependencies...${NC}"
	@cd web && pnpm install

# Run linters for all components
lint: lint-api lint-web

# Run API linter
lint-api:
	@echo "${YELLOW}Running API linter...${NC}"
	@cd api && make lint

# Run web linter
lint-web:
	@echo "${YELLOW}Running web linter...${NC}"
	@cd web && pnpm run lint

# Clean up resources
clean:
	@echo "${YELLOW}Stopping Bespin...${NC}"
	@if [ -f .api.pid ]; then kill $$(cat .api.pid) 2>/dev/null || true; rm .api.pid; fi
	@if [ -f .web.pid ]; then kill $$(cat .web.pid) 2>/dev/null || true; rm .web.pid; fi
	@docker stop bespin-redis 2>/dev/null || true
	@docker rm bespin-redis 2>/dev/null || true
	@echo "${GREEN}Bespin stopped.${NC}"

# Stop and remove all Docker containers
docker-clean: check-tools
	@echo "${YELLOW}Stopping and removing all Docker containers...${NC}"
	@docker-compose down
	@echo "${GREEN}All Docker containers stopped and removed.${NC}"
