# Multi-stage build for Go
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for go mod download
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy all source code
COPY . .

# Build the binary - output name should match what local build creates
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Debug: List files to see what was created
RUN ls -la

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage (correct name: server)
COPY --from=builder /app/server .

# Make sure it's executable
RUN chmod +x server

# Create non-root user
RUN adduser -D -s /bin/sh metrics
RUN chown metrics:metrics /root/server
USER metrics

# Expose ports
EXPOSE 8080 3001

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3001/health || exit 1

# Run the binary (correct name: server)
CMD ["./server"]