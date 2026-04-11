# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Cache dependencies before copying full source
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o main cmd/main.go

# Final stage
FROM scratch

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy SSL certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set environment variables
ENV TZ=Asia/Jakarta

# Run the binary
ENTRYPOINT ["/app/main"] 