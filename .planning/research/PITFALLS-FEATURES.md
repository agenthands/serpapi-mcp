# Domain Pitfalls (Features Dimension)

**Domain:** MCP server — search API gateway
**Researched:** 2026-04-15
**Note:** This covers pitfalls specific to MCP feature adoption. See PITFALLS.md for testing, typing, and error handling pitfalls.

## Critical Pitfalls

### Pitfall 1: `stateless_http=True` Silently Disables MCP Features
**What goes wrong:** The server is created with `mcp.http_app(stateless_http=True)`. This tells FastMCP not to maintain session state. Any MCP feature that depends on session persistence — `ctx.get_state()`/`ctx.set_state()`, resource subscriptions, server-initiated notifications — will silently fail or behave as no-ops.
**Why it happens:** Stateless HTTP is the simplest deployment model and matches the server's stateless proxy design. Developers may not realize which MCP capabilities require session state.
**Consequences:** Progress reporting (`ctx.report_progress()`) works per-request but doesn't accumulate across calls. Session state features are effectively broken. Adding prompts that rely on conversation context won't work.
**Prevention:** Document explicitly which MCP features work and don't work in `stateless_http=True` mode. Don't add features that require session state (subscriptions, cross-call state). Progress reporting is safe because it's per-request.
**Detection:** Features that seem to work in development (with a browser-based test client maintaining a session) but fail in production (where each HTTP request is independent).

### Pitfall 2: `get_http_request()` Blocks In-Memory Testing
**What goes wrong:** The `search` tool calls `get_http_request()` to extract the API key from the Starlette request context. This context only exists during real HTTP requests. FastMCP's `Client(mcp)` in-memory testing pattern has no HTTP request, so every tool call fails with "Error: Unable to access API key from request context."
**Why it happens:** FastMCP's testing docs promote `Client(transport=mcp)` for testing, but this only works for tools that don't depend on HTTP request context.
**Consequences:** Cannot write unit tests for the search tool using the fast, in-memory Client pattern. Must use Starlette TestClient or httpx for integration tests, which is slower and more complex.
**Prevention:** Two approaches: (1) Refactor `search` to accept `api_key` as an explicit parameter with request context as a fallback, enabling unit tests with FastMCP Client. (2) Use Starlette TestClient / httpx AsyncClient for integration tests that exercise auth middleware.
**Detection:** Test failures with "Error: Unable to access API key from request context" when using `Client(mcp)`.

### Pitfall 3: String-Prefixed Errors Instead of MCP Protocol Errors
**What goes wrong:** The `search` tool returns errors as plain strings prefixed with `"Error:"` (e.g., `"Error: Rate limit exceeded."`). MCP has a dedicated `isError` flag on tool results. When `isError=True`, clients display errors properly. String returns are displayed as if the search succeeded.
**Why it happens:** `return f"Error: {msg}"` is the simplest approach and appears to work — the string comes back and shows the error text.
**Consequences:** AI agents interpret error strings as successful search results. They may try to parse `"Error: Rate limit exceeded"` as a search result. Clients can't programmatically distinguish errors from successful responses.
**Prevention:** Use FastMCP's `ToolError` exception class: `raise ToolError("Rate limit exceeded")`. This sets `isError=True` in the MCP protocol response.
**Detection:** MCP Inspector showing `200 OK` with `isError: false` for clearly failed requests. AI agents trying to interpret error strings.

### Pitfall 4: MCP Features Added Without Client Adoption Research
**What goes wrong:** Implementing MCP prompts, completions, or notifications without verifying that major MCP clients (Claude Desktop, Cursor, etc.) actually consume these capabilities. Some MCP capabilities are spec-defined but not yet widely supported by clients.
**Why it happens:** The MCP spec defines a rich set of capabilities. It's tempting to implement all of them. Without client adoption research, you build features nobody uses.
**Consequences:** Wasted development effort. Features that are implemented but never triggered by any client. Maintenance burden for code paths that aren't exercised in production.
**Prevention:** Research MCP client adoption before implementing features. Currently: Tools and Resources are universally supported. Prompts are supported by Claude Desktop. Completions and Tasks have limited client support. Sampling is rarely supported. Prioritize features that major clients will actually consume.
**Detection:** GitHub issues or user requests asking for specific MCP features. Check MCP client release notes for capability support.

## Moderate Pitfalls

### Pitfall 5: Blocking Sync `serpapi.search()` in Async Handler
**What goes wrong:** `serpapi.search()` is a synchronous blocking call. Running it directly in an async tool handler blocks the ASGI event loop. Under concurrent load (10+ simultaneous requests), one slow SerpApi call blocks all other requests.
**Why it happens:** The `serpapi` Python package provides a sync API. Using it directly in an `async def` tool is the simplest approach.
**Prevention:** Wrap in `asyncio.to_thread()`: `data = await asyncio.to_thread(serpapi.search, search_params)`. This offloads the blocking call to a thread pool.
**Detection:** Slow response times when multiple concurrent requests hit the server. Event loop stalls visible in metrics.

