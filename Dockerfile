# Use the official Golang image
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod ./
COPY go.sum* ./ 2>/dev/null || true

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o news-cms .

# Create a smaller final image
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/news-cms /app/news-cms

# Create empty .env file if it doesn't exist
RUN touch /app/.env

# Copy .env if it exists (will be replaced at runtime if needed)
COPY .env* /app/ 2>/dev/null || true

# Copy CA certificate if it exists
COPY ca-cert.pem /app/ 2>/dev/null || true

# Expose the port
EXPOSE 8080

# Run the binary
CMD ["./news-cms"]
