# MCP-RIPEStat

[![Lint](https://github.com/taihen/mcp-ripestat/actions/workflows/lint.yml/badge.svg)](https://github.com/taihen/ppp-exporter/actions/workflows/lint.yml)
[![Test](https://github.com/taihen/mcp-ripestat/actions/workflows/test.yml/badge.svg)](https://github.com/taihen/ppp-exporter/actions/workflows/test.yml)
[![Build](https://github.com/taihen/mcp-ripestat/actions/workflows/build.yml/badge.svg)](https://github.com/taihen/ppp-exporter/actions/workflows/build.yml)
[![Release](https://github.com/taihen/mcp-ripestat/actions/workflows/release.yml/badge.svg)](https://github.com/taihen/mcp-ripestat/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/taihen/mcp-ripestat)](https://goreportcard.com/report/github.com/taihen/mcp-ripestat)

**`mcp-ripestat` is a Go-based MCP server that query data from the RIPEstat
Data API.**

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

## Network Operations & Monitoring

Quickly look up network information (ASN, prefix, routing, geolocation) for any
IP address or prefix as part of incident response or troubleshooting.
Integrate with monitoring dashboards to enrich alerts with RIPEstat data.

## Security Analysis

Automate enrichment of security events (e.g., SIEM logs, IDS alerts) with
RIPEstat data to provide context about suspicious IPs or networks. Investigate
the ownership and reputation of IP addresses seen in logs or threat intelligence
needs.

## Automation & Scripting

Use the MCP server as a local API for scripts and automation tools to fetch
RIPEstat data without each script needing to implement its own API logic.
Batch process lists of IPs or prefixes for network inventory or audit tasks.

## AI & Assistant Integration

Enable AI assistants to answer questions about IP addresses, ASNs, or network
blocks using up-to-date RIPEstat data. Provide context-aware suggestions or
actions in developer tools or chatbots.

## Centralized Access for Teams

Allow multiple team members to access RIPEstat data through a single, consistent
endpoint, reducing the need for individual API keys or configurations.
Enforce organizational policies, logging, or access controls at the MCP
server level.

## Data Enrichment for Applications

Enrich internal applications (e.g., asset management, customer portals) with
authoritative network information from RIPEstat. Combine RIPEstat data with other
sources for richer analytics or reporting.

## Education & Research

Provide students or researchers with easy access to Internet infrastructure data
for analysis, visualization, or learning projects.

## Architectural Rationale

`mcp-ripestat` is implemented as an HTTP server rather than a command-line
interface (CLI) tool to facilitate centralized deployment. This architecture
allows the service to be installed on a single server and accessed by multiple
team members and other MCP clients across a network. It promotes reusability,
simplifies client-side configuration, and provides a single, consistent access
point for RIPEstat data.

> [!WARNING]
> At current stage this MCP server does not provide authentication.

## Quick Start

### Build

To build the binary, run:

```sh
go build -o mcp-ripestat ./cmd/mcp-ripestat
```

### Run

To run the server locally:

```sh
go build -o mcp-ripestat ./cmd/mcp-ripestat
./mcp-ripestat
```

By default, the server runs on port `8080`. You can change this by setting the
`--port` flag. You can also enable debug logging with the `--debug` flag.

```sh
./mcp-ripestat --port=8888 --debug
```

### MCP Client Configuration

To integrate the server with MCP client, add the following configuration to your
global `mcp.json` file (e.g. for [Cursor](https://www.cursor.com/)
`~/.cursor/mcp.json` on macOS/Linux).

The example [mcp.json](./mcp.json) in this repository can be used as a
reference.

**Example Configuration:**

Merge the `ripestat` object from this repository's `mcp.json` into your existing
MCP client configuration file. For example:

```json
{
  "anthropic": {
    "url": "...",
    "endpoints": []
  },
  "ripestat": {
    "name": "mcp-ripestat-local",
    "description": "Local MCP RIPEstat server for RIPEstat Data API integration.",
    "url": "http://localhost:8080",
    "endpoints": [
      {
        "path": "/network-info",
        "method": "GET",
        "description": "Get network information for an IP address or prefix. Query param: resource"
      },
      {
        "path": "/as-overview",
        "method": "GET",
        "description": "Get an overview of an Autonomous System (AS). Query param: resource"
      }
    ]
  }
}
```

## API Endpoints

### `GET /.well-known/mcp/manifest.json`

Returns the MCP manifest for MCP client which describes the server's capabilities.

**Example:**

Development Testing:

```sh
curl 'http://localhost:8080/.well-known/mcp/manifest.json'
```

**Sample response:**

```json
{
  "name": "mcp-ripestat",
  "description": "A server for the RIPEstat Data API, providing network information for IP addresses and prefixes.",
  "functions": [
    {
      "name": "getNetworkInfo",
      "description": "Get network information for an IP address or prefix.",
      "parameters": [
        {
          "name": "resource",
          "type": "string",
          "required": true,
          "description": "The IP address or prefix to query."
        }
      ],
      "returns": {
        "type": "object"
      }
    },
    {
      "name": "getASOverview",
      "description": "Get an overview of an Autonomous System (AS).",
      "parameters": [
        {
          "name": "resource",
          "type": "string",
          "required": true,
          "description": "The AS number to query."
        }
      ],
      "returns": {
        "type": "object"
      }
    }
  ]
}
```

### `GET /network-info`

Returns network information for an IP address or prefix using the RIPEstat
`network-info` data API.

**Query parameters:**

- `resource`: The IP address or prefix to query (e.g., `140.78.90.50`).

**Example:**

MCP Client Prompt:

> What is the network info for 140.78.90.50?

Development Testing:

```sh
curl 'http://localhost:8080/network-info?resource=140.78.90.50'
```

**Sample response:**

```json
{
  "messages": [],
  "see_also": [],
  "version": "1.1",
  "data_call_name": "network-info",
  "data_call_status": "supported",
  "cached": false,
  "data": {
    "asns": ["1205"],
    "prefix": "140.78.0.0/16"
  },
  "query_id": "...",
  "process_time": 3,
  "server_id": "...",
  "build_version": "...",
  "status": "ok",
  "status_code": 200,
  "time": "..."
}
```

### `GET /as-overview`

Returns an overview for an AS (Autonomous System) using the RIPEstat
`as-overview` data API.

**Query parameters:**

- `resource`: The AS number to query (e.g., `3333`).

**Example:**

```sh
curl 'http://localhost:8080/as-overview?resource=3333'
```

## Contributing

Contributions are welcome! Please read [contributing guidelines](CONTRIBUTING.md)
to see how you can participate.

## Development

The development process is organized into sprints, with each sprint focused on a
specific feature from the RIPEstat API. A detailed ledger of all the features,
divided by sprints, is available in the [sprints](.github/SPRINTS.md)
documentation.

> [!NOTE]
> If a particular feature is not included, please feel free to open
> [issue](https://github.com/taihen/mcp-ripestat/issues?q=sort%3Aupdated-desc+is%3Aissue+is%3Aopen)
> to discuss it.

## License

This project is licensed under the MIT License. See the [license](LICENSE) file
for details.
