#!/bin/bash

# Colors for output
GREEN="\033[0;32m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
NC="\033[0m"

# Print a message with color
print_message() {
  echo -e "$1$2${NC}"
}

# Check if a command exists
check_command() {
  if ! command -v $1 &> /dev/null; then
    print_message "${RED}" "$1 is not installed. Please install $1 first."
    exit 1
  fi
}

# Check for required tools
check_tools() {
  print_message "${YELLOW}" "Checking for required tools..."
  check_command "docker"
  check_command "docker-compose"
  check_command "go"
  check_command "pnpm"
}

# Start the application in development mode (local)
start_dev() {
  check_tools

  print_message "${GREEN}" "Starting Bespin in development mode..."

  # Build the API server
  print_message "${YELLOW}" "Building API server..."
  cd api && make build && cd ..

  # Start Redis
  print_message "${YELLOW}" "Starting Redis..."
  docker run -d --name bespin-redis -p 6379:6379 redis:alpine || (print_message "${RED}" "Failed to start Redis. It might be already running." && docker start bespin-redis || true)

  # Start API server
  print_message "${YELLOW}" "Starting API server..."
  cd api && ./bin/bespin-api & echo $! > ../.api.pid
  cd ..

  # Start web client
  print_message "${YELLOW}" "Starting web client..."
  cd web && pnpm run dev & echo $! > ../.web.pid
  cd ..

  print_message "${GREEN}" "Bespin is running! Press Ctrl+C to stop."
  print_message "${GREEN}" "API: http://localhost:3002"
  print_message "${GREEN}" "Web: http://localhost:8000"

  # Set up trap to clean up on exit
  trap cleanup INT TERM

  # Wait for processes to finish
  wait
}

# Start the application using Docker Compose
start_docker() {
  check_tools

  print_message "${GREEN}" "Starting Bespin using Docker Compose..."
  docker-compose up -d

  print_message "${GREEN}" "Bespin is running in Docker!"
  print_message "${GREEN}" "API: http://localhost:3002"
  print_message "${GREEN}" "Web: http://localhost:8000"
}

# Clean up resources
cleanup() {
  print_message "${YELLOW}" "Stopping Bespin..."

  if [ -f .api.pid ]; then
    kill $(cat .api.pid) 2>/dev/null || true
    rm .api.pid
  fi

  if [ -f .web.pid ]; then
    kill $(cat .web.pid) 2>/dev/null || true
    rm .web.pid
  fi

  docker stop bespin-redis 2>/dev/null || true
  docker rm bespin-redis 2>/dev/null || true

  print_message "${GREEN}" "Bespin stopped."
  exit 0
}

# Stop and remove all Docker containers
docker_clean() {
  check_tools

  print_message "${YELLOW}" "Stopping and removing all Docker containers..."
  docker-compose down

  print_message "${GREEN}" "All Docker containers stopped and removed."
}

# Show usage information
show_usage() {
  echo "Bespin - Cloud Job Processing Platform"
  echo ""
  echo "Usage: ./run.sh [command]"
  echo ""
  echo "Commands:"
  echo "  dev         - Start Bespin in development mode (local)"
  echo "  docker      - Start Bespin using Docker Compose"
  echo "  clean       - Clean up resources"
  echo "  docker-clean - Stop and remove all Docker containers"
  echo "  help        - Show this help message"
  echo ""
}

# Main script logic
case "$1" in
  dev)
    start_dev
    ;;
  docker)
    start_docker
    ;;
  clean)
    cleanup
    ;;
  docker-clean)
    docker_clean
    ;;
  help|*)
    show_usage
    ;;
esac
