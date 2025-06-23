# Contributing to MCP RIPEstat

Thank you for considering contributing to MCP RIPEstat! This document provides simple guidelines for contributing to this project.

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue on GitHub with a clear description, steps to reproduce, and relevant environment details.

### Suggesting Enhancements

For enhancement suggestions, create an issue with a clear description and any relevant examples.

### Pull Requests

1. Fork the repository
2. Create a new branch for your changes
3. Make your changes
4. Run tests and linters
5. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.24.4 or higher
- Make

```bash
# Clone the repository
git clone https://github.com/taihen/mcp-ripestat.git
cd mcp-ripestat

# Install dependencies
make deps

# Run tests
make test

# Run linters
make lint
```

## Coding Standards

- Follow standard Go code style and best practices
- Use `gofmt` and `goimports` to format your code
- Write tests for new functionality
- Document exported functions, types, and variables

## License

By contributing to this project, you agree that your contributions will be licensed under the project's [MIT License](LICENSE).
