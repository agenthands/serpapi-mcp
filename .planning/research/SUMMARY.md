# Project Research Summary

**Project:** SerpApi MCP Server — Hardening & Extension
**Domain:** MCP (Model Context Protocol) server — AI agent search gateway
**Researched:** 2026-04-15
**Confidence:** HIGH

## Executive Summary

The SerpApi MCP Server is a stateless Python proxy that exposes SerpApi search capabilities to AI agents via the Model Context Protocol. It uses FastMCP on Starlette/uvicorn with Streamable HTTP transport and path-based API key authentication. The server currently functions but has zero test coverage, no type annotations (despite a mypy strict config), fragile string-based error handling, and no input validation — making it unsuitable for production reliability.

The recommended approach follows the critical dependency path: establish a test suite first (which requires refactoring module-level side effects), then add type annotations and CI enforcement, then overhaul error handling to use MCP's native `ToolError` and proper exception hierarchy, and finally layer on input validation and startup checks. This order is forced by dependencies — you can't safely refactor error handling without tests, and you can't write tests against the current module-level side effects.

The key risk is the `get_http_request()` dependency in the search tool, which prevents in-process testing with FastMCP's recommended `Client` pattern. The correct mitigation is dual test layers: `Starlette TestClient` for HTTP-dependent tests (auth, middleware, search tool with real request context) and `FastMCPTransport Client` for in-memory tests (resource listing, prompt logic). Additionally, `serpapi` uses `requests` internally — not `httpx` — so mocking must use the `responses` library, not `pytest-httpx`.

## Key Findings

### Recommended Stack

