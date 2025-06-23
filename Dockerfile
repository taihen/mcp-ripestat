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

# Use a minimal alpine image for the final stage
FROM alpine:3.19

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/mcp-ripestat /app/mcp-ripestat

# Expose the default port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/mcp-ripestat"]
