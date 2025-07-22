FROM golang:1.24.5-alpine AS builder
WORKDIR /app
COPY go.mod ./
# There are no dependencies in go.mod, so we don't need to run `go mod download` here.
# RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=${VERSION}" -o /app/bin/mcp-ripestat ./cmd/mcp-ripestat
FROM ghcr.io/taihen/base-image:v2025.07.21
WORKDIR /app
COPY --from=builder /app/bin/mcp-ripestat /app/mcp-ripestat
EXPOSE 8080
ENTRYPOINT ["/app/mcp-ripestat"]
