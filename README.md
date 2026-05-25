# MCP RIPEstat

[![CI/CD](https://github.com/taihen/mcp-ripestat/actions/workflows/ci.yml/badge.svg)](https://github.com/taihen/mcp-ripestat/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/taihen/mcp-ripestat)](https://goreportcard.com/report/github.com/taihen/mcp-ripestat)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/taihen/mcp-ripestat/badge)](https://scorecard.dev/viewer/?uri=github.com/taihen/mcp-ripestat)

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

This MCP server provides intelligent access to RIPEstat Data API through **7 directly accessible tools**:

**Consolidated Semantic Tools** (6 tools that automatically detect resource types and route requests):
- **🔍 investigateResource** - Primary investigation tool with auto-detection and intelligent routing
- **📊 analyzeRouting** - BGP and routing analysis with timeframe support  
- **📋 queryRegistry** - Registry and administrative data lookup
- **🔒 validateSecurity** - Security and compliance checks (RPKI, abuse contacts)
- **🌐 exploreRelationships** - Network topology and relationships
- **🌍 searchByLocation** - Geographic analysis by country

**Utility Tool** (1 tool):
- **🔧 getWhatsMyIP** - Get caller's public IP address

> [!NOTE]
> Only these 7 tools are directly callable. Individual RIPEstat endpoints (like `getNetworkInfo`, `getASOverview`, etc.) are internal implementation details and are automatically called by the consolidated tools based on resource type and requested operations.

**Resource Detection**: Automatically identifies IP addresses, prefixes, ASNs, and country codes without requiring explicit type specification.

**Operation Routing**: Maps semantic operations to the most relevant RIPEstat endpoints based on resource type and context.

See [ENDPOINT_PARITY](ENDPOINT_PARITY.md) for implementation status of underlying RIPEstat endpoints.

## Architectural Rationale

**HTTP Transport Choice**: This server implements MCP over **Streamable HTTP**
rather than stdio transport, enabling deployment as a standalone network
service that multiple MCP clients can access concurrently without process
spawning overhead.

**Legacy Protocol Fallback**: The server maintains backwards compatibility with
the deprecated transports.

**Concurrent Request Management**: Semaphore-based rate limiting operates
per-server instance rather than per-client-connection, managing RIPE API quotas
across multiple concurrent sessions.

> [!WARNING]
> At current stage this MCP server does not provide authentication. The initial
> version of MCP released on 2024-11-05 did not support authorization. However,
> in the 2025-03-26 update, the MCP protocol introduced an OAuth 2.1-based
> authorization mechanism. It is still not widely adopted and might be subject
> to change in the future - therefore, we recommend using a firewall or other
> native MCP solution, such as MCPProxy, to restrict access to the server.

## Installation

### Prerequisites

- Go 1.26.2 or higher
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

### Health Check Endpoints

The server provides essential monitoring endpoints:

- `/status` - Server status with uptime, version, and health information
- `/warmup` - Warmup endpoint to prevent cold starts in containerized deployments

These endpoints are essential for load balancers, monitoring systems, and deployment orchestration.

## MCP Protocol Support

### Streamable HTTP Transport

• Endpoint: /mcp (streaming occurs over the same request/response channel)
• Protocol: Stream-framed HTTP (per MCP spec 2025-06-18)
• Status: Default transport for all MCP clients implementing the 2025-06-18 spec
• Features: Bidirectional streaming, incremental responses, zero-copy frames

### JSON-RPC 2.0 Endpoint

• Endpoint: /mcp
• Protocol: JSON-RPC 2.0
• Status: Recommended production endpoint (replaces REST)
• Features: Full MCP handshake, capability negotiation, tool invocation, compatible with Cursor IDE and other MCP-compliant clients

## MCP Client Configuration

To use this MCP server locally, simply copy and paste the
[MCP client configuration](./mcp.json) into your MCP client

- **Cursor**: macOS/Linux: `~/.cursor/mcp.json`
- **Claude Code (local server)**: `claude mcp add --transport http ripestat http://localhost:8080/mcp`
- **Codex CLI (local server)**: `codex mcp add ripestat --url http://localhost:8080/mcp`

### Demo Server

A demo MCP server is running at `https://mcp-ripestat.taihen.org/mcp`. Feel
free to try it out, but there are no uptime promises.

#### Add Demo Server from CLI

Use user scope to make the server available across projects.

**Claude Code:**

```bash
claude mcp add-json ripestat --scope user '{"type":"http","url":"https://mcp-ripestat.taihen.org/mcp"}'
```

**Codex CLI:**

```bash
codex mcp add ripestat --url https://mcp-ripestat.taihen.org/mcp
```

Verify configuration:

```bash
claude mcp list
codex mcp list
```

#### Using the Demo Server with Codex Config File

Codex reads MCP server definitions from `~/.codex/config.toml` (global) or
`.codex/config.toml` (project-local), not from `mcp-demo.json`.

Add this server entry:

```toml
[mcp_servers.ripestat]
url = "https://mcp-ripestat.taihen.org/mcp"
```

Then restart Codex and run `/mcp` to verify that `ripestat` is connected.

You can now ask prompts such as:

- `Investigate 8.8.8.8 and summarize ownership, routing, and abuse contacts.`
- `Analyze routing for AS3333 and show key anomalies.`

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

## Container Images

Release builds publish multi-platform container images to GitHub Container Registry:

```bash
docker pull ghcr.io/taihen/mcp-ripestat:<version>
docker pull ghcr.io/taihen/mcp-ripestat:latest
```

Published images include BuildKit provenance and SBOM attestations. Release images
also receive a signed GitHub Artifact Attestation for SLSA build provenance,
attached to the image digest in GHCR.

Verify the provenance attestation for a release image with GitHub CLI:

```bash
gh attestation verify oci://ghcr.io/taihen/mcp-ripestat:<version> \
  --repo taihen/mcp-ripestat
```

The release workflow targets SLSA Build Level 2 for container images by building
on GitHub-hosted Actions and publishing signed provenance for the resulting image
digest.

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
