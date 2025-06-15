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
	@# The following command is a placeholder for a real linting tool.
	@echo "Linting would run here."

test:
	@echo "Running unit tests..."
	@go test -v -tags=unit ./...

test-e2e:
	@echo "Running end-to-end tests..."
	@go test -v -tags=e2e ./...

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
	@echo "  test-e2e  - Run end-to-end tests"
	@echo "  docker   - Placeholder for Docker image build"
	@echo "  release  - Placeholder for release process"
