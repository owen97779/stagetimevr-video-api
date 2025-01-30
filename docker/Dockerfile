# ==========================
# 1. Build Stage (Build Binary)
# ==========================
FROM golang:1.21 AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum (for dependency caching)
COPY ./ ./
RUN go mod download

# Build the application
RUN go build -o /app/main cmd/main.go

# ==========================
# 2. Final Stage (Run the App)
# ==========================
FROM debian:bookworm-slim

# Set working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Expose a port (if your app runs a web server)
EXPOSE 8080

# Run the application
CMD ["./main"]
