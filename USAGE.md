# Usage

How to configure, run, integrate, and troubleshoot the SerpApi MCP Server.

## Configuration and Running

### CLI Flags

| Flag | Env var | Default | Description |
|------|---------|---------|-------------|
| `--host` | `MCP_HOST` | `0.0.0.0` | Host to bind the server to |
| `--port` | `MCP_PORT` | `8000` | Port to bind the server to |
| `--cors-origins` | `MCP_CORS_ORIGINS` | `*` | Comma-separated list of allowed CORS origins |
| `--auth-disabled` | `MCP_AUTH_DISABLED` | `false` | Disable API key authentication (for testing) |
| `--engines-dir` | `ENGINES_DIR` | `engines` | Path to directory containing engine JSON schemas |

Environment variable helpers: `envOr` (string), `envIntOr` (int, falls back on invalid), `envBoolOr` (accepts `1`, `true`, `yes` as truthy values).

### Starting the Server

Use defaults:

```bash
serpapi-mcp
```

Override with CLI flags:

```bash
serpapi-mcp --port 3000 --host 127.0.0.1
```

Override with environment variables:

```bash
MCP_PORT=3000 serpapi-mcp
MCP_HOST=127.0.0.1 MCP_PORT=3000 serpapi-mcp
```

### Version Check

```bash
serpapi-mcp --version
```

Output format: `serpapi-mcp {version} (commit: {commit}, built: {date})`

### Environment Variables

The server reads environment variables directly — there is no `.env` file loading. Set variables via shell environment or a process manager:

```bash
export MCP_HOST=0.0.0.0
export MCP_PORT=8000
serpapi-mcp
```

See [.env.example](.env.example) for a template with default values.

## API Key Authentication

The server supports two authentication methods for the `/mcp` endpoint. Bearer header authentication takes priority over path-based authentication.

### Path-Based Authentication (Recommended)

Include your API key directly in the URL path:

- **Self-hosted:** `http://localhost:8000/YOUR_API_KEY/mcp`
- **Hosted:** `https://mcp.serpapi.com/YOUR_API_KEY/mcp`

The auth middleware strips the key segment from the URL path before forwarding to the MCP handler, so downstream handlers see `/mcp` rather than `/{KEY}/mcp`.

### Header-Based Authentication

Include your API key in the `Authorization` header:

```
Authorization: Bearer YOUR_API_KEY
```

When both methods are provided, Bearer header takes priority over the path-based key.

### Authentication Priority

1. `Authorization: Bearer {KEY}` header (checked first)
2. `/{KEY}/mcp` path pattern (fallback)

### Auth-Disabled Mode

For local development and testing, disable authentication entirely:

```bash
serpapi-mcp --auth-disabled
# or
MCP_AUTH_DISABLED=true serpapi-mcp
```

When auth is disabled, the auth middleware is bypassed — all requests are allowed without an API key.

### Health Endpoint Exemption

The `/health` endpoint is exempt from authentication and always returns:

```json
{"status": "healthy", "service": "SerpApi MCP Server"}
```

### Authentication Errors

Missing API key returns HTTP 401:

```json
{"error": "Missing API key. Use path format /{API_KEY}/mcp or Authorization: Bearer {API_KEY} header"}
```

## MCP Client Integration

The server uses the Streamable HTTP protocol as defined by the MCP specification. There are no SSE or WebSocket transport alternatives — only Streamable HTTP.

### Claude Desktop

Add to your Claude Desktop configuration file:

**Hosted:**
```json
{
  "mcpServers": {
    "serpapi": {
      "url": "https://mcp.serpapi.com/YOUR_SERPAPI_API_KEY/mcp"
    }
  }
}
```

**Self-hosted:**
```json
{
  "mcpServers": {
    "serpapi": {
      "url": "http://localhost:8000/YOUR_SERPAPI_API_KEY/mcp"
    }
  }
}
```

### Cursor

Add to your Cursor MCP settings:

**Hosted:**
```json
{
  "mcpServers": {
    "serpapi": {
      "url": "https://mcp.serpapi.com/YOUR_SERPAPI_API_KEY/mcp"
    }
  }
}
```

**Self-hosted:**
```json
{
  "mcpServers": {
    "serpapi": {
      "url": "http://localhost:8000/YOUR_SERPAPI_API_KEY/mcp"
    }
  }
}
```

### VS Code Copilot

Add to your VS Code `settings.json` or `.vscode/mcp.json`:

**Hosted:**
```json
{
  "mcp": {
    "servers": {
      "serpapi": {
        "url": "https://mcp.serpapi.com/YOUR_SERPAPI_API_KEY/mcp"
      }
    }
  }
}
```

**Self-hosted:**
```json
{
  "mcp": {
    "servers": {
      "serpapi": {
        "url": "http://localhost:8000/YOUR_SERPAPI_API_KEY/mcp"
      }
    }
  }
}
```

### Windsurf

Add to your Windsurf MCP configuration:

**Hosted:**
```json
{
  "mcpServers": {
    "serpapi": {
      "url": "https://mcp.serpapi.com/YOUR_SERPAPI_API_KEY/mcp"
    }
  }
}
```

**Self-hosted:**
```json
{
  "mcpServers": {
    "serpapi": {
      "url": "http://localhost:8000/YOUR_SERPAPI_API_KEY/mcp"
    }
  }
}
```

## Engine Discovery and Search

### Engine Index Resource

The `serpapi://engines` resource returns an index of all available engines:

```json
{
  "count": 107,
  "engines": ["baidu", "bing", "duckduckgo", "google", "google_light", "..."],
  "resources": ["serpapi://engines/baidu", "serpapi://engines/bing", "..."],
  "schema": {
    "note": "Each engine resource uses a flat schema: params are engine-specific; common_params are shared SerpApi parameters.",
    "params_key": "params",
    "common_params_key": "common_params"
  }
}
```

### Per-Engine Resource

Each engine has a dedicated resource at `serpapi://engines/<engine>` that returns the full parameter schema, including required and optional parameters:

```
serpapi://engines/google_light
serpapi://engines/google
serpapi://engines/bing
```

### Search Tool

The server provides a single `search` tool. All engines and parameters go through the `params` dict.

**Default engine:** `google_light`
**Default mode:** `complete`

#### Example Search Payloads

Basic search (uses default engine `google_light` and default mode `complete`):

```json
{"name": "search", "arguments": {"params": {"q": "search query"}}}
```

Specific engine:

```json
{"name": "search", "arguments": {"params": {"q": "search query", "engine": "google"}}}
```

Compact mode (removes metadata fields):

```json
{"name": "search", "arguments": {"params": {"q": "search query"}, "mode": "compact"}}
```

With location:

```json
{"name": "search", "arguments": {"params": {"q": "coffee shops", "engine": "google", "location": "Austin, TX"}}}
```

#### Compact Mode

When `mode` is `"compact"`, the following top-level fields are removed from the response:

- `search_metadata`
- `search_parameters`
- `search_information`
- `pagination`
- `serpapi_pagination`

This produces smaller responses suitable for AI agent contexts where metadata is not needed.

## Error Reference

All search tool errors use MCP-compliant `IsError=true` with a flat JSON body format:

```json
{"error": "<code>", "message": "<description>"}
```

### Authentication Errors

**401 Missing API Key** (from auth middleware):

```json
{"error": "Missing API key. Use path format /{API_KEY}/mcp or Authorization: Bearer {API_KEY} header"}
```

- **Cause:** No API key provided in URL path or Authorization header.
- **Fix:** Include your API key in the URL path (`/{KEY}/mcp`) or as a `Bearer` token in the `Authorization` header.

**missing_api_key** (from search tool, when auth-disabled mode bypassed auth):

```json
{"error": "missing_api_key", "message": "No API key found in request context. Use path format /{API_KEY}/mcp or Authorization: Bearer {API_KEY} header."}
```

- **Cause:** Auth middleware was bypassed (auth-disabled mode) but no API key was available in context for SerpApi calls.
- **Fix:** Use auth-enabled mode and pass your API key via path or header, or configure a default key.

### Validation Errors

**invalid_engine**:

```json
{"error": "invalid_engine", "message": "invalid_engine: engine 'googl' not found. Available engines: baidu, bing, duckduckgo, google, google_light, ..."}
```

- **Cause:** Engine name is misspelled or not in the loaded engine list.
- **Fix:** Check `serpapi://engines` for valid engine names.

**invalid_mode**:

```json
{"error": "invalid_mode", "message": "invalid_mode: mode must be 'complete' or 'compact', got 'fast'"}
```

- **Cause:** Mode value is not `"complete"` or `"compact"`.
- **Fix:** Use `"complete"` for full results or `"compact"` for metadata-stripped results.

**missing_params**:

```json
{"error": "missing_params", "message": "missing_params: missing required parameter(s) q for engine 'google'"}
```

- **Cause:** Required parameters for the engine were omitted.
- **Fix:** Check `serpapi://engines/<engine>` for the engine's required parameters.

### SerpApi Errors

**rate_limited** (HTTP 429):

```json
{"error": "rate_limited", "message": "Rate limit exceeded. Please try again later."}
```

- **Cause:** SerpApi rate limit exceeded for your API key.
- **Fix:** Wait and retry, or upgrade your SerpApi plan at [serpapi.com](https://serpapi.com).

**invalid_api_key** (HTTP 401):

```json
{"error": "invalid_api_key", "message": "Invalid SerpApi API key. Check your API key in the path or Authorization header."}
```

- **Cause:** The provided SerpApi API key is invalid or revoked.
- **Fix:** Verify your key at [serpapi.com/dashboard](https://serpapi.com/dashboard).

**forbidden** (HTTP 403):

```json
{"error": "forbidden", "message": "SerpApi API key forbidden. Verify your subscription and key validity."}
```

- **Cause:** API key does not have the required subscription level.
- **Fix:** Check your SerpApi subscription status.

**search_error** (catch-all):

```json
{"error": "search_error", "message": "<context-specific error details>"}
```

- **Cause:** Various — invalid URL, HTTP error, JSON parse failure, or unexpected SerpApi response.
- **Fix:** Check the error message for specifics. Common causes include network issues, malformed parameters, or SerpApi service errors.

## Troubleshooting

### Symptom → Cause → Fix

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| "Connection refused" | Wrong host/port or server not running | Check `MCP_HOST` and `MCP_PORT` match your config; verify the server is running with `curl http://localhost:8000/health` |
| "401 Unauthorized" or "Missing API key" | No API key provided | Include your API key in the URL path (`/{KEY}/mcp`) or `Authorization: Bearer {KEY}` header |
| "Invalid SerpApi API key" | Wrong or revoked key | Verify your key at [serpapi.com/dashboard](https://serpapi.com/dashboard) |
| "Engine not found" | Typo in engine name | Check `serpapi://engines` for valid names — engine names are lowercase with underscores (e.g., `google_light`, not `Google Light`) |
| "Missing params" | Required parameter omitted | Check `serpapi://engines/<engine>` for required params — most engines require `q` at minimum |
| "Rate limit exceeded" | Too many requests | Wait and retry, or upgrade your SerpApi plan |
| "Failed to load engine schemas" | Missing engines/ directory | Check `--engines-dir` flag points to the directory containing engine JSON files; run from repo root or set `ENGINES_DIR` environment variable |
| Empty search results | Engine requires additional params | Some engines need `location`, `gl`, or `hl` params — check `serpapi://engines/<engine>` for required fields |
| CORS errors in browser | Origin not allowed | Set `--cors-origins` to your client origin (e.g., `http://localhost:3000`) or `*` for all origins |

### Debug Tips

**Structured logging:** The server uses Go's `slog` for structured logging. Check stderr output for request metadata including `correlation_id`, `engine`, `mode`, and `params_count`.

**Check engine schemas:** Use MCP resources to discover available engines and their parameters:

- `serpapi://engines` — lists all engines with their resource URIs
- `serpapi://engines/<engine>` — full parameter schema for a specific engine

**Test with curl:**

```bash
# Health check (no auth required)
curl http://localhost:8000/health

# Search with path-based auth
curl -X POST "http://localhost:8000/YOUR_KEY/mcp" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"search","arguments":{"params":{"q":"test"}}},"id":1}'

# Search with header-based auth
curl -X POST "http://localhost:8000/mcp" \
  -H "Authorization: Bearer YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"search","arguments":{"params":{"q":"test"}}},"id":1}'
```

**Correlation IDs:** Each request receives a 32-character hex correlation ID generated from `crypto/rand`. Provide your own via the `X-Correlation-ID` request header for distributed tracing. The correlation ID is echoed back in the `X-Correlation-ID` response header and included in structured log entries.

## Next Steps

- [INSTALL.md](INSTALL.md) — Installation instructions for all platforms (binary download, `go install`, build from source)
- [ARCHITECTURE.md](ARCHITECTURE.md) — Deep-dive into package layout, request flows, and subsystem designs
- [README.md](README.md) — Project overview and quickstart guide