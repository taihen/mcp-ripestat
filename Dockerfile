FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/mcp-ripestat ./cmd/mcp-ripestat

# Use taihen/base-image for the final stage
FROM ghcr.io/taihen/base-image:v2025.06.24

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/mcp-ripestat /app/mcp-ripestat

# Expose the default port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/mcp-ripestat"]
