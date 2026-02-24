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
RUN apk --no-cache add ca-certificates nodejs npm git github-cli

# Install Claude CLI
RUN npm install -g @anthropic-ai/claude-code

# Create app user
RUN addgroup -S app && adduser -S app -G app

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/webhook-server .
COPY --from=builder /build/config ./config
COPY --from=builder /build/migrations ./migrations

# Create work directory for Claude sessions
RUN mkdir -p /tmp/claude-sessions && chown -R app:app /tmp/claude-sessions

# Change ownership
RUN chown -R app:app /app

# Switch to app user
USER app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["./webhook-server"]
