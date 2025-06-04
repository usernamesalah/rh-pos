# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install Air for hot reload (using a specific version compatible with Go 1.23)
RUN go install github.com/cosmtrek/air@v1.49.0

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# Final stage
FROM golang:1.23-alpine

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .
COPY --from=builder /go/bin/air /usr/local/bin/air
COPY --from=builder /app/air.toml .
COPY --from=builder /app/go.mod .
COPY --from=builder /app/go.sum .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"] 