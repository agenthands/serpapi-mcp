# SerpApi MCP Server

[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8.svg)](https://go.dev/)
[![CI](https://github.com/agenthands/serpapi-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/agenthands/serpapi-mcp/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-86%25-green.svg)]()

A Go-based MCP (Model Context Protocol) server that exposes [SerpApi](https://serpapi.com) search capabilities through a single HTTP endpoint. Supports 100+ search engines through a single `search` tool, with per-engine parameter schemas available as MCP resources. Built for MCP client integrators (configure and connect) and Go developers (build from source, extend the codebase).

## Quickstart

### Hosted service

Get an API key from the [SerpApi dashboard](https://serpapi.com/dashboard), then configure your MCP client:

```json
{
  "mcpServers": {
    "serpapi": {
      "url": "https://mcp.serpapi.com/YOUR_SERPAPI_API_KEY/mcp"
    }
  }
}
```

See [USAGE.md](USAGE.md) for detailed configuration options.

### Self-hosting

Download a binary from [GitHub Releases](https://github.com/agenthands/serpapi-mcp/releases), or install with Go:

```bash
go install github.com/agenthands/serpapi-mcp@latest
serpapi-mcp --host 0.0.0.0 --port 8000
```

Configure your MCP client:

```json
{
  "mcpServers": {
    "serpapi": {
      "url": "http://localhost:8000/YOUR_SERPAPI_API_KEY/mcp"
    }
  }
}
```

See [INSTALL.md](INSTALL.md) for all installation methods.

## Authentication

Two methods are supported:

- **Path-based** (recommended): `/{YOUR_API_KEY}/mcp`
- **Header-based**: `Authorization: Bearer YOUR_API_KEY`

```bash
curl "https://mcp.serpapi.com/your_key/mcp" -d '...'
curl "https://mcp.serpapi.com/mcp" -H "Authorization: Bearer your_key" -d '...'
```

See [USAGE.md](USAGE.md) for authentication details.

## Search tool

The server provides a single `search` tool. All engines and parameters go through the `params` dict:

```json
{"name": "search", "arguments": {"params": {"q": "search query"}}}
```

Default engine is `google_light`. Set `mode` to `complete` (default) or `compact`. Engine parameter schemas are available as MCP resources at `serpapi://engines` (index) and `serpapi://engines/<engine>`.

See [USAGE.md](USAGE.md) for full parameter reference and examples.

## Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) — Package layout, request flows, subsystem designs
- [INSTALL.md](INSTALL.md) — Installation for all platforms
- [USAGE.md](USAGE.md) — Configuration, MCP client integration, error reference

## License

MIT License — see [LICENSE](LICENSE) for details.