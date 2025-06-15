# mcp-ripestat

`mcp-ripestat` is a Go-based, single-binary, locally run MCP server that fetches and serves data from the RIPEstat Data API.

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