The existing core stack (FastMCP, Starlette, uvicorn, serpapi) is solid and well-documented. Three new dependencies are needed: `responses` (dev, for mocking serpapi's `requests` calls), `jsonschema` (prod, for engine schema validation at startup), and `ruff` (dev, replacing flake8). The critical testing insight is that FastMCP provides a `Client` class for in-memory testing, but the search tool cannot use this pattern because it relies on HTTP request context for API key extraction.

**Core technologies:**
- **FastMCP >=2.13**: MCP server framework — best Python MCP framework, provides `Client` testing, Context, middleware
- **serpapi >=0.1.5**: SerpApi client — uses `requests` internally (NOT httpx), critical for mock strategy
- **responses >=0.25.0** (new): Mock library — intercepts `requests` calls for realistic serpapi mocking
- **jsonschema >=4.23.0** (new): Validate engine JSON — lightweight startup validation
- **ruff >=0.8.0** (new): Linter — replaces flake8, 10-100x faster, auto-fix
- **pytest-asyncio >=0.24.0** (upgrade): Async test runner — stable `auto` mode

### Expected Features

**Must have (table stakes):**
- **Test suite** — Zero tests currently; production servers need tests; FastMCP Client + TestClient provide clear testing patterns
- **Type annotations** — mypy config requires `disallow_untyped_defs` but server.py has zero annotations; mechanical but necessary
- **Consistent error handling** — Current approach uses string matching on exception messages; replace with `ToolError` and exception hierarchy
- **Input validation** — `search` tool accepts `params: dict[str, Any] = {}` with zero validation; validate engine names and required params
- **CI pipeline** — Only format check runs on PRs; add pytest, mypy, ruff steps
- **Startup validation** — No validation of engine JSON files; fail fast on malformed schemas

**Should have (competitive):**
- **MCP Prompts** — Pre-built search prompt templates; low effort, high value for LLM agent UX
- **Request correlation IDs** — Unique ID per request for cross-log tracing
- **Structured logging middleware** — FastMCP provides this as a one-line swap
- **MCP Context logging** — Dual-channel logging (CloudWatch for ops, MCP for clients)
- **Progress reporting** — `ctx.report_progress()` for slow engines

**Defer (v2+):**
- **MCP Completions** — Medium effort; requires wiring engine schemas into completion handlers
- **Tool `listChanged` notifications** — Low value for hosted service; schemas change only across deployments
- **Caching, rate limiting, OAuth, database** — All explicitly scoped out as anti-features

### Architecture Approach

The architecture follows a middleware-to-handler pipeline: Starlette handles HTTP with a middleware stack (correlation ID → metrics → auth → CORS), then FastMCP handles MCP protocol routing to tools/resources/prompts. The search tool is the centerpiece: validate params → inject API key from request state → call serpapi → apply mode filter → return JSON. Engine schemas are loaded at startup and served as MCP resources. All state is ephemeral — no database, no sessions.

**Major components:**
1. **Middleware Stack** (CorrelationIdMiddleware → Metrics → Auth → CORS) — request processing, auth, observability
2. **FastMCP Instance** — registers tools, resources, prompts; handles MCP JSON-RPC protocol
3. **Search Tool** — validate params, inject API key, call serpapi, filter results, handle errors
4. **Engine Schema System** — load/validate `engines/*.json`, serve as MCP resources
5. **Observability** — CloudWatch EMF metrics, structured logging, correlation IDs, healthcheck

### Critical Pitfalls

1. **`get_http_request()` blocks in-process testing** — The search tool extracts API keys from HTTP request context, making `FastMCPTransport Client` tests fail. Must use `Starlette TestClient` for HTTP-dependent tests and refactor `search` to support explicit `api_key` parameter.
2. **Module-level engine registration prevents test isolation** — `for _engine_path in _get_engine_files():` runs at import time, causing disk I/O and state pollution. Must extract into `register_engines()` function called from `main()`.
3. **String-prefixed errors break MCP protocol** — `return "Error: ..."` doesn't set the `isError` flag, causing AI agents to interpret errors as successful results. Must use `ToolError` exception class.
4. **Mutable default `params={}`** — Classic Python bug; shared dict across calls. Replace with `params=None`.
5. **Blocking `serpapi.search()` in async handler** — Synchronous call blocks the event loop. Wrap in `asyncio.to_thread()`.

## Implications for Roadmap

### Phase 1: Test Suite & Module Refactoring
**Rationale:** Tests are the foundation for all subsequent refactoring. Module-level side effects must be removed before the server can be imported in test fixtures. The `get_http_request()` dependency must be understood and handled from day one.
**Delivers:** Working test suite with correct patterns (TestClient for HTTP tests, FastMCPTransport for resource tests), refactored module structure with `register_engines()` function, fixed mutable defaults.
**Addresses:** Test suite, CI pipeline (partially — add pytest step)
**Avoids:** Pitfalls #1 (in-process testing pattern), #2 (module-level side effects), #4 (mutable defaults)

### Phase 2: Type Annotations & CI Hardening
**Rationale:** With tests in place, type annotations are safe mechanical work. Add mypy and ruff to CI to prevent regressions. Move `load_dotenv()` into `main()`.
**Delivers:** Fully annotated `server.py`, mypy passing in CI, ruff replacing flake8, complete CI pipeline.
**Uses:** ruff, pytest, mypy — all now verified by CI
**Avoids:** Pitfall: `load_dotenv()` module-level side effects

### Phase 3: Error Handling & MCP Protocol Compliance
**Rationale:** Replace fragile string-matching error handling with proper exception hierarchy. Use `ToolError` for MCP error responses. Add `asyncio.to_thread()` for blocking serpapi calls. Fix CORS configuration.
**Delivers:** Structured error handling with typed exceptions, `isError=True` on tool failures, non-blocking serpapi calls, correct CORS config.
**Avoids:** Pitfalls #3 (string errors), #5 (blocking async), security issue (CORS credentials+wildcard)

### Phase 4: Input Validation & Startup Checks
**Rationale:** With error handling in place, validate search params against engine schemas. Add jsonschema validation at startup. Fail fast on malformed engine JSON.
**Delivers:** Engine name validation, required parameter checking, jsonschema startup validation, clear error messages for invalid inputs.
**Uses:** jsonschema for engine validation, engine schema data for param validation

### Phase 5: Observability & MCP Enhancements
**Rationale:** Quality-of-life improvements that build on the solid foundation. Correlation IDs, structured logging, MCP Context logging, progress reporting, curated MCP prompts.
**Delivers:** Request correlation IDs, FastMCP `StructuredLoggingMiddleware`, `ctx.info()`/`ctx.debug()` logging, `ctx.report_progress()`, 5-8 search prompt templates.
**Uses:** FastMCP Context, FastMCP middleware

### Phase Ordering Rationale

- **Phase 1 must come first** because module-level side effects prevent any test from importing the server cleanly, and `get_http_request()` mandates specific testing patterns that all subsequent tests must follow
- **Phase 2 before Phase 3** because mypy type annotations make error handling code safer and more discoverable — `serpapi.exceptions.HTTPError` annotations guide proper exception handling
- **Phase 3 before Phase 4** because input validation errors need proper `ToolError` responses — can't validate inputs correctly until error handling uses MCP protocol properly
- **Phase 5 last** because observability and prompts are additive features that don't affect core correctness

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 3:** SerpApi Python client exception structure needs verification — the `serpapi.exceptions.HTTPError` response attribute behavior varies; should verify during planning
- **Phase 4:** Engine JSON schema structure may have edge cases not fully documented — validate against actual `engines/*.json` files during planning

Phases with standard patterns (skip research-phase):
- **Phase 1:** Well-documented testing patterns from FastMCP and Starlette; `responses` library is straightforward
- **Phase 2:** Mechanical type annotation work; mypy and ruff are industry-standard
- **Phase 5:** FastMCP Context and middleware APIs are well-documented

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | FastMCP, Starlette, serpapi APIs verified via Context7 and codebase inspection. `responses` library and jsonschema are ecosystem standards. |
| Features | HIGH | Table stakes are derived from codebase gaps (zero tests, no types, string error handling). Differentiators are well-documented FastMCP features. Anti-features are clearly scoped out. |
| Architecture | HIGH | Patterns based on FastMCP docs and Starlette conventions. Component boundaries are straightforward for a stateless proxy. |
| Pitfalls | HIGH | Top 5 pitfalls verified against FastMCP docs and codebase inspection. `get_http_request()` issue confirmed via API docs. Module-level side effects visible in `server.py`. |

**Overall confidence:** HIGH

### Gaps to Address

- **SerpApi exception structure:** The `serpapi.exceptions.HTTPError` response attribute and status code access pattern needs verification during Phase 3 planning. Research indicated "LOW confidence" on this specific point — should read the actual `serpapi` package source.
- **FastMCP `ToolError` behavior in Streamable HTTP transport:** Confirmed that `ToolError` sets `isError=True`, but should verify during Phase 1 that the error message formatting is correct when transmitted via Streamable HTTP.
- **Engine JSON schema variability:** The `build-engines.py` script generates schemas from the SerpApi playground. Schema structure might vary across engines more than expected. Validate during Phase 4 planning by checking a representative sample.

## Sources

### Primary (HIGH confidence)
- FastMCP documentation (Context7, `/prefecthq/fastmcp`) — testing patterns, Client API, Context object, middleware, lifespan, ToolError
- MCP specification (Context7, `/modelcontextprotocol/modelcontextprotocol`) — ServerCapabilities, isError flag, transport types
- Starlette TestClient documentation (Context7, `/kludex/starlette`) — sync and async test patterns
- pytest-asyncio documentation (Context7, `/pytest-dev/pytest-asyncio`) — asyncio auto mode configuration
- `responses` library (Context7, `/getsentry/responses`) — mocking `requests` library calls
- Existing codebase (`src/server.py`, `pyproject.toml`, `build-engines.py`) — current architecture, error patterns, middleware order

### Secondary (MEDIUM confidence)
- Starlette `BaseHTTPMiddleware` known issues — community documentation, commonly reported
- FastMCP `http_app` configuration options — verified via installed package inspection

### Tertiary (LOW confidence)
- SerpApi Python client exception structure — training data, needs verification against actual package source during Phase 3

---
*Research completed: 2026-04-15*
*Ready for roadmap: yes*