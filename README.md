# MCP-RIPEStat

`mcp-ripestat` is a Go-based, locally run MCP server that fetches and serves data from the RIPEstat Data API.

## About RIPEstat

RIPEstat is a large-scale information service and the RIPE NCC's open data platform. It provides essential data on IP address space and Autonomous System Numbers (ASNs), along with related statistics for specific hostnames and countries. This service is a valuable tool for network operators, security researchers, and anyone interested in the structure and performance of the Internet, offering insights into routing, registration data, DNS, and geographical information.

For more information, visit the [RIPEstat website](https://stat.ripe.net/) and consult the [Data API documentation](https://stat.ripe.net/docs/data_api).

## Disclaimer

This software is currently under active development. It may not be stable and is subject to change. Please use with caution.

## Quick Start

### Build

To build the binary, run:

```sh
go build -o mcp-ripestat ./cmd/mcp-ripestat
```

### Run

To run the server:

```sh
./mcp-ripestat
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Development Sprints

The development process is organized into sprints, with each sprint focused on a specific feature from the RIPEstat API. A detailed ledger of all sprints, including branches, target versions, and features, is available in the [SPRINTS.md](.github/SPRINTS.md) file.
