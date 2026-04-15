# Pitfalls Research

**Domain:** Python MCP server hardening (testing, typing, validation, error handling)
**Researched:** 2026-04-15
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: `get_http_request()` Makes In-Process MCP Testing Impossible

**What goes wrong:**
The `search` tool calls `get_http_request()` from `fastmcp.server.dependencies` to extract the API key from the Starlette request context. This context only exists during real HTTP requests. FastMCP's recommended testing pattern — `Client(transport=FastMCPTransport(mcp))` — runs tools in-process without any HTTP request. Every tool call fails with `"Error: Unable to access API key from request context"`.

**Why it happens:**
FastMCP's testing docs show `Client(transport=mcp)` as the primary way to test tools, but this only works for tools that don't depend on HTTP request context. Developers naturally try the documented pattern first, get confusing runtime errors, and waste time debugging the framework rather than their code.

**How to avoid:**
- Test the `search` tool via HTTP-level integration tests using Starlette's `TestClient` or `httpx.AsyncClient` with `ASGITransport`, not via `FastMCPTransport`
- Use `FastMCPTransport`-based `Client` only for testing resource listing/reading (engine schemas), which don't depend on request context
- Refactor `search` to accept `api_key` as an explicit parameter (with optional injection from request context as fallback) so it can be tested in isolation with FastMCPTransport
- For the current code, the correct integration test pattern is:

```python
from starlette.testclient import TestClient

def test_search_with_auth():
    app = mcp.http_app(middleware=[...], stateless_http=True, json_response=True)
    with TestClient(app) as client:
        response = client.post("/test-key/mcp", json={...})
        # Assert on MCP protocol response
```

**Warning signs:**
- Test for `search` returns "Error: Unable to access API key from request context" when using `Client(transport=FastMCPTransport(mcp))`
- Writing `# type: ignore` or mock patches for `get_http_request` in tool-level unit tests
- Tests that need excessive mocking of internal FastMCP machinery

**Phase to address:**
Phase 1 (Test Suite) — Must establish the correct testing patterns from the start. If the first tests use FastMCPTransport for the search tool, all subsequent tests will be built on a broken foundation.

---

### Pitfall 2: Module-Level Side Effects Prevent Test Isolation

**What goes wrong:**
`server.py` executes `for _engine_path in _get_engine_files():` at module import time (lines 71-77), registering 100+ engine resources. Importing `server.py` in a test triggers disk I/O to read all engine JSON files and registers every resource on the module-level `mcp` object. This makes tests:
1. Fail if `engines/` directory doesn't exist relative to the import path
2. Slow (reading 100+ JSON files on every import)
3. State-polluted (resources persist across tests on the module-level `mcp` instance)
4. Impossible to test engine loading in isolation

**Why it happens:**
FastMCP's `@mcp.resource()` decorator and `mcp.add_resource()` are designed for eager registration. The module-level loop is the straightforward way to register dynamic resources. But module-level side effects are a well-known Python testing anti-pattern — the module's import becomes a non-trivial operation.

