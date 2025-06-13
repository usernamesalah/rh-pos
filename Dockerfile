# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy the rest of the source code
COPY . .

# Build the application with optimizations
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
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