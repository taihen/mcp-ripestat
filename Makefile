# Makefile for mcp-ripestat

.PHONY: all build test test-coverage e2e-test lint clean run deps fmt help

# Default target
all: lint test build

# Build the application
build:
	@echo "Building mcp-ripestat..."
	go build -o bin/mcp-ripestat ./cmd/mcp-ripestat

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//' | awk '{if ($$1 < 90) {print "Test coverage is below 90%"; exit 1} else {print "Test coverage is", $$1"%"}}'

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
	@echo "Running mcp-ripestat..."
	./bin/mcp-ripestat

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
	@echo "  help          - Show this help message"
