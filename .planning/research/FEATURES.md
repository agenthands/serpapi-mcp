# Feature Landscape

**Domain:** MCP (Model Context Protocol) server — search API gateway
**Researched:** 2026-04-15

## Table Stakes

Features users expect. Missing = product feels incomplete or untrustworthy.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Test suite** | Production servers must have tests. Zero tests currently. CI only checks formatting. | Medium | FastMCP provides `Client` for in-memory testing — no HTTP server needed. Mock `serpapi.search()` for unit tests, use real client for integration tests. Cover: auth middleware, search tool (valid/invalid engine, valid/invalid mode, error responses), resource loading, healthcheck. |
| **Type annotations** | `mypy` config requires `disallow_untyped_defs` but `server.py` has zero annotations. CI can't enforce types. | Low | Mechanical work — add type hints to all function signatures. FastMCP already uses typed patterns. Unblocks `mypy` in CI. |
| **Consistent error handling** | Current error handling uses string matching on exception messages (`"429" in str(e)`). Fragile, doesn't distinguish error classes, returns plain error strings not MCP error responses. | Medium | 1. Replace string-matching with proper exception handling (catch `serpapi.exceptions.HTTPError` by status code). 2. Use FastMCP's `ErrorHandlingMiddleware` for consistent error → MCP error response conversion. 3. Use `mask_error_details=True` in production to avoid leaking internals. |
| **Input validation** | `search` tool accepts `params: dict[str, Any] = {}` with zero validation. Invalid engine names, wrong parameter types, and missing required fields all get forwarded to SerpApi producing confusing errors. | Medium | Validate engine name against loaded schemas. Validate required parameters per engine. Return structured MCP error responses with actionable messages. The engine JSON files contain type/required info already. |
| **CI pipeline beyond format** | Only `uv format --check` runs on PRs. No tests, no type checking, no linting. Broken code can reach main. | Low | Add `pytest`, `mypy`, and `flake8` steps to the `format_check.yml` workflow (or create a separate `ci.yml`). ~20 lines of YAML. |
| **Graceful startup validation** | If `engines/` directory is missing or contains malformed JSON, the server starts but serves broken resources. No validation at startup. | Low | Add a startup check: verify `ENGINES_DIR` exists, JSON parses, required fields present. Fail fast with clear error message. |

## Differentiators

Features that set product apart. Not expected, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **MCP Prompts** | Pre-built search prompt templates let AI agents use search more effectively. E.g., "research topic" prompt structures a multi-query research flow; "competitive analysis" prompt defines comparison parameters. | Low | FastMCP supports `@mcp.prompt()` decorator. Define 5-8 curated prompts that demonstrate effective search patterns. Directly improves LLM agent experience. |
| **MCP Completions** | Auto-completion for engine names and parameter values. Agents discover valid options interactively instead of guessing. | Medium | MCP spec has a `completions` capability. FastMCP supports it. Complete engine names, location values, and enum-type parameters. Requires wiring engine schema data into completion handlers. |
| **Progress reporting** | For complex queries, report progress back to the MCP client. Gives users visibility that something is happening. | Low | FastMCP `Context` object provides `ctx.report_progress(progress, total)`. Straightforward to add to the search tool. Most beneficial for slower engines (Google full, Google Flights, etc.). |
| **MCP Context logging** | Send structured log messages to MCP clients (not just CloudWatch). Agents can see what happened during search. | Low | `ctx.info()`, `ctx.debug()`, `ctx.warning()` in FastMCP Context. Replace some `logger.info()` calls with MCP Context logging. Dual-channel: CloudWatch for ops, MCP logging for client. |
| **Tool `listChanged` notification** | When `build-engines.py` produces new engine schemas, notify connected clients that the tool/tool schema has changed. | Medium | MCP spec supports `notifications/tools/list_changed`. Requires maintaining state about engine schema changes. Most useful during dev/staging when schemas update frequently. Lower priority for hosted service. |
| **Structured logging middleware** | Replace ad-hoc `logger.info(json.dumps(emf_event))` with FastMCP's `StructuredLoggingMiddleware`. Machine-readable JSON logs for CloudWatch/ELK. | Low | FastMCP provides this as a built-in middleware. One-line swap. Coexists with existing `RequestMetricsMiddleware`. |
| **Request correlation IDs** | Assign a unique ID per request for tracing across logs, metrics, and error responses. | Low | Starlette middleware to generate/propagate `X-Request-ID`. Include in all log entries and error responses. Essential for production debugging. |
| **Engine schema validation on startup** | Validate that all `engines/*.json` files conform to an expected schema (have required fields, no unexpected types). | Low | Load and validate each file at startup. Use JSON Schema or simple structural checks. Fail fast vs. serving broken resources. Already listed as an active requirement. |

