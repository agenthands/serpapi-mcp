---
phase: 03-search-validation-observability
type: research
level: 1
date: 2026-04-16
---

# Phase 3 Research: Search, Validation & Observability

**Finding level:** Quick verification — well-understood domain with established patterns.

## Standard Stack

- **HTTP Client:** `net/http.Client` with custom timeout — no external deps (per Phase 1 constraint)
- **Structured Logging:** `log/slog` — already used throughout codebase
- **MCP Tool API:** `mcp.AddTool` or `mcp.Server.AddTool` for registration; `CallToolResult.SetError(err)` for MCP-compliant errors with `IsError: true`
- **Input Validation:** Against in-memory engine schema data loaded by `engines.LoadAndRegister()`

## Architecture Patterns

### Go SDK Tool Registration

Two approaches available:
1. **`mcp.AddTool[In, Out]`** — Generic typed handler with automatic schema inference from Go struct types. Handler signature: `func(ctx, *CallToolRequest, In) (*CallToolResult, Out, error)`. SDK handles input validation via JSON schema.
2. **`mcp.Server.AddTool`** — Raw handler with `json.RawMessage` input schema. Handler signature: `func(ctx, *CallToolRequest) (*CallToolResult, error)`. Caller handles all validation.

**Decision: Use `mcp.Server.AddTool` with raw schema** — the `params` dict is dynamic (engine-specific keys), can't be represented as a fixed Go struct. The `mode` parameter is the only typed field. We validate engine name, mode, and required params ourselves (VAL-01 through VAL-03).

### CallToolResult & Error Handling

```go
type CallToolResult struct {
    Content           []Content `json:"content"`
    StructuredContent any       `json:"structuredContent,omitempty"`
    IsError           bool      `json:"isError,omitempty"`
}

func (r *CallToolResult) SetError(err error) // Sets Content to error text, IsError to true
```

For D-01 (MCP-compliant errors): Use `SetError()` or manually set `IsError: true` with structured JSON content.
For D-02 (flat error JSON body): Create `{"error": "<code>", "message": "<description>"}` as `TextContent` in `Content` field.

### SerpApi HTTP Call

- Endpoint: `https://serpapi.com/search`
- Method: GET with query parameters (api_key, engine, engine-specific params)
- Response: JSON body

### Compact Mode (D-08)

Remove 5 top-level keys from response JSON before returning:
- `search_metadata`, `search_parameters`, `search_information`, `pagination`, `serpapi_pagination`

Implementation: Parse SerpApi response as `map[string]any`, delete keys, re-serialize to JSON.

### Engine Validation Access

Current state: `engines.LoadAndRegister()` keeps `engineNames []string` and `schemas map[string]*engineSchema` as package-internal (unexported). Validation requires:
- **VAL-01:** List of valid engine names → Need `engines.EngineNames()` or similar exported accessor
- **VAL-03:** Required params per engine → Need `engines.RequiredParams(name)` or export schema data

Three approaches:
1. Add exported accessor functions to `internal/engines/engines.go`
2. Move validation logic into `internal/engines/` package
3. Return engine names and schemas from `LoadAndRegister` for downstream use

**Recommended:** Approach 1 — Add `EngineNames()` and `RequiredParams(name)` to engines package. Keeps validation data co-located with schema loading, simple to implement.

### Correlation IDs (OBS-01, OBS-02)

Options:
1. **HTTP middleware** — Generate UUID in middleware, store in request context, extract in tool handler
2. **Tool-side generation** — Generate in search tool handler directly

**Recommended:** Middleware — correlation IDs should span the entire request lifecycle (auth, routing, tool execution), not just the tool call. Inject via `context.WithValue` in a lightweight middleware.

Correlation ID format: Use `crypto/rand` hex string (16 bytes = 32 hex chars). No UUID library needed (stdlib-only constraint).

## Don't Hand-Roll

- MCP error format — Use `CallToolResult.SetError()` for consistent MCP behavior, then override Content with structured JSON body for D-02
- SerpApi URL construction — Use `url.URL` and `url.Values` for proper query parameter encoding
- Engine name list — Access via exported function from engines package, don't re-read filesystem

## Common Pitfalls

1. **Dynamic params dict** — Can't use typed Go struct for tool input. Must use `map[string]any` with raw JSON schema.
2. **API key propagation** — Go SDK's `CallToolRequest` must carry the HTTP request context. Verify `request.Context()` propagates through `APIKeyFromContext`.
3. **Compact mode on non-object responses** — SerpApi occasionally returns arrays or non-object JSON. Compact mode delete-keys logic only works on `map[string]any` — guard against nil or non-dict responses.
4. **SerpApi HTTP errors as JSON** — Error responses from SerpApi may not be JSON. Handle both JSON and non-JSON error bodies gracefully.
5. **Engine schema required params structure** — `required: true` is a field-level boolean in engine JSON params, not a JSON Schema `required` array. Check field-by-field.

## Validation Architecture

### Dimension: Error Behavior

**Observable:** Every error from the search tool returns `IsError: true` with a flat error JSON body (`{"error": "...", "message": "..."}`).

**Cases:**
- Invalid engine name → `{"error": "invalid_engine", "message": "Engine 'x' not found. Available: ..."}` with `IsError: true`
- Invalid mode → `{"error": "invalid_mode", "message": "Mode must be 'complete' or 'compact'"}` with `IsError: true`
- Missing required params → `{"error": "missing_params", "message": "Missing required parameter 'q' for engine 'google_light'"}` with `IsError: true`
- SerpApi 429 → `{"error": "rate_limited", "message": "Rate limit exceeded..."}` with `IsError: true`
- SerpApi 401 → `{"error": "invalid_api_key", "message": "Invalid SerpApi API key..."}` with `IsError: true`
- SerpApi 403 → `{"error": "forbidden", "message": "SerpApi API key forbidden..."}` with `IsError: true`
- Network/5xx/other → `{"error": "search_error", "message": "..."}` with `IsError: true`

**Verification:** Each error type produces CallToolResult with IsError=true and Content containing flat JSON.

### Dimension: Compact Mode

**Observable:** Compact mode response lacks the 5 specified fields while retaining all other data.

**Cases:**
- Complete mode → full response with all fields
- Compact mode → response minus search_metadata, search_parameters, search_information, pagination, serpapi_pagination
- Empty/missing fields → deletion is no-op (pop with None semantics)

**Verification:** Deserialize complete and compact responses, assert 5 keys absent in compact.

### Dimension: Input Validation

**Observable:** Invalid inputs are rejected before any SerpApi HTTP call.

**Cases:**
- Unknown engine → rejection with available engine list
- Mode outside "complete"/"compact" → rejection
- Missing required param for engine → rejection with param name

**Verification:** Mock-free tests can validate rejection logic without network calls.

### Dimension: Correlation ID Propagation

**Observable:** Every log line during a search request includes the same correlation ID.

**Cases:**
- Correlation ID generated at middleware level
- ID stored in request context
- Search tool extracts and logs with correlation ID
- Error paths also include correlation ID

**Verification:** Capture log output during search, assert correlation_id present in all entries for a request.

---

*Research completed: 2026-04-16*
*Discovery level: 1 (Quick Verification)*