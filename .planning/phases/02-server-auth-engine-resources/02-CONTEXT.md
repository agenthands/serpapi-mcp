# Phase 2: Server, Auth & Engine Resources - Context

**Gathered:** 2026-04-16
**Status:** Ready for planning

<domain>
## Phase Boundary

A running Go MCP server accepts authenticated connections and serves engine parameter schemas as MCP resources. This phase delivers: Streamable HTTP transport, healthcheck endpoint, CORS support, graceful shutdown, API key auth middleware (path-based and Bearer header), engine schema loading with startup validation, and engine list + per-engine schema resources.

</domain>

<decisions>
## Implementation Decisions

### Auth middleware & key passing
- **D-01:** Unified auth middleware handles both header (`Authorization: Bearer {KEY}`) and path-based (`/{KEY}/mcp`) authentication in a single middleware — matches Python's single-middleware model, simpler handler chain
- **D-02:** API key passed to downstream handlers via custom context key (not go-sdk's TokenInfo) — decouples from SDK's token model, explicit and simple
- **D-03:** Path-based auth: middleware strips the API key segment from the URL path (`/{KEY}/mcp` → `/mcp`) before forwarding to the MCP handler, same pattern as Python
- **D-04:** Auth errors return 401 with JSON body (`{"error": "Missing API key..."}`) — consistent with Python server's error format

### Engine loading strategy
- **D-05:** All 107 engine schemas loaded into memory at startup — validates and parses every JSON file, fails fast on corrupt/missing (ENG-04), fast per-request serving with no disk I/O after init
- **D-06:** All 107 engines registered as individual static resources via `AddResource` calls at startup — explicit, each engine has its own fixed URI (`serpapi://engines/{engine}`), matches Python's per-engine factory pattern

### Configuration & flags
- **D-07:** Server host/port support both env vars (`MCP_HOST`, `MCP_PORT`) and CLI flags (`--host`, `--port`) — CLI flags take precedence over env vars; standard Go practice
- **D-08:** Use stdlib `flag` package for CLI parsing — no additional external dependency, keeps go-sdk as the only external dep per D-18 from Phase 1
- **D-09:** Server startup logs confirmation message with port and engine count (satisfies OBS-03, implemented early since it's trivial and useful for debugging)

### Healthcheck & CORS
- **D-10:** Healthcheck endpoint at `/health` returning 200 OK — matches MCP-02 requirement; no backward-compatible `/healthcheck` alias
- **D-11:** CORS defaults to `allow_origins=["*"]` with all methods and headers — matches Python behavior; a `--cors-origins` CLI flag allows restricting origins when needed
- **D-12:** CORS middleware applied at the HTTP handler level before the MCP StreamableHTTPHandler — standard Go middleware wrapping order

### Agent's Discretion
- Exact middleware chain ordering (auth → CORS → MCP handler vs. CORS → auth → MCP handler) — either works, agent picks the most logical order
- Engine schema struct representation in Go (map vs. typed struct) — agent decides based on readability and performance
- Graceful shutdown timeout value — standard Go pattern with reasonable default
- Error response JSON format details beyond the `error` key — agent defines exact structure

### Folded Todos
(None folded into scope)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Context
- `.planning/PROJECT.md` — Vision, constraints, key decisions, API compatibility requirements
- `.planning/REQUIREMENTS.md` — MCP-01 through ENG-05 requirements for this phase
- `.planning/ROADMAP.md` — Phase 2 goal, success criteria, and plan outline

### Legacy Reference
- `legacy/src/server.py` — Python MCP server reference: ApiKeyMiddleware pattern, engine resource registration, CORS config, healthcheck handler, search tool signature

### Go SDK Documentation
- `go-sdk/mcp` package — StreamableHTTPHandler, AddResource, AddResourceTemplate, Server, Resource types
- `go-sdk/auth` package — RequireBearerToken middleware reference (for understanding SDK patterns, though we use custom auth)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `cmd/serpapi-mcp/main.go` — Entry point stub with version/commit/date vars; will be extended to initialize MCP server
- `internal/server/server.go` — Empty package declaration; will house MCP server setup and middleware
- `internal/engines/engines.go` — Empty package declaration; will house engine schema loading and resource registration
- `engines/*.json` (107 files) — Engine schemas consumed at runtime; validated and loaded into memory at startup
- `legacy/src/server.py` — Full Python reference implementation for auth middleware, engine resources, CORS, and healthcheck

### Established Patterns
- Go module path: `github.com/agenthands/serpapi-mcp`
- Internal packages grouped by domain: `internal/server`, `internal/search`, `internal/engines`
- go-sdk v1.5.0 as sole external dependency
- goreleaser for multi-platform binary distribution (no Docker)

### Integration Points
- Auth middleware wraps `StreamableHTTPHandler` via standard Go `http.Handler` chaining (AUTH-03)
- Engine package registers resources on the MCP `Server` instance at startup
- Main function: parse CLI flags → load engines → create MCP server → register resources → wrap with middleware → start HTTP server → graceful shutdown
- Environment variables: `MCP_HOST` (default `0.0.0.0`), `MCP_PORT` (default `8000`) — matching Python defaults

</code_context>

<specifics>
## Specific Ideas

- Auth middleware should feel like the Python version: single middleware, clean path stripping, JSON error on missing key
- Engine resources should be straightforward to discover — list all at `serpapi://engines`, individual schemas at `serpapi://engines/{engine}`
- The Go binary should "just work" with sensible defaults (0.0.0.0:8000, wide-open CORS) matching Python's zero-config experience

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 02-server-auth-engine-resources*
*Context gathered: 2026-04-16*