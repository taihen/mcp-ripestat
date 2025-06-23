# MCP RIPEstat

[![Lint](https://github.com/taihen/mcp-ripestat/actions/workflows/lint.yml/badge.svg)](https://github.com/taihen/ppp-exporter/actions/workflows/lint.yml)
[![Test](https://github.com/taihen/mcp-ripestat/actions/workflows/test.yml/badge.svg)](https://github.com/taihen/ppp-exporter/actions/workflows/test.yml)
[![Build](https://github.com/taihen/mcp-ripestat/actions/workflows/build.yml/badge.svg)](https://github.com/taihen/ppp-exporter/actions/workflows/build.yml)
[![Release](https://github.com/taihen/mcp-ripestat/actions/workflows/release.yml/badge.svg)](https://github.com/taihen/mcp-ripestat/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/taihen/mcp-ripestat)](https://goreportcard.com/report/github.com/taihen/mcp-ripestat)


A Model Context Protocol (MCP) server for the RIPEstat Data API, providing network information for IP addresses and prefixes.

> [!CAUTION]
> This MCP server is currently under active development. It may not be stable
> and is subject to change. Please use with caution.

## About RIPEstat

RIPEstat is a large-scale information service and the RIPE NCC's open data
platform. It provides essential data on IP address space and Autonomous System
Numbers (ASNs), along with related statistics for specific hostnames and
countries. This service is a valuable tool for network operators, security
researchers, and anyone interested in the structure and performance of the
Internet, offering insights into routing, registration data, DNS, and
geographical information.

For more information, visit the [RIPEstat website](https://stat.ripe.net/) and
consult the [Data API documentation](https://stat.ripe.net/docs/data_api).

## Use Cases

Using MCP Client allows you to leverage a Large Language Model (LLM) to query
RIPEStat using natural language. This means you can ask complex questions about
network information, IP addresses, ASNs, and other RIPEStat data without needing
to know the underlying API or query syntax. The client translates your natural
language queries into appropriate API calls, making it easier and more intuitive
to access and analyze RIPEStat data for research, troubleshooting, or network
management tasks.

## Features

- Network information for IP addresses and prefixes
- AS overview for Autonomous System numbers
- Announced prefixes for Autonomous Systems
- Routing status for IP prefixes
- Whois information for IP addresses, prefixes, and ASNs
- Abuse contact finder for IP addresses and prefixes
- RPKI validation status for ASN and prefix combinations
- ASN neighbours for Autonomous Systems (upstream/downstream relationships)

## Architectural Rationale

`mcp-ripestat` is implemented as an HTTP server rather than a command-line
interface (CLI) tool to facilitate centralized deployment. This architecture
allows the service to be installed on a single server and accessed by multiple
team members and other MCP clients across a network. It promotes reusability,
simplifies client-side configuration, and provides a single, consistent access
point for RIPEstat data.

> [!WARNING]
> At current stage this MCP server does not provide authentication.

## Installation

### Prerequisites

- Go 1.24.4 or higher
- Make

```bash
# Clone the repository
git clone https://github.com/taihen/mcp-ripestat.git
cd mcp-ripestat

# Install dependencies
make deps

# Build the application
make build
```

## Usage

```bash
# Run the server
./bin/mcp-ripestat

# Run with custom port
./bin/mcp-ripestat --port 8081

# Enable debug logging
./bin/mcp-ripestat --debug

# Show help
./bin/mcp-ripestat --help
```

## MCP Client Configuration

To use this MCP server, simply copy and paste the [MCP client configuration](./mcp.json) into your MCP client (e.g. for Cursor, place it in `~/.cursor/mcp.json` on macOS/Linux).

## API Endpoints

- `/network-info` - Get network information for an IP address or prefix
- `/as-overview` - Get an overview of an Autonomous System (AS)
- `/announced-prefixes` - Get a list of prefixes announced by an Autonomous System (AS)
- `/routing-status` - Get the routing status for an IP prefix
- `/whois` - Get whois information for an IP address, prefix, or ASN
- `/abuse-contact-finder` - Get abuse contact information for an IP address or prefix
- `/rpki-validation` - Get RPKI validation status for a resource (ASN) and prefix combination
- `/asn-neighbours` - Get ASN neighbours for an Autonomous System (upstream/downstream relationships)

## Testing

```bash
# Run tests
make test

# Run tests with coverage report
make test-coverage

# Run end-to-end tests
make e2e-test

# Run linters
make lint

# Format code
make fmt

# Clean build artifacts
make clean

# Install dependencies
make deps
```

### Project Structure

The project is organized into the following packages:

- `cmd/mcp-ripestat` - Main application entry point
- `internal/ripestat` - Core functionality
  - `abusecontactfinder` - Abuse contact finder data
  - `announcedprefixes` - Announced prefixes data
  - `asnneighbours` - ASN neighbours data
  - `asoverview` - AS overview data
  - `client` - HTTP client for RIPEstat API
  - `config` - Configuration handling
  - `errors` - Error types and handling
  - `logging` - Logging utilities
  - `networkinfo` - Network information data
  - `routingstatus` - Routing status data
  - `rpkivalidation` - RPKI validation status data
  - `types` - Common type definitions
  - `util` - Utility functions for IP, ASN validation, string manipulation, etc.
  - `whois` - Whois information data

## Contributing

Contributions are welcome! Please read [contributing guidelines](CONTRIBUTING.md)
to see how you can participate.

## Development

The development process is organized into sprints, with each sprint focused on a
specific feature. A detailed ledger of all the past and feature changes,
divided by sprints, is available in the [sprints](.github/SPRINTS.md)
documentation.

> [!NOTE]
> If a particular feature is not included, please feel free to open
> [issue](https://github.com/taihen/mcp-ripestat/issues?q=sort%3Aupdated-desc+is%3Aissue+is%3Aopen)
> to discuss it.

## License

This project is licensed under the MIT License. See the [license](LICENSE) file
for details.
