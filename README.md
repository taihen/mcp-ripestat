# MCP-RIPEStat

[![Lint](https://github.com/taihen/mcp-ripestat/actions/workflows/lint.yml/badge.svg)](https://github.com/taihen/ppp-exporter/actions/workflows/lint.yml)
[![Test](https://github.com/taihen/mcp-ripestat/actions/workflows/test.yml/badge.svg)](https://github.com/taihen/ppp-exporter/actions/workflows/test.yml)
[![Build](https://github.com/taihen/mcp-ripestat/actions/workflows/build.yml/badge.svg)](https://github.com/taihen/ppp-exporter/actions/workflows/build.yml)
[![Release](https://github.com/taihen/mcp-ripestat/actions/workflows/release.yml/badge.svg)](https://github.com/taihen/mcp-ripestat/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/taihen/mcp-ripestat)](https://goreportcard.com/report/github.com/taihen/mcp-ripestat)

`mcp-ripestat` is a Go-based, locally run MCP server that fetches and serves data from the RIPEstat Data API.

## About RIPEstat

RIPEstat is a large-scale information service and the RIPE NCC's open data platform. It provides essential data on IP address space and Autonomous System Numbers (ASNs), along with related statistics for specific hostnames and countries. This service is a valuable tool for network operators, security researchers, and anyone interested in the structure and performance of the Internet, offering insights into routing, registration data, DNS, and geographical information.

For more information, visit the [RIPEstat website](https://stat.ripe.net/) and consult the [Data API documentation](https://stat.ripe.net/docs/data_api).

## Architectural Rationale

`mcp-ripestat` is implemented as an HTTP server rather than a command-line interface (CLI) tool to facilitate centralized deployment. This architecture allows the service to be installed on a single server and accessed by multiple team members and other MCP clients across a network. It promotes reusability, simplifies client-side configuration, and provides a single, consistent access point for RIPEstat data.

## Disclaimer

This software is currently under active development. It may not be stable and is subject to change. Please use with caution.

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

By default, the server runs on port `8080`. You can change this by setting the `--port` flag. You can also enable debug logging with the `--debug` flag.

```sh
./mcp-ripestat --port=8888 --debug
```

## API Endpoints

### `GET /network-info`

Returns network information for an IP address or prefix using the RIPEstat `network-info` data API.

**Query parameters:**

- `resource` (required): The IP address or prefix to query (e.g., `140.78.90.50`).

**Example:**

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

## MCP Configuration for Local Development

To integrate the server with Cursor (and other MCP clients), add the following configuration to your global `mcp.json` file (e.g., `~/.cursor/mcp.json` on macOS/Linux).

The example [mcp.json](./mcp.json) in this repository can be used as a reference.

**Example Configuration:**

Merge the `ripestat` object from this repository's `mcp.json` into your existing configuration file. For example:

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
      }
    ]
  }
}
```

After adding this and restarting Cursor, you can test it with a prompt like: `What is the network info for 140.78.90.50?`

## License

This project is licensed under the MIT License. See the [license](LICENSE) file for details.

## Contributing

Contributions are welcome! Please read [contributing guidelines](CONTRIBUTING.md) to see how you can participate.

## Development Sprints

The development process is organized into sprints, with each sprint focused on a specific feature from the RIPEstat API. A detailed ledger of all the features, divided by sprints, is available in the [sprints](.github/SPRINTS.md) documentation.
