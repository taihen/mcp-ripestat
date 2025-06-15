.PHONY: all build clean lint test docker release

all: build

build:
	@echo "Building mcp-ripestat..."
	@go build -o mcp-ripestat ./cmd/mcp-ripestat

clean:
	@echo "Cleaning up..."
	@rm -f mcp-ripestat

lint:
	@echo "Linting..."
	@echo "Note: This runs linters via GitHub Actions. Run 'make test' for local checks."
	@golangci-lint run

test:
	@echo "Running tests..."
	@go test -v ./...

docker:
	@echo "Building Docker image..."
	@echo "Note: Docker build is handled by the release CI workflow."
	@# docker build -t ghcr.io/taihen/mcp-ripestat:latest .

release:
	@echo "Creating a release..."
	@echo "Note: Releases are handled automatically by CI on push to main."
	@echo "Commit with a conventional commit message to trigger a release."

help:
	@echo "Available targets:"
	@echo "  all      - Build the binary (default)"
	@echo "  build    - Build the binary"
	@echo "  clean    - Remove the built binary"
	@echo "  lint     - Run linters (primarily via CI)"
	@echo "  test     - Run unit tests"
	@echo "  docker   - Placeholder for Docker image build"
	@echo "  release  - Placeholder for release process" 