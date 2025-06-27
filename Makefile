# Makefile for mcp-ripestat

# Get version from git tag, fallback to commit hash if no tag
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD)
BINARY_NAME ?= mcp-ripestat
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

.PHONY: all build build-cross test test-coverage check-coverage e2e-test lint clean run deps fmt help

# Default target
all: fmt lint test test-coverage e2e-test build

# Build the application
build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/mcp-ripestat

# Run tests
test:
	@echo "Running tests..."
	go mod download
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Build for cross-platform (used by CI)
build-cross:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH) version $(VERSION)..."
	@BINARY_NAME="$(BINARY_NAME)-$(GOOS)-$(GOARCH)"; \
	if [ "$(GOOS)" = "windows" ]; then \
		BINARY_NAME="$$BINARY_NAME.exe"; \
	fi; \
	go build -v -ldflags="-s -w -X main.version=$(VERSION)" -o "$$BINARY_NAME" ./cmd/mcp-ripestat

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Check coverage threshold
check-coverage:
	@COVERAGE=$$(go tool cover -func=coverage.txt | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $$COVERAGE%"; \
	if [ "$$(echo "$$COVERAGE < 90" | bc -l)" = "1" ]; then \
		echo "Error: Test coverage is below 90%"; \
		exit 1; \
	fi

# Run end-to-end tests
e2e-test:
	@echo "Running end-to-end tests..."
	go test -v -tags=e2e ./e2e/...

# Run linting
lint:
	@echo "Running linters..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -l -s -w .
	goimports -local github.com/taihen/mcp-ripestat -w .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./bin/$(BINARY_NAME)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Help target
help:
	@echo "Available targets:"
	@echo "  all           - Run lint, test, and build"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  e2e-test      - Run end-to-end tests"
	@echo "  lint          - Run linters"
	@echo "  fmt           - Format code"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Build and run the application"
	@echo "  deps          - Install dependencies"
	@echo "  build-cross   - Build for cross-platform (used by CI)"
	@echo "  check-coverage- Check if coverage meets threshold"
	@echo "  help          - Show this help message"
