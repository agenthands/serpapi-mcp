# Requirements: SerpApi MCP Server

**Defined:** 2026-04-15
**Core Value:** AI agents can search any SerpApi-supported engine through a single, authenticated MCP endpoint with structured parameter discovery.

## v1 Requirements

Requirements for hardening the existing server. Each maps to roadmap phases.

### Testing

- [ ] **TEST-01**: Unit tests for search tool using FastMCP Client (in-memory, no HTTP)
- [ ] **TEST-02**: Unit tests for engine resource loading and schema retrieval via FastMCP Transport
- [ ] **TEST-03**: Integration tests for API key authentication middleware via Starlette TestClient
- [ ] **TEST-04**: Integration tests for healthcheck endpoint
- [ ] **TEST-05**: Unit tests for error handling paths (rate limit, invalid key, server errors) using responses mock
- [ ] **TEST-06**: Unit tests for compact mode field removal
- [ ] **TEST-07**: Unit tests for emit_metric CloudWatch EMF format
- [ ] **TEST-08**: Refactor module-level engine registration into a callable `register_engines()` function to enable test isolation

### Type Safety

- [ ] **TYPE-01**: Add type annotations to all functions in server.py (mypy `disallow_untyped_defs` compliance)
- [ ] **TYPE-02**: Add type annotations to all functions in build-engines.py
- [ ] **TYPE-03**: mypy strict mode passes with zero errors on the entire codebase

### Error Handling

- [ ] **ERR-01**: Replace string-prefix error returns (`"Error: ..."`) with FastMCP `ToolError` exceptions for MCP-compliant error signaling
- [ ] **ERR-02**: Replace string-matching on exception messages (`"429" in str(e)`) with proper exception type and status code checking
- [ ] **ERR-03**: Wrap blocking sync `serpapi.search()` call in `asyncio.to_thread()` to prevent event loop blocking
- [ ] **ERR-04**: Fix mutable default argument `params: dict[str, Any] = {}` to `params: dict[str, Any] | None = None`
- [ ] **ERR-05**: Consistent error response format across all error paths

### Validation

- [ ] **VAL-01**: Validate `params` input against engine schema on startup (fail fast on corrupt/missing engine JSON)
- [ ] **VAL-02**: Validate `mode` parameter accepts only "complete" or "compact" at the tool input schema level (not in function body)
- [ ] **VAL-03**: Structural validation of engine JSON files (required keys: engine, params, common_params)

### Observability

- [ ] **OBS-01**: Add request correlation IDs to middleware chain for traceability across logs
- [ ] **OBS-02**: Replace `logger.info(json.dumps(emf_event))` with FastMCP Context logging (`ctx.info()` / `ctx.debug()`)
- [ ] **OBS-03**: Add `ErrorHandlingMiddleware` from FastMCP for uncaught exception handling

### CI

- [ ] **CI-01**: Replace flake8 with ruff in dev dependencies (zero config migration, 10-100x faster)
- [ ] **CI-02**: Upgrade pytest-asyncio to >=0.24 for stable `asyncio_mode = "auto"` support
- [ ] **CI-03**: Add mypy check to CI pipeline
- [ ] **CI-04**: Add pytest to CI pipeline

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### MCP Features

- **MCP-01**: MCP Prompts for curated search templates (e.g., "search news about {topic}")
- **MCP-02**: Progress reporting via `ctx.report_progress()` for long-running searches
- **MCP-03**: Sampling capability for client-side result processing

### Infrastructure

- **INFRA-01**: Structured logging middleware (FastMCP's `StructuredLoggingMiddleware`)
- **INFRA-02**: Graceful shutdown handling
- **INFRA-03**: Response compression middleware
- **INFRA-04**: Request timeout configuration

## Out of Scope

| Feature | Reason |
|---------|--------|
| WebSocket/SSE transport | Streamable HTTP is the current transport; MCP spec is moving toward Streamable HTTP |
| Multi-tenant API key management | Each request carries its own key; no server-side key store needed |
| Result caching | SerpApi caches on their side; duplicating adds complexity with no value |
| Rate limiting on server side | SerpApi handles rate limiting per API key; server-side limiting adds latency |
| OAuth/API key storage | Auth is stateless; no database or session store |
| Admin dashboard | Operations handled via AWS CloudWatch and Copilot |
| Database or persistent storage | Server is stateless by design |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| TEST-01 | — | Pending |
| TEST-02 | — | Pending |
| TEST-03 | — | Pending |
| TEST-04 | — | Pending |
| TEST-05 | — | Pending |
| TEST-06 | — | Pending |
| TEST-07 | — | Pending |
| TEST-08 | — | Pending |
| TYPE-01 | — | Pending |
| TYPE-02 | — | Pending |
| TYPE-03 | — | Pending |
| ERR-01 | — | Pending |
| ERR-02 | — | Pending |
| ERR-03 | — | Pending |
| ERR-04 | — | Pending |
| ERR-05 | — | Pending |
| VAL-01 | — | Pending |
| VAL-02 | — | Pending |
| VAL-03 | — | Pending |
| OBS-01 | — | Pending |
| OBS-02 | — | Pending |
| OBS-03 | — | Pending |
| CI-01 | — | Pending |
| CI-02 | — | Pending |
| CI-03 | — | Pending |
| CI-04 | — | Pending |

**Coverage:**
- v1 requirements: 23 total
- Mapped to phases: 0
- Unmapped: 23 ⚠️

---
*Requirements defined: 2026-04-15*
*Last updated: 2026-04-15 after initial definition*