FROM golang:1.24-alpine

# Set up the workspace
WORKDIR /workspace

# Copy both modules
COPY worker /workspace/worker/
COPY api /workspace/api/

# Set working directory to the API module
WORKDIR /workspace/api

# Download dependencies and build the application
RUN go mod download && \
    go mod tidy && \
    mkdir -p bin && \
    go build -o bin/bespin-api ./cmd/api

# Set the working directory to where the binary is
WORKDIR /workspace/api/bin

# Expose the port
EXPOSE 3002

# Run the application
CMD ["./bespin-api"]
