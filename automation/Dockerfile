# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webhook-server ./cmd/webhook-server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates nodejs npm git github-cli wget

# Install Claude CLI
RUN npm install -g @anthropic-ai/claude-code

# Create non-root user for running the application
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/webhook-server .
COPY --from=builder /build/config ./config
COPY --from=builder /build/migrations ./migrations

# Copy repo management scripts
COPY scripts/ ./scripts/
RUN chmod +x ./scripts/*.sh

# Create work directory for Claude sessions and set ownership
RUN mkdir -p /tmp/claude-sessions && \
    chown -R appuser:appuser /app /tmp/claude-sessions

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["./webhook-server"]
