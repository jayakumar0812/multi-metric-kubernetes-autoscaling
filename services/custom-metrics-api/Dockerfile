# Multi-stage build for Custom Metrics API
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git
RUN apk add --no-cache git

# Copy go mod files
COPY src/go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY src/ ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o custom-metrics-api .

# Final stage
FROM alpine:latest

# Install ca-certificates and openssl for TLS support
RUN apk --no-cache add ca-certificates openssl

WORKDIR /root/

# Create directory for TLS certificates
RUN mkdir -p /etc/certs && chown -R root:root /etc/certs

# Copy TLS certificates from local path
COPY tls.crt /etc/certs/
COPY tls.key /etc/certs/

# Copy the binary from builder stage
COPY --from=builder /app/custom-metrics-api .

# Create non-root user (but keep certs accessible)
RUN adduser -D -s /bin/sh metrics

# Change ownership of the binary to metrics user
RUN chown metrics:metrics ./custom-metrics-api

# Make sure metrics user can read certificates
RUN chmod 755 /etc/certs

USER metrics

# Expose HTTPS port
EXPOSE 8443

# Health check using HTTPS with certificate skip
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-check-certificate --no-verbose --tries=1 --spider https://localhost:8443/healthz || exit 1

# Set environment variables for HTTPS
ENV PORT=8443
ENV TLS_CERT_FILE=/etc/certs/tls.crt
ENV TLS_KEY_FILE=/etc/certs/tls.key

# Run the binary
CMD ["./custom-metrics-api"]