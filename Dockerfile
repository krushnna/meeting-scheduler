# Use Golang Alpine as builder
FROM golang:1.24.1-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to leverage caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .


# Build the application binary
RUN CGO_ENABLED=0 GOOS=linux go build -o meeting-scheduler .

# Use a minimal alpine image for production
FROM alpine:3.18

# Install necessary dependencies
RUN apk --no-cache add ca-certificates tzdata bash

WORKDIR /root/

# Copy the compiled binary, .env file, and docs folder from builder stage
COPY --from=builder /app/meeting-scheduler .
COPY --from=builder /app/.env .
COPY --from=builder /app/docs ./docs

# Expose the application port
EXPOSE 8080

# Set the default command
CMD ["./meeting-scheduler"]
