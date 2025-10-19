# Build stage
FROM golang:1.25.3-alpine3.22 AS builder

# Set working dir
WORKDIR /app

# Downlaod GO modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./

# Build with static linking for linux
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/marketplace

# Runtime stage
FROM alpine:3.22.2

WORKDIR /app

# Add system group and user and set home to /app
RUN addgroup -S appgroup && adduser -S appuser -G appgroup -h /app

# Copy only the compiled binary from the 'builder' stage
COPY --from=builder --chown=appuser:appgroup /app/marketplace .

# Switch to appuser account
USER appuser

# Ports
EXPOSE 3000

# Run
ENTRYPOINT [ "/app/marketplace" ]

# Health check to verify the app is ready (requires a health endpoint)
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null http://localhost:3000/health || exit 1
