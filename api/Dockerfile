FROM golang:1.24-alpine

WORKDIR /app

# Copy the source code
COPY . .

# Download dependencies and build the application
RUN go mod download && \
    go mod tidy && \
    mkdir -p bin && \
    go build -o bin/bespin-api ./cmd/api

# Set the working directory to where the binary is
WORKDIR /app/bin

# Expose the port
EXPOSE 3002

# Run the application
CMD ["./bespin-api"]
