# Start from golang alpine base for smallest footprint
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for downloading dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/gateway

# Start a new smaller stage
FROM alpine:latest

WORKDIR /app

# Add certificates for HTTPS (OpenRouter API)
RUN apk --no-cache add ca-certificates tzdata

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

# Expose port 8080 and 9090
EXPOSE 8080 9090

# Run the executable
CMD ["./main"]
