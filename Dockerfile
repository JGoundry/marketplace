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
RUN CGO_ENABLED=0 GOOS=linux go build -o /marketplace

# Runtime stage
FROM alpine:3.22.2

# Copy only the compiled binary from the 'builder' stage
COPY --from=builder /marketplace .

# Ports
EXPOSE 3000

# Run
CMD [ "/marketplace" ]
