# Use the official Golang image
FROM golang:1.21-alpine

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy the .env file
COPY .env ./

# Copy CA certificate for database connection
COPY ca-cert.pem ./ca-cert.pem

# Copy the rest of the source code
COPY . .

# Build the application
RUN go build -o news-cms .

# Expose the port defined in .env (8080)
EXPOSE 8080

# Run the binary
CMD ["./news-cms"]
