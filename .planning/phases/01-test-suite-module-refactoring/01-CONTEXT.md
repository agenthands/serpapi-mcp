# Phase 1: Test Suite & Module Refactoring - Context

**Gathered:** 2026-04-15
**Status:** Ready for planning

<domain>
## Phase Boundary

Establish a working test suite covering core functionality (search tool, engine resources, auth middleware, healthcheck, error paths, metrics, compact mode) and refactor module-level side effects (engine registration loop) into testable functions. Fix mutable default argument. Add pytest to CI.

This phase delivers the testing foundation that all subsequent hardening phases depend on.

</domain>

<decisions>
## Implementation Decisions

### Testing Approach
- **D-01:** Dual-layer testing — FastMCP `Client` for unit tests (search tool, engine resources, compact mode, metrics), Starlette `TestClient` for integration tests (auth middleware, healthcheck endpoint)
- **D-02:** Mock SerpApi HTTP calls using `responses` library (intercepts `requests` calls made by `serpapi` Python client, not `httpx`)
- **D-03:** Test `ApiKeyMiddleware` via Starlette `TestClient` since `get_http_request()` requires HTTP context
- **D-04:** Test engine resources via FastMCP `Client` with in-memory `FastMCPTransport`

### Module Refactoring
- **D-05:** Extract module-level engine registration loop into a callable `register_engines(mcp, engines_dir)` function to enable test isolation
- **D-06:** Create `create_app()` factory function that constructs the Starlette app with all middleware, for use in test fixtures
- **D-07:** Keep `main()` as the entry point that calls `create_app()` — minimal structural change
- **D-08:** Single-file `src/server.py` structure remains — no package split at this phase

### Test Organization
- **D-09:** Pytest files in `tests/` directory at project root with `conftest.py` providing shared fixtures
- **D-10:** `pytest-asyncio` with `asyncio_mode = "auto"` (requires >=0.24 per STACK.md research)
- **D-11:** Test files: `test_search.py` (search tool unit tests), `test_engines.py` (resource tests), `test_middleware.py` (auth integration), `test_healthcheck.py`, `test_metrics.py`, `test_compact_mode.py`

### CI Design
- **D-12:** Add new GitHub Actions workflow for PR checks running pytest, mypy, and ruff
- **D-13:** Keep existing `format_check.yml` (uv format check) — new workflow adds quality gates, not replaces format check
- **D-14:** CI should fail on any test failure, mypy error, or ruff lint error

### Mutable Default Fix
- **D-15:** Change `params: dict[str, Any] = {}` to `params: dict[str, Any] | None = None` with `if params is None: params = {}` inside function body

### Agent's Discretion
- Test fixture details (what exactly to mock, how many test cases per feature)
- Error message format in tests (assert vs pytest.raises patterns)
- conftest.py fixture naming conventions

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Research
- `.planning/research/STACK.md` — Technology recommendations with rationale (FastMCP testing patterns, pytest-asyncio, responses library, ruff)
- `.planning/research/FEATURES.md` — Feature landscape including testing strategy features
- `.planning/research/ARCHITECTURE.md` — Architecture patterns including app factory and module structure recommendations
- `.planning/research/PITFALLS.md` — Domain pitfalls including module-level registration blocking tests, `get_http_request()` context requirement
- `.planning/research/SUMMARY.md` — Synthesized research with phase ordering rationale

### Project
- `.planning/PROJECT.md` — Project context, constraints, validated requirements
- `.planning/REQUIREMENTS.md` — v1 requirements with REQ-IDs (TEST-01 through TEST-08, ERR-04, CI-02, CI-04)
- `.planning/ROADMAP.md` — Phase 1 goal, success criteria, and plan outlines

### Source Code
- `src/server.py` — Primary implementation file (330 lines, all logic in one file)
- `src/__init__.py` — Empty package init
- `build-engines.py` — Engine schema generator (not in scope for this phase)
- `engines/*.json` — Generated engine schemas (not in scope, but tests should verify they load correctly)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `FastMCP` instance (`mcp` variable) — can be imported and used in tests via `Client(mcp)` for in-memory testing
- `_get_engine_files()` — already a helper function (not module-level), can be reused
- `_engine_resource_factory()` — factory pattern for creating engine resources
- `emit_metric()` — standalone utility function, testable in isolation

### Established Patterns
- Single-file architecture in `src/server.py` — all logic in one module
- Module-level side effects: engine registration loop (`for _engine_path in _get_engine_files():`) runs at import time
- `ApiKeyMiddleware` extracts key from path or header, stores in `request.state.api_key`
- `RequestMetricsMiddleware` emits CloudWatch EMF format via `logger.info(json.dumps(...))`
- `search` tool uses `get_http_request()` to access `request.state.api_key`

### Integration Points
- `main()` constructs Starlette app via `mcp.http_app()`, adds `healthcheck` route, runs with uvicorn
- `ENGINES_DIR` resolves relative to `src/server.py` using `Path(__file__).resolve().parents[1] / "engines"`
- Auth middleware strips API key from path before forwarding to MCP handler
- `serpapi.search()` is a synchronous blocking call that must not block the event loop in production

</code_context>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches for Python MCP server testing.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 01-test-suite-module-refactoring*
*Context gathered: 2026-04-15*