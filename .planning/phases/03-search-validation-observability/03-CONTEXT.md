# Phase 3: Search, Validation & Observability - Context

**Gathered:** 2026-04-16
**Status:** Ready for planning

<domain>
## Phase Boundary

AI agents can search any SerpApi engine through the MCP tool with validated inputs and proper error handling. This phase delivers: search tool with complete/compact modes, MCP-compliant error responses, input validation (engine names, mode, required params), and structured logging with request correlation IDs.

</domain>

<decisions>
## Implementation Decisions

### Error responses
- **D-01:** MCP-compliant errors use `IsError: true` flag in `CallToolResult` — not Python's `"Error: ..."` string prefix pattern. This is the MCP standard for tool errors and lets MCP clients programmatically detect failures.
- **D-02:** Error JSON body is a flat structure: `{"error": "<code>", "message": "<human-readable>"}` — consistent shape for all error types, easy for clients to parse without handling varying fields.
- **D-03:** SerpApi HTTP error mapping matches Python's granularity: 429 (rate limit), 401 (invalid key), 403 (forbidden), plus a generic catch-all for unexpected errors (network, timeout, 5xx, etc.). No need for per-status-code differentiation beyond these three — they cover the real failure scenarios.

### Search tool behavior
- **D-04:** Search tool signature: `search(params: dict, mode: string="complete")` — same as Python, all engines and parameters go through the `params` dict (SRCH-01).
- **D-05:** Default engine is `google_light` — matches Python (SRCH-02).
- **D-06:** API key extracted from request context via `APIKeyFromContext(ctx)` — Phase 2 established this pattern.
- **D-07:** SerpApi HTTP calls via `net/http.Client` with reasonable timeouts — no external HTTP client library (Phase 1 constraint: go-sdk only external dep).

### Compact mode
- **D-08:** Compact mode removes the same 5 fields as Python: `search_metadata`, `search_parameters`, `search_information`, `pagination`, `serpapi_pagination` — faithful port behavior.

### Input validation
- **D-09:** Invalid engine names rejected with error message listing available engines (VAL-01) — engine names checked against in-memory schemas loaded at startup (Phase 2).
- **D-10:** Mode parameter accepts only `"complete"` or `"compact"` (VAL-02) — matches Python validation.
- **D-11:** Required parameters validated per engine schema (VAL-03) — check `required: true` fields in the engine's `params` section.

### Observability
- **D-12:** Structured logging via `log/slog` with request correlation IDs (OBS-01, OBS-02) — `slog` already used throughout the codebase.
- **D-13:** Startup confirmation message with port and engine count (OBS-03) — already implemented in Phase 2.

### Agent's Discretion
- Exact HTTP client timeout values for SerpApi calls
- Correlation ID generation method (UUID, random hex, etc.)
- How correlation IDs propagate through request context (middleware injection vs. tool-side generation)
- Compact mode field removal implementation (iterate-and-delete vs. rebuild-without)
- Whether validation errors also use the flat error JSON format or a different shape
- Error response for when API key is missing from tool-level context (shouldn't happen if auth middleware works, but needs a fallback)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Context
- `.planning/PROJECT.md` — Vision, constraints (no external deps beyond go-sdk, no Python client port), API compatibility requirements
- `.planning/REQUIREMENTS.md` — SRCH-01 through SRCH-07, VAL-01 through VAL-03, OBS-01 through OBS-03
- `.planning/ROADMAP.md` — Phase 3 goal, success criteria, plan outline

### Legacy Reference
- `legacy/src/server.py` — Python search tool: params dict, mode validation, compact mode field removal list, SerpApi HTTP error handling (429/401/403), `extract_error_response` helper

### Existing Go Code
- `internal/server/auth.go` — `APIKeyFromContext(ctx)` function for extracting API key from request context
- `internal/engines/engines.go` — Engine schema loading, in-memory schema map, engine name validation
- `internal/server/server.go` — MCP server setup, `slog.Logger` usage, handler chain
- `cmd/serpapi-mcp/main.go` — Initialization flow, flag parsing

### Go SDK
- `go-sdk/mcp` package — `AddTool`, `CallToolResult`, `IsError` flag for tool error responses

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/server/auth.go:APIKeyFromContext()` — Extracts API key from request context; search tool will call this to get the key for SerpApi
- `internal/engines/engines.go:LoadAndRegister()` — Returns count and validates schemas; the in-memory schema data (engine names, required params) can be reused for input validation
- `internal/server/server.go:MCPServer` — Holds `MCPServer` and `HTTPHandler`; search tool registers on `MCPServer` via `AddTool`
- `cmd/serpapi-mcp/main.go` — Initialization flow already wires engines → server → run; search tool registration follows the same pattern
- `log/slog` — Already used throughout; structured logging with key-value pairs is the established pattern

### Established Patterns
- MCP tool/resource registration via `mcp.Server.AddTool`/`AddResource` — consistent with Phase 2
- Custom context keys for cross-middleware data transfer — established in Phase 2
- JSON error responses with `{"error": "..."}` format — auth errors already use this; search tool errors will extend it with `{"error": "...", "message": "..."}`
- stdlib-only HTTP handling — `net/http` for SerpApi calls, no external HTTP client

### Integration Points
- Search tool registration in `main.go` or a new `internal/search/search.go` — currently empty stub package exists
- Tool handler receives `mcp.CallToolRequest`, extracts `params` and `mode` from `request.Params.Arguments`
- API key from `APIKeyFromContext(request.Context())` — requires the MCP SDK to propagate HTTP request context through to tool handlers
- SerpApi HTTP endpoint: `https://serpapi.com/search` with query parameters including `api_key`, `engine`, and engine-specific params
- Compact mode field removal applied to the JSON response dict before returning

</code_context>

<specifics>
## Specific Ideas

- Error format should be consistent: always `{"error": "<code>", "message": "<description>"}` regardless of error type — clients can rely on the same shape
- Compact mode should remove the same 5 fields as Python — faithful port, no surprises for existing users
- Validation error listing available engines helps MCP client developers discover valid engine names quickly

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 03-search-validation-observability*
*Context gathered: 2026-04-16*