**How to avoid:**
- Extract engine registration into an explicit `register_engines(mcp: FastMCP, engines_dir: Path)` function called from `main()`
- In tests, import `server` module once (scoped fixture) and don't re-register engines per test
- For engine-related tests, create a fresh `FastMCP` instance and register only the engines needed
- Keep `_engine_resource_factory` and `_get_engine_files` as-is (they're pure functions), but move the registration loop out of module scope

```python
# Refactored: move from module-level to function
def register_engines(mcp: FastMCP, engines_dir: Path | None = None) -> None:
    if engines_dir is None:
        engines_dir = ENGINES_DIR
    for engine_path in _get_engine_files(engines_dir):
        engine_name = engine_path.stem
        if not re.fullmatch(r"[a-z0-9_]+", engine_name):
            logger.warning("Skipping invalid engine filename: %s", engine_name)
            continue
        mcp.add_resource(_engine_resource_factory(engine_name, engine_path))

# In main():
register_engines(mcp)
```

**Warning signs:**
- `ImportError` or `FileNotFoundError` in tests that merely import `server.py`
- Engine count differs between test and production environments
- Tests that need `engines/` directory to exist on disk
- Slow test suite startup (>1 second before first test runs)

**Phase to address:**
Phase 1 (Test Suite) — Must refactor engine registration before writing tests that import the module. This is a prerequisite for testable code.

---

### Pitfall 3: BaseHTTPMiddleware Breaks Streaming and Complicates Testing

**What goes wrong:**
`ApiKeyMiddleware` and `RequestMetricsMiddleware` extend `BaseHTTPMiddleware`. This Starlette class has well-documented issues:
1. **Request body consumption:** `BaseHTTPMiddleware` reads the entire request body into memory, which breaks streaming request/response patterns and can cause hangs with certain ASGI servers
2. **Path mutation in middleware:** `ApiKeyMiddleware` mutates `request.scope["path"]` and `request.scope["raw_path"]` — this modifies shared state and can cause subtle bugs where later middleware or route matching sees the modified path
3. **Test order sensitivity:** Because middleware mutates request state, test order can matter if tests don't properly clean up

**Why it happens:**
`BaseHTTPMiddleware` is the most readable way to write Starlette middleware. The path-rewriting pattern (`/{KEY}/mcp` → `/mcp`) seems natural for URL-based auth. The documentation doesn't prominently warn about the streaming issues.

**How to avoid:**
- For the auth middleware, consider a pure ASGI middleware instead of `BaseHTTPMiddleware`. It avoids the body-consumption issue entirely:

```python
class ApiKeyMiddleware:
    def __init__(self, app: ASGIApp) -> None:
        self.app = app

    async def __call__(self, scope: Scope, receive: Receive, send: Send) -> None:
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return
        # ... auth logic via scope inspection, no Request object needed
```

- Alternatively, since the MCP server uses `json_response=True` (no streaming), `BaseHTTPMiddleware` is acceptable for now — but document this dependency explicitly
- For path rewriting: use `scope` manipulation in pure ASGI rather than creating a new `Request` object
- In tests, always use `TestClient` as a context manager to ensure proper ASGI lifecycle

**Warning signs:**
- Middleware tests pass in isolation but fail when run with other tests
- Intermittent test failures (body already consumed errors)
- `422 Unprocessable Entity` responses from MCP endpoint (body was consumed by middleware before reaching handler)
- CORS errors that only appear in production (middleware ordering matters)

**Phase to address:**
Phase 2 (Error Handling / Type Annotations) — The middleware itself works; the risk is in testing. If tests are flaky, this is why. Pure-ASGI refactor is optional but recommended if streaming is ever needed.

---

### Pitfall 4: Mutable Default Argument in Tool Signature

**What goes wrong:**
The `search` tool is defined as `async def search(params: dict[str, Any] = {}, mode: str = "complete")`. The mutable default `{}` is a classic Python bug — if the dict were ever mutated in place, subsequent calls would see the modifications. More immediately, mypy's `disallow_untyped_defs` isn't the issue here (the types are annotated), but `mypy --strict` with `disallow_any_generics` would flag `dict[str, Any]` because `Any` masks type errors.

**Why it happens:**
Convenient default that makes `params` optional. The intent is "if you don't specify params, use defaults." Python's default argument evaluation makes this share a single dict across calls.

**How to avoid:**
Replace with `None` default:

```python
async def search(params: dict[str, Any] | None = None, mode: str = "complete") -> str:
    if params is None:
        params = {}
```

This is the canonical Python pattern. It's also more explicit about the intent.

**Warning signs:**
- Linter warnings (ruff, flake8-bugbear) about mutable default arguments
- State leaking between test calls of the `search` tool
- mypy strict mode errors on `dict[str, Any]`

**Phase to address:**
Phase 2 (Type Annotations) — Fix as part of annotation pass. Low risk but should be caught early.

---

### Pitfall 5: MCP Error Responses as String Prefixes Instead of Protocol Errors

**What goes wrong:**
The `search` tool returns error information as plain strings prefixed with `"Error:"` (e.g., `"Error: Rate limit exceeded. Please try again later."`). This is not how MCP handles errors — MCP has a dedicated `isError` flag on tool results. When `isError=True`, clients display the error properly in their UI; when it's a string, clients display it as if the search succeeded but returned weird text.

**Why it happens:**
The simplest approach — `return f"Error: {msg}"` — appears to work. The MCP protocol documentation on error handling is not prominent in FastMCP's tutorials. Developers see the string come back and assume it works.

**How to avoid:**
FastMCP supports raising errors that set `isError=True`:

```python
from fastmcp.exceptions import ToolError

# Instead of: return "Error: Rate limit exceeded"
# Use:
raise ToolError("Rate limit exceeded. Please try again later.")
```

Or return structured error content:

```python
# If you want to return error info without raising:
from fastmcp.server.context import Context

@mcp.tool()
async def search(params: dict[str, Any] | None = None, mode: str = "complete") -> str:
    # ... validation ...
    if mode not in ("complete", "compact"):
        raise ToolError("Invalid mode. Must be 'complete' or 'compact'")
```

This lets MCP clients properly distinguish between successful and failed tool calls.

**Warning signs:**
- AI agents interpreting `"Error: ..."` as search result text instead of error messages
- Client-side error handling that parses string prefixes (`if result.startswith("Error:")`)
- MCP inspector showing 200 OK for clearly failed requests

**Phase to address:**
Phase 2 (Error Handling) — This is a core error handling improvement. The `search` tool should use `ToolError` for all error conditions, not string returns.

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Mutable default `params={}` | Shorter function calls | Shared state bugs, linter warnings, mypy issues | Never — use `None` default |
| Module-level engine registration | Simple, runs at startup | Un-testable, slow test startup, can't isolate | Only during initial prototyping; must refactor before tests |
| String-prefixed errors (`"Error: ..."`) | Quick to implement | Clients can't distinguish errors from results, breaks MCP protocol | Never — use `ToolError` or `isError=True` |
| `Any` in type annotations (`dict[str, Any]`) | Passes mypy immediately | Hides real type errors in `params` dict | Acceptable for external API responses (like SerpApi results); unacceptable for internal config |
| `extract_error_response(exception)` with no type annotation | Works for all exceptions | mypy error, fragile exception traversing, breaks if SerpApi changes exception structure | Only until proper exception handling is added |
| `load_dotenv()` at module level | Works in dev | Can override test env vars, causes side effects on import | Move into `main()` or guard behind `if __name__` |
| Hardcoded `ENGINES_DIR = Path(__file__).resolve().parents[1] / "engines"` | Simple, works | Can't override for testing, tied to source layout | Acceptable if function parameter added as override path |

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| **SerpApi client** | Catching `HTTPError` and parsing `"429"` from `str(e)` — string matching on exception messages | Use `e.status_code` if available, or wrap SerpApi calls in a typed error class that normalizes HTTP status codes |
| **SerpApi client** | Not handling network timeouts or connection errors separately from API errors | Add explicit timeout handling (`httpx.TimeoutError`) and connection error handling (`httpx.ConnectError`) — the `serpapi` library uses httpx internally |
| **FastMCP testing** | Using `Client(transport=FastMCPTransport(mcp))` for tools that use `get_http_request()` | Use Starlette `TestClient` or `httpx.AsyncClient` with `ASGITransport` for HTTP-dependent tools |
| **FastMCP version** | Assuming the FastMCP API is stable — `>=2.13.0.2` allows breaking changes | Pin the FastMCP version in `pyproject.toml` (use `~=` or pin the minor), test against locked version |
| **Starlette TestClient** | Using `TestClient(app)` without context manager | Use `with TestClient(app) as client:` to trigger lifespan events and proper cleanup |
| **pytest-asyncio** | Using default `"strict"` mode and forgetting `@pytest.mark.asyncio` on every test | Set `asyncio_mode = "auto"` in `pyproject.toml` `[tool.pytest.ini_options]` per FastMCP convention |
| **SerpApi exceptions** | Assuming `serpapi.exceptions.HTTPError` has a `.response` attribute with `.json()` — response structure varies | Always handle `ValueError`/`AttributeError` when parsing exception responses (code already does this, but tests must verify) |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Loading 100+ engine JSON files on every import | Slow test startup (>2s), slow cold starts | Lazy loading or caching; move engine loading out of module scope | Noticeable with >200 test runs or serverless cold starts |
| `json.dumps(data, indent=2)` in `search` tool | Larger response payloads than necessary; `indent=2` adds ~30% size overhead | Use `indent=None` for production, `indent=2` only in debug mode; or let the client format | High request volume where bandwidth matters |
| `BaseHTTPMiddleware` reading full request body | Increased memory usage per request, potential hangs on large inputs | Move to pure ASGI middleware that doesn't consume the body | Large JSON payloads (>1MB) or streaming requests |
| SerpApi blocking calls in async tool | `serpapi.search()` is synchronous — blocks the event loop | Run in `asyncio.to_thread()` to avoid blocking the ASGI event loop | Concurrent MCP sessions (>10 simultaneous tool calls) |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| API key logged or leaked in error strings | API key exposure in logs and error responses | Never include the API key in log messages or error strings. Currently safe, but error parsing must never accidentally forward the key |
| Healthcheck bypasses auth but returns service info | Information disclosure (service name, version) | Healthcheck intentionally skips auth — acceptable for a health endpoint. Keep response minimal (don't add version/server headers) |
| `CORS allow_origins=["*"]` with `allow_credentials=True` | This combination is rejected by browsers per spec (CORS spec forbids wildcard origin with credentials). Browsers will block requests | Use `allow_credentials=False` with `allow_origins=["*"]`, or specify explicit origins with `allow_credentials=True` |
| No request size limit | DOS via oversized JSON payloads | Add middleware or Starlette config to limit request body size (e.g., 1MB) |
| Path-based API key in URL | API key appears in access logs, browser history, referer headers | Already known and accepted (Smithery convention). Document the trade-off clearly |

## UX Pitfalls

Common user experience mistakes in this domain (MCP server for AI agents).

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Returning `"Error: ..."` as tool result text | AI agent interprets error as successful search result and tries to use it | Use `ToolError` / `isError=True` so agents see it as a failure |
| `params` dict without schema validation | AI agent sends malformed params, gets cryptic SerpApi errors | Validate required params exist before calling SerpApi; return clear `ToolError` for missing `q` parameter |
| `mode` as free-form string | Agent sends `mode="full"` by mistake, gets unexpected compact output | Validate `mode` against `["complete", "compact"]` and raise `ToolError` for invalid values (partially done, but returns string error) |
| Error messages include raw SerpApi error JSON | Agent can't parse the nested JSON string within text content | Catch SerpApi errors and return human-readable messages with the structured data as a separate field |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Test suite:** Often missing HTTP-level integration tests — verify that `/mcp` endpoint actually responds to MCP protocol requests with proper auth
- [ ] **Test suite:** Often missing tests for `ApiKeyMiddleware` path-rewriting behavior — verify `/{KEY}/mcp` strips the key and routes correctly
- [ ] **Test suite:** Often missing tests for engine resource listing — verify resources appear via MCP protocol and JSON schemas are valid
- [ ] **Type annotations:** Often missing annotations on `async def dispatch(self, request: Request, call_next)` in middleware — mypy flags inherited untyped methods
- [ ] **Error handling:** Often missing `ToolError` usage — verify errors use MCP's `isError` flag, not string prefixes
- [ ] **Schema validation:** Often validates engine JSON too strictly — verify validation is structural (keys exist, types match) not value-level (specific field values)
- [ ] **CORS configuration:** Often has `allow_origins=["*"]` with `allow_credentials=True` — verify this is either split or credentials are dropped
- [ ] **Search tool:** Often doesn't run `serpapi.search()` off the event loop — verify `asyncio.to_thread()` is used for the blocking call

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| In-process tests fail due to `get_http_request()` | LOW | Add `Starlette TestClient` integration tests; refactor `search` to accept `api_key` parameter with request-context fallback |
| Module-level side effects prevent testing | MEDIUM | Extract `register_engines()` function; call from `main()` instead of module level; adjust tests to call it explicitly |
| Mutable default causes state leak | LOW | Change `params={}` to `params=None` with `if params is None: params = {}` — one-line fix |
| String errors instead of `ToolError` | LOW | Replace `return "Error: ..."` with `raise ToolError(...)` — mechanical refactoring |
| mypy wall of errors on untyped codebase | MEDIUM | Annotate functions one at a time, run mypy per-file, use `# type: ignore[no-untyped-def]` as temporary suppression during migration |
| Engine schema validation too strict | LOW | Relax validation to structural checks only; allow unknown keys with `additionalProperties: true` if using JSON Schema |
| `BaseHTTPMiddleware` causing hangs | MEDIUM | Rewrite as pure ASGI middleware — more code but straightforward pattern |
| Blocking `serpapi.search()` in async handler | LOW | Wrap in `asyncio.to_thread(serpapi.search, search_params)` — single line change |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| `get_http_request()` blocks in-process testing | Phase 1 (Test Suite) | Write first test using both `TestClient` and `FastMCPTransport`; verify both work |
| Module-level engine registration | Phase 1 (Test Suite) | Refactor to `register_engines()` function; verify tests can import server without side effects |
| Mutable default argument | Phase 2 (Type Annotations) | Fix `params={}` → `params=None`; verify mypy passes |
| String error responses | Phase 2 (Error Handling) | Replace with `ToolError`; verify MCP Inspector shows `isError=True` |
| `BaseHTTPMiddleware` pitfalls | Phase 2 (Error Handling) | Document that `json_response=True` mitigates streaming risk; add ASGI middleware pattern as reference |
| Engine schema validation too strict | Phase 3 (Schema Validation) | Validate structure only; verify that new SerpApi engines with unknown fields still pass |
| `load_dotenv()` at module level | Phase 2 (Type Annotations) | Move into `main()`; verify tests can set env vars without interference |
| Blocking sync `serpapi.search()` | Phase 2 (Error Handling) | Wrap in `asyncio.to_thread()`; verify concurrent requests don't block event loop |
| CORS `allow_credentials=True` with wildcard origin | Phase 2 (Error Handling) | Fix configuration; verify browser requests still work |
| mypy strict mode wall of errors | Phase 2 (Type Annotations) | Annotate incrementally per function; verify mypy passes file-by-file |
| `extract_error_response` fragile exception traversal | Phase 2 (Error Handling) | Add typed exception handling for `serpapi.exceptions.HTTPError` with status code checking |
| `pytest-asyncio` mode configuration | Phase 1 (Test Suite) | Add `asyncio_mode = "auto"` to `pyproject.toml`; verify async tests run without decorators |

## Sources

- FastMCP testing documentation (testing.mdx, development/tests.mdx) — Context7 verified, HIGH confidence
- Starlette TestClient documentation — Context7 verified, HIGH confidence  
- mypy `disallow_untyped_defs` documentation — Context7 verified, HIGH confidence
- FastMCP `get_http_request()` API — Context7 verified, HIGH confidence
- FastMCP `ToolError` exception class — Context7 verified, HIGH confidence
- FastMCP `http_app(stateless_http=True, json_response=True)` configuration — Codebase verified, HIGH confidence
- Starlette `BaseHTTPMiddleware` known issues — Community documentation, MEDIUM confidence
- SerpApi `google-search-results` Python package exception structure — Training data, LOW confidence (verify with package docs)
- MCP protocol specification for `isError` flag — Context7 verified, HIGH confidence

---
*Pitfalls research for: SerpApi MCP Server hardening (testing, typing, validation, error handling)*
*Researched: 2026-04-15*