### Pitfall 6: FastMCP Version Pinning Too Loose
**What goes wrong:** `pyproject.toml` specifies `fastmcp>=2.13.0.2`. This allows any 2.x version including breaking changes. FastMCP is under active development and its API has been evolving.
**Why it happens:** It's common to specify minimum versions without upper bounds.
**Prevention:** Pin to a minor version range: `fastmcp>=2.13.0,<3.0.0` or use `~=2.13`. Test against the pinned version in CI.
**Detection:** `pip install -U fastmcp` breaking the server after a FastMCP release.

### Pitfall 7: Engine Schema Validation Too Strict or Too Loose
**What goes wrong:** If you validate engine schemas strictly (only known fields, specific value types), adding a new SerpApi engine or parameter breaks validation. If you validate too loosely (no validation), malformed schemas serve garbage to clients.
**Why it happens:** The right level of validation is somewhere between "none" and "JSON Schema strict mode." It's hard to know where the line is when SerpApi's schema is external and evolving.
**Prevention:** Validate structure only: required keys exist (`engine`, `params`, `common_params`), types are correct (dicts, lists, strings), but allow unknown additional keys with `additionalProperties: true`. Never validate specific parameter names or values.
**Detection:** `build-engines.py` producing valid schemas that fail overly strict validation.

### Pitfall 8: `datetime.utcnow()` Deprecation in Healthcheck
**What goes wrong:** The healthcheck handler uses `datetime.utcnow().isoformat() + "Z"`, which is deprecated in Python 3.12+. It produces naive datetime objects that may cause timezone issues.
**Prevention:** Use `datetime.now(timezone.utc).isoformat()` instead. Python 3.13 is the project's target version.
**Detection:** Deprecation warnings in logs (if Python is configured to show them).

## Minor Pitfalls

### Pitfall 9: CORS `allow_origins=["*"]` with `allow_credentials=True`
**What goes wrong:** The current CORS config has both `allow_origins=["*"]` and `allow_credentials=True`. Per the CORS specification, browsers reject responses with wildcard origin AND credentials. This combination is a no-op in browsers.
**Prevention:** Either use `allow_origins=["*"]` with `allow_credentials=False` (acceptable for a public MCP API), or specify explicit origins with `allow_credentials=True`.
**Detection:** Browser-based clients failing CORS preflight checks.

### Pitfall 10: Healthcheck Doesn't Verify Downstream
**What goes wrong:** The healthcheck returns `{"status": "healthy"}` without checking if engine schemas loaded or SerpApi is reachable. A "healthy" server could be unable to serve search results.
**Prevention:** Keep the healthcheck shallow (just confirms process is running). Don't check SerpApi reachability — that would cascade SerpApi outages into healthcheck failures, which could trigger unnecessary restarts.
**Detection:** Healthcheck returning "healthy" while all searches fail.

### Pitfall 11: API Key Exposure in Logs
**What goes wrong:** API keys are in URL paths (`/{KEY}/mcp`). If logging middleware logs full URLs or request headers, keys leak into logs. The `ApiKeyMiddleware` rewrites `request.scope["path"]` but the original path may appear in access logs.
**Prevention:** Ensure logging middleware redacts API keys. Configure uvicorn access logging to not log full paths.
**Detection:** Access logs containing API key strings.

### Pitfall 12: Smithery Config References Wrong Path
**What goes wrong:** `smithery.yaml` has `commandFunction` pointing to `src/serpapi-mcp-server/server.py`, but the actual entry point is `src/server.py`. This could break Smithery deployment.
**Prevention:** Verify and fix the path in `smithery.yaml` to point to `src/server.py`.
**Detection:** Smithery deployment failing to find the server module.

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| Test suite | `get_http_request()` blocks in-memory testing | Use Starlette TestClient for HTTP-dependent tests; FastMCP Client for resource-only tests |
| Test suite | Module-level side effects prevent test isolation | Refactor engine registration into explicit function called from `main()` |
| Error handling | String-prefixed errors instead of `ToolError` | Use `raise ToolError(msg)` for all error conditions |
| MCP Prompts | Prompts not consumed by clients | Verify Claude Desktop prompt support before investing effort |
| MCP Completions | Completion API may not be stable in FastMCP | Pin FastMCP version; test completion behavior before release |
| Input validation | Over-validating params blocks valid SerpApi parameters | Validate structure only; allow unknown params to pass through |
| Progress reporting | `stateless_http=True` limits progress to single requests | Document limitation; use per-request progress only |
| `stateless_http=True` | Adding session-dependent features that silently fail | Document which MCP features don't work in stateless mode |

## Sources

- FastMCP documentation (Context7): Context object, stateless_http behavior, ToolError — HIGH confidence
- MCP specification (Context7): ServerCapabilities, isError flag, session requirements — HIGH confidence
- Source code (`src/server.py`): stateless_http=True, string errors, get_http_request usage — HIGH confidence
- Starlette BaseHTTPMiddleware: known issues with streaming and body consumption — MEDIUM confidence
- MCP client adoption: Claude Desktop supports Tools, Resources, Prompts; Completions/Tasks support varies — MEDIUM confidence (needs verification against current client versions)
- SerpApi Python package exception structure: needs source code verification — LOW confidence