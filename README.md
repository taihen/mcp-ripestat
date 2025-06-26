# MCP RIPEstat

[![CI/CD](https://github.com/taihen/mcp-ripestat/actions/workflows/ci.yml/badge.svg)](https://github.com/taihen/mcp-ripestat/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/taihen/mcp-ripestat)](https://goreportcard.com/report/github.com/taihen/mcp-ripestat)

A Model Context Protocol (MCP) server for the RIPEstat Data API, providing
network information for IP addresses and prefixes.

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

### Investigation Workflows

For examples, investigation workflows, and usage patterns, see [PROMPTS](PROMPTS.md).

## Features

**Core Network Analysis:**

- Network information for IP addresses and prefixes
- AS overview for Autonomous System numbers
- Whois information for IP addresses, prefixes, and ASNs

**Routing Intelligence:**

- Announced prefixes for Autonomous Systems
- Routing status for IP prefixes
- Routing history for IP addresses, prefixes, and ASNs (historical BGP visibility data)
- ASN neighbours for Autonomous Systems (upstream/downstream relationships)
- Looking Glass data for IP prefixes (BGP routing information from RIPE RIS)

**Security & Compliance:**

- RPKI validation status for ASN and prefix combinations
- Abuse contact finder for IP addresses and prefixes

**Utility:**

- Health check and warmup endpoints for monitoring and deployment

## Architectural Rationale

`mcp-ripestat` is implemented as an MCP HTTP (MCP 2025 compatible) server rather
than a command-line interface (stdio) tool to facilitate centralized deployment
and access control.

This architecture allows the service to be installed on a single instance and
be accessed by many users and other MCP clients across a network in
private mode - not exposed to the internet.

> [!WARNING]
> At current stage this MCP server does not provide authentication.
> Use firewall or other L3 networking to restrict access to the server.

## Installation

### Prerequisites

- Go 1.24.4 or higher
- Make

> [!INFO]
> No External Dependencies: This project uses only Go standard library.

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

# Run the server
./bin/mcp-ripestat

# Show help
./bin/mcp-ripestat --help
```

### Health Check Endpoints

The server provides essential monitoring endpoints:

- `/status` - Server status with uptime, version, and health information
- `/warmup` - Warmup endpoint to prevent cold starts in containerized deployments

These endpoints are essential for load balancers, monitoring systems, and deployment orchestration.

## MCP Protocol Support

This server implements the Model Context Protocol (MCP) 2025 specification with
JSON-RPC 2.0 transport. It provides two interfaces:

### JSON-RPC 2.0 Endpoint (Recommended)

- **Endpoint**: `/mcp`
- **Protocol**: MCP 2025 JSON-RPC 2.0
- **Usage**: Compatible with Cursor IDE and other MCP clients
- **Features**: Full MCP handshake, capability negotiation, tool calling

### Legacy REST API (Removed in v2.0.0)

> [!FAIL] 
> **BREAKING CHANGE**: All legacy REST API endpoints have been removed in v2.0.0. Use the MCP JSON-RPC endpoint instead.

This is a major breaking change. Legacy REST endpoints have been completely removed to keep the codebase compact and focused on the MCP protocol. All functionality previously available through REST endpoints is now accessible through the `/mcp` endpoint using the MCP protocol.

## MCP Client Configuration

To use this MCP server locally, simply copy and paste the
[MCP client configuration](./mcp.json) into your MCP client (e.g. for Cursor,
place it in `~/.cursor/mcp.json` on macOS/Linux).

### Demo Server

A demo MCP server is running at `https://mcp-ripestat.taihen.org/mcp`. Feel free to try it out, but there are no uptime promises.

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

## Contributing

Contributions are welcome! Please read [contributing guidelines](CONTRIBUTING.md)
to see how you can participate.

## Development

The development process is organized into sprints, with each sprint focused on a
specific feature. A detailed ledger of all the past and feature changes,
divided by sprints, is available in the [sprints](SPRINTS.md) documentation.

> [!NOTE]
> If a particular feature is not included, please feel free to open
> [issue](https://github.com/taihen/mcp-ripestat/issues?q=sort%3Aupdated-desc+is%3Aissue+is%3Aopen)
> to discuss it.

## License

This project is licensed under the MIT License. See the [license](LICENSE) file
for details.
