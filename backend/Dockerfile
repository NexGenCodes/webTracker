# Step 1: Build the Go binary
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -ldflags="-s -w" -o bot ./cmd/bot

# Step 2: Create the final lightweight image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Install runtime dependencies (ca-certificates for HTTPS/WhatsApp)
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/bot .

# Copy assets and sql (needed for migrations and rendering)
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/sql ./sql

# Create logs directory
RUN mkdir -p /app/logs

# Expose the API port
EXPOSE 5000

# Run the application
CMD ["./bot"]
