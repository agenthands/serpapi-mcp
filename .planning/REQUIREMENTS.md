# Requirements: SerpApi MCP Server — Go Rewrite

**Defined:** 2026-04-15
**Core Value:** AI agents can search any SerpApi-supported engine through a single, authenticated MCP endpoint with structured parameter discovery and proper MCP-compliant error handling.

## v1 Requirements

Requirements for the Go rewrite. Each maps to roadmap phases.

### Project Setup

- [x] **SETUP-01**: Go module initialized with `modelcontextprotocol/go-sdk` as only external dependency
- [x] **SETUP-02**: Standard Go project layout (`cmd/serpapi-mcp/main.go`, `internal/` packages)
- [x] **SETUP-03**: Legacy Python code moved to `legacy/` directory (src/, build-engines.py, pyproject.toml, etc.)
- [x] **SETUP-04**: CI workflow for Go: lint (golangci-lint), vet, test on PRs
- [x] **SETUP-05**: goreleaser configuration for multi-platform binary builds (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)

### MCP Server Core

- [x] **MCP-01**: Streamable HTTP transport using `modelcontextprotocol/go-sdk` StreamableHTTPHandler
- [x] **MCP-02**: Healthcheck endpoint at `/health` returning 200 OK
- [x] **MCP-03**: CORS support matching Python server behavior
- [x] **MCP-04**: Graceful shutdown on SIGINT/SIGTERM using `signal.NotifyContext()`

### Authentication

- [x] **AUTH-01**: API key extraction from URL path (`/{KEY}/mcp`) — maintains Python server compatibility
- [x] **AUTH-02**: API key extraction from `Authorization: Bearer {KEY}` header — maintains Python server compatibility
- [x] **AUTH-03**: Auth middleware composed with StreamableHTTPHandler via standard Go `http.Handler` wrapping

### Search Tool

- [x] **SRCH-01**: Single `search` tool accepting `params` dict with `engine`, `mode`, and SerpApi parameters
- [x] **SRCH-02**: Default engine is `google_light` (matching Python server)
- [x] **SRCH-03**: Complete mode returns full SerpApi response
- [x] **SRCH-04**: Compact mode removes specified fields from response
- [x] **SRCH-05**: MCP-compliant error responses using `IsError: true` flag (not string prefixes)
- [x] **SRCH-06**: SerpApi HTTP calls via `net/http.Client` with reasonable timeouts
- [x] **SRCH-07**: Proper handling of SerpApi rate limits (429), auth errors (401/403), and server errors (5xx)

### Engine Resources

- [x] **ENG-01**: Engine schemas loaded from `engines/*.json` at startup
- [x] **ENG-02**: Engine list resource at `serpapi://engines` returning all available engine names
- [x] **ENG-03**: Per-engine schema resource at `serpapi://engines/{engine}` returning parameter schema
- [x] **ENG-04**: Startup validation of engine schemas — fail fast on corrupt or missing JSON
- [x] **ENG-05**: Engine schema generation remains via `build-engines.py` (Python script in CI, not ported to Go)

### Input Validation

- [x] **VAL-01**: Reject invalid engine names with clear error message listing available engines
- [x] **VAL-02**: Validate `mode` parameter accepts only "complete" or "compact"
- [x] **VAL-03**: Validate required SerpApi parameters per engine schema

### Observability

- [x] **OBS-01**: Structured logging via `log/slog` with request correlation IDs
- [x] **OBS-02**: Correlation ID included in all log entries for a request
- [x] **OBS-03**: Startup log message confirming server ready, port, and engine count

### Testing

- [ ] **TEST-01**: Unit tests for search tool (mocking SerpApi HTTP responses)
- [ ] **TEST-02**: Unit tests for engine resource loading and schema retrieval
- [ ] **TEST-03**: Integration tests for auth middleware (path-based and header-based key extraction)
- [ ] **TEST-04**: Integration tests for healthcheck endpoint
- [ ] **TEST-05**: Unit tests for compact mode field removal
- [ ] **TEST-06**: Unit tests for input validation (invalid engine, invalid mode, missing params)

## v2 Requirements

Deferred to future release.

### MCP Features

- **MCP-V2-01**: MCP Prompts for curated search templates
- **MCP-V2-02**: MCP Completions for engine and parameter auto-suggestion
- **MCP-V2-03**: Progress reporting for long-running searches

### Infrastructure

- **INFRA-01**: Docker image build (alternative deployment)
- **INFRA-02**: Response compression middleware
- **INFRA-03**: Request timeout configuration via CLI flags

## Out of Scope

| Feature | Reason |
|---------|--------|
| OAuth/API key storage | Auth is stateless; each request carries its own key |
| Result caching | SerpApi caches on their side |
| Admin dashboard | Operations handled via binary flags and logs |
| Database/persistent storage | Server is stateless by design |
| SSE/WebSocket transport | Streamable HTTP is MCP's current standard |
| Python build-engines.py port to Go | Keep Python script in CI; Go server only consumes JSON |
| CloudWatch EMF metrics | Go binary has no AWS-specific dependencies |
| serpapi Python client port | Go calls SerpApi HTTP API directly |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| SETUP-01 | Phase 1 | Complete |
| SETUP-02 | Phase 1 | Complete |
| SETUP-03 | Phase 1 | Complete |
| SETUP-04 | Phase 1 | Complete |
| SETUP-05 | Phase 1 | Complete |
| MCP-01 | Phase 2 | Complete |
| MCP-02 | Phase 2 | Complete |
| MCP-03 | Phase 2 | Complete |
| MCP-04 | Phase 2 | Complete |
| AUTH-01 | Phase 2 | Complete |
| AUTH-02 | Phase 2 | Complete |
| AUTH-03 | Phase 2 | Complete |
| ENG-01 | Phase 2 | Complete |
| ENG-02 | Phase 2 | Complete |
| ENG-03 | Phase 2 | Complete |
| ENG-04 | Phase 2 | Complete |
| ENG-05 | Phase 2 | Complete |
| SRCH-01 | Phase 3 | Complete |
| SRCH-02 | Phase 3 | Complete |
| SRCH-03 | Phase 3 | Complete |
| SRCH-04 | Phase 3 | Complete |
| SRCH-05 | Phase 3 | Complete |
| SRCH-06 | Phase 3 | Complete |
| SRCH-07 | Phase 3 | Complete |
| VAL-01 | Phase 3 | Complete |
| VAL-02 | Phase 3 | Complete |
| VAL-03 | Phase 3 | Complete |
| OBS-01 | Phase 3 | Complete |
| OBS-02 | Phase 3 | Complete |
| OBS-03 | Phase 3 | Complete |
| TEST-01 | Phase 4 | Pending |
| TEST-02 | Phase 4 | Pending |
| TEST-03 | Phase 4 | Pending |
| TEST-04 | Phase 4 | Pending |
| TEST-05 | Phase 4 | Pending |
| TEST-06 | Phase 4 | Pending |

**Coverage:**
- v1 requirements: 36 total
- Mapped to phases: 36
- Unmapped: 0 ✓

---
*Requirements defined: 2026-04-15*
*Last updated: 2026-04-15 after roadmap creation (traceability mapped to phases)*