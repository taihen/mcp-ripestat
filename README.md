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

## Features

- Network information for IP addresses and prefixes
- AS overview for Autonomous System numbers
- Announced prefixes for Autonomous Systems
- Routing status for IP prefixes
- Whois information for IP addresses, prefixes, and ASNs
- Abuse contact finder for IP addresses and prefixes
- RPKI validation status for ASN and prefix combinations
- ASN neighbours for Autonomous Systems (upstream/downstream relationships)
- Looking Glass data for IP prefixes (BGP routing information from RIPE RIS)
- What's My IP functionality with proxy header support for IP detection

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

# Disable What's My IP endpoint (for shared environments this is undesirable)
./bin/mcp-ripestat --disable-whats-my-ip

# Show help
./bin/mcp-ripestat --help
```

### Proxy Support

The What's My IP endpoint (`/whats-my-ip`) automatically detects the real client
IP address when the server is behind a proxy or load balancer. It supports the
following proxy headers:

- `X-Forwarded-For` - Standard proxy header (uses first IP in chain)
- `X-Real-IP` - Alternative proxy header
- `CF-Connecting-IP` - Cloudflare-specific header

When running behind a proxy or Cloudflare tunnel, the endpoint will
automatically use these headers to determine the actual client IP address
instead of the proxy's IP.

### Shared Environment Configuration

For shared team environments, you should disable the `What's My IP` endpoint
using the `--disable-whats-my-ip` flag. This prevents team members from seeing
each server IP addresses when using a shared MCP server instance.

## MCP Protocol Support

This server implements the Model Context Protocol (MCP) 2025 specification with
JSON-RPC 2.0 transport. It provides two interfaces:

### JSON-RPC 2.0 Endpoint (Recommended)
- **Endpoint**: `/mcp`
- **Protocol**: MCP 2025 JSON-RPC 2.0
- **Usage**: Compatible with Cursor IDE and other MCP clients
- **Features**: Full MCP handshake, capability negotiation, tool calling

### Legacy REST API
- **Endpoints**: Individual REST endpoints (e.g., `/network-info`)
- **Protocol**: HTTP REST with query parameters
- **Usage**: Direct API access and backward compatibility

## MCP Client Configuration

To use this MCP server locally, simply copy and paste the
[MCP client configuration](./mcp.json) into your MCP client (e.g. for Cursor,
place it in `~/.cursor/mcp.json` on macOS/Linux).

## Example MCP Queries

Once configured, you can ask your AI assistant natural language questions that
will be translated into RIPEstat API calls. Here are some example queries you
can use:

### Network Information & IP Analysis

- "What network information can you tell me about the IP address 8.8.8.8?"
- "Analyze the network details for the prefix 193.0.0.0/21"
- "Show me the network information for 2001:db8::/32"

### Autonomous System (AS) Investigation

- "Give me an overview of AS3333"
- "What prefixes are announced by AS15169?"
- "Show me the routing status for the prefix 8.8.8.0/24"
- "Who are the upstream and downstream neighbors of AS1205?"

### Security & Abuse Investigation

- "Find the abuse contact information for IP address 192.0.2.1"
- "What's the abuse contact for the network containing 203.0.113.50?"

### WHOIS & Registration Data

- "Show me WHOIS information for AS64512"
- "Get WHOIS data for the IP address 198.51.100.1"
- "What's the WHOIS information for the prefix 2001:db8::/32?"

### RPKI Validation

- "Validate RPKI status for AS3333 announcing prefix 193.0.0.0/21"
- "Check if AS15169 is authorized to announce 8.8.8.0/24"

### BGP & Routing Analysis

- "Show me BGP routing information for prefix 193.0.0.0/21 from RIPE's looking glass"
- "Get looking glass data for 2001:7fb::/32 with 24-hour history"

### IP Detection & Connectivity

- "What's my public IP address?"
- "Detect my current IP and show network information about it"

### Complex Network Analysis Queries

- "Analyze the network infrastructure behind cloudflare.com - show me the IP,
  AS information, and any related prefixes"
- "Investigate potential abuse from this IP range: 203.0.113.0/24 - show network
  info, WHOIS, and abuse contacts"
- "Compare the network paths and AS relationships between Google's DNS (8.8.8.8)
  and Cloudflare's DNS (1.1.1.1)"
- "Perform a comprehensive security analysis of AS13335 including announced
  prefixes, neighbors, and RPKI validation status"

## API Endpoints

- `/network-info` - Get network information for an IP address or prefix
- `/as-overview` - Get an overview of an Autonomous System (AS)
- `/announced-prefixes` - Get a list of prefixes announced by an Autonomous System (AS)
- `/routing-status` - Get the routing status for an IP prefix
- `/whois` - Get whois information for an IP address, prefix, or ASN
- `/abuse-contact-finder` - Get abuse contact information for an IP address or prefix
- `/rpki-validation` - Get RPKI validation status for a resource (ASN) and prefix combination
- `/asn-neighbours` - Get ASN neighbours for an Autonomous System (upstream/downstream relationships)
- `/looking-glass` - Get Looking Glass data for IP prefixes (BGP routing information from RIPE RIS)
- `/whats-my-ip` - Get the caller's public IP address with proxy header support

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
  - `lookingglass` - Looking Glass data
  - `client` - HTTP client for RIPEstat API
  - `config` - Configuration handling
  - `errors` - Error types and handling
  - `logging` - Logging utilities
  - `networkinfo` - Network information data
  - `routingstatus` - Routing status data
  - `rpkivalidation` - RPKI validation status data
  - `types` - Common type definitions
  - `util` - Utility functions for IP, ASN validation, string manipulation, etc.
  - `whatsmyip` - What's My IP functionality with proxy header support
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
