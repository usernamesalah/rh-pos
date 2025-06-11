# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata git

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main cmd/main.go

# Final stage
FROM scratch

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy necessary files from builder
COPY --from=builder /app/migrations ./migrations

# Copy SSL certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/app/main"] 