## Anti-Features

Features to explicitly NOT build.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **WebSocket/SSE transport** | Already scoped out. Streamable HTTP is the current MCP standard. SSE is legacy. | Maintain Streamable HTTP only. |
| **Multi-tenant API key management** | Each request carries its own key. No need for server-side key storage, rotation, or scopes. API key is a passthrough to SerpApi. | Keep path/header auth. SerpApi manages key lifecycle. |
| **Result caching** | SerpApi handles caching server-side. Client-side caching would serve stale results and add complexity (invalidation, TTL). | Let SerpApi cache. Clients can cache if they want. |
| **Server-side rate limiting** | SerpApi enforces per-key rate limits. Duplicating adds latency and config surface. | Rely on SerpApi's rate limiting. Return SerpApi rate limit errors to clients. |
| **OAuth authentication flow** | The API key model is simple, well-understood, and matches SerpApi's own auth. Adding OAuth (PKCE, token exchange, refresh) is massive complexity for no real benefit. | Keep API key auth. Path-based for hosted, header-based for programmatic clients. |
| **MCP Resource subscriptions** | Engine schemas change only across deployments, not during a running session. Adding subscribe/listChanged for resources adds complexity with no practical benefit. | Reload server when schemas change. Don't implement resource subscription. |
| **MCP Sampling** | Letting the server request LLM completions from the client for query refinement would add massive complexity and create latency loops. Search queries are direct — agents compose them, not the server. | Skip sampling. The server is a stateless search proxy. |
| **Database / persistent storage** | Adding a database for caching, analytics, or session state contradicts the stateless proxy design and adds operational overhead. | Stay stateless. CloudWatch EMF for metrics. No server-side storage. |
| **Custom MCP tasks** | The MCP spec has `tasks` (experimental). They add durable execution semantics not needed for a stateless search proxy. | Keep the single `search` tool. No task management. |

## Feature Dependencies

```
Type annotations ────────────────────────────────────┐
                                                      │
Test suite (foundational) ──→ Error handling refactor ─┤
                  │                                   │
                  └──→ Input validation ───────────────┤
                                                      │
CI pipeline improvements ─→ all of the above ─────────┤
                                                      │
Startup validation (standalone) ──────────────────────┤
                                                      │
FastMCP Context adoption ──→ Progress reporting       │
                           ──→ MCP Context logging     │
                                                      │
MCP Prompts (standalone) ─────────────────────────────┤
                                                      │
MCP Completions ──→ requires engine schema adoption    │
                                                      │
Structured logging middleware (standalone) ───────────┤
                                                      │
Request correlation IDs (standalone) ──────────────────┘

Critical path: Test suite → Error handling → Input validation
               (must be done in this order)
```

## MVP Recommendation

Prioritize:
1. **Test suite** — Foundation for everything else; can't ship production without it
2. **Type annotations** — Unblocks mypy in CI; mechanical but necessary
3. **CI pipeline improvements** — Add pytest, mypy, flake8 to CI; prevents regressions
4. **Error handling** — Replace string-matching with proper exception hierarchy

Defer:
- **MCP Prompts**: Low-hanging fruit but not table stakes; do after core quality
- **MCP Completions**: Medium effort; requires schema wiring; nice-to-have
- **Progress reporting**: Only useful for slow engines; low priority
- **Tool `listChanged` notifications**: Low value for hosted service; skip for now

## Sources

- FastMCP documentation (Context7, `/prefecthq/fastmcp`): testing patterns, middleware, Context object, prompt/completion support — HIGH confidence
- MCP specification (Context7, `/modelcontextprotocol/modelcontextprotocol`): ServerCapabilities (logging, completions, prompts, resources, tools, tasks), notifications — HIGH confidence
- Existing codebase analysis: `server.py`, `pyproject.toml`, `build-engines.py` — HIGH confidence
- MCP ecosystem patterns (Smithery, server.json registries): observed standard patterns — MEDIUM confidence