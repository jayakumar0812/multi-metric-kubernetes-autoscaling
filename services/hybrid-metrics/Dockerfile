# CREATE: services/hybrid-metrics/Dockerfile
# Dockerfile for your hybrid metrics server with correct structure

FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for go mod download
RUN apk add --no-cache git

# Copy go mod files (assuming they're in the root of hybrid-metrics)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy all source code
COPY . .

# Build the binary from cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hybrid-metrics-server ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/hybrid-metrics-server .

# Create non-root user
RUN adduser -D -s /bin/sh metrics

# Change ownership
RUN chown metrics:metrics /root/hybrid-metrics-server
USER metrics

# Expose ports (adjust based on your server configuration)
EXPOSE 8080 3001

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3001/health || exit 1

# Run the binary
CMD ["./hybrid-metrics"]