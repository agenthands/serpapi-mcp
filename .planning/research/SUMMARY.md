# Research Summary: SerpApi MCP Server Go Rewrite

**Synthesized:** 2026-04-15
**Confidence:** HIGH

## Key Findings

### Stack
- **Official Go MCP SDK** (`modelcontextprotocol/go-sdk` v1.5.0+) — recommended. Stable v1.x API, official MCP org + Google backed, `StreamableHTTPHandler` returns `http.Handler` enabling clean middleware composition.
- **Go 1.25+** required by official SDK (stable since Aug 2025).
- **net/http stdlib** — no framework needed for a 2-route server. Go 1.22+ ServeMux handles method routing.
- **log/slog** (stdlib) — structured logging, sufficient for proxy service.
- **goreleaser** — de-facto standard for Go multi-platform binary releases.
- **Only one external dep**: `modelcontextprotocol/go-sdk`.

### Features
- **#1 bug to fix**: MCP error responses — Python returns `"Error: ..."` strings instead of `IsError: true` flag. Go SDK handles this natively (`mcp.NewToolResultError()`).
- **#2 bug to fix**: Input validation — Python forwards invalid engine names/modes directly to SerpApi.
- **14 P1 features**: Core server, auth middleware, search tool, engine resources, response modes, proper errors, input validation, structured logging, startup validation, graceful shutdown, healthcheck, CORS, test suite, multi-platform builds.
- **Defer**: MCP Completions, MCP Prompts, OAuth (P3 or anti-feature).

### Architecture
- Standard Go layout: `cmd/serpapi-mcp/main.go`, `internal/` packages.
- `getServer func(*http.Request) *mcp.Server` pattern — key architectural decision for API key flow.
- Engine schemas loaded from `engines/*.json` at startup (all stdlib JSON, no library needed).
- SerpApi HTTP calls via `net/http.Client` directly — no Go client library exists.
- Auth middleware composes with `StreamableHTTPHandler` via standard Go middleware pattern.

### Pitfalls
- **MCP protocol compliance**: Go SDK response format must match what existing Python clients expect.
- **Error response format**: Most critical pitfall — returning Go errors instead of MCP `IsError=true` results breaks clients.
- **JSON schema differences**: Python dicts → Go structs require careful mapping; engine schemas have variable parameter structures.
- **Over-engineering risk**: Go makes it easy to add interfaces, generics, etc. — keep it simple.
- **Engine schema generation**: `build-engines.py` scrapes SerpApi playground — need to decide whether to port to Go, keep as Python CI step, or embed.

## Recommended Build Order

1. **Project scaffolding** — Go module, directory structure, CI setup
2. **Core MCP server** — SDK integration, transport, healthcheck
3. **Auth middleware** — API key extraction from path + header
4. **Search tool** — SerpApi HTTP client, response modes, proper error handling
5. **Engine resources** — Schema loading, MCP resource registration
6. **Validation + observability** — Input validation, structured logging, startup checks
7. **Build + release** — goreleaser, multi-platform binaries