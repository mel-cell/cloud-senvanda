FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy Go modules manifests
COPY go.mod go.sum ./

# Download modules
RUN go mod download

# Copy source code
COPY . .

# Build
# CGO_ENABLED=0 untuk binary static (ngirit size & kompatibel)
RUN CGO_ENABLED=0 GOOS=linux go build -o /pocketbase cmd/api/main.go

# --- Final Stage ---
FROM alpine:latest

WORKDIR /app

# Install basic tools (optional, for debugging)
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /pocketbase /app/pocketbase

# Expose port
EXPOSE 8090

# Command to run
CMD ["/app/pocketbase", "serve", "--http=0.0.0.0:8090"]
