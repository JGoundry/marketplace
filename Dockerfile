FROM golang:alpine

# Set working dir
WORKDIR /app

# Downlaod GO modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./

# Build with static linking for linux
RUN CGO_ENABLED=0 GOOS=linux go build -o /marketplace

# Ports
EXPOSE 3000

# Run
CMD [ "/marketplace" ]