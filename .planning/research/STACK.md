# Technology Stack

**Project:** SerpApi MCP Server — Hardening & Extension
**Researched:** 2026-04-15
**Existing stack baseline:** FastMCP 2.13+ on Starlette + uvicorn, Python 3.13, uv package manager

## Recommended Stack

### Core Framework (existing, verified)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| FastMCP | >=2.13.0.2 | MCP server framework | Already in use. Best Python MCP framework. Provides tools, resources, `Client` for testing, middleware. Active development by Prefect team. |
| Starlette | >=0.50.0 | ASGI framework | FastMCP's HTTP transport layer. Already in use. `TestClient` for sync middleware tests, `httpx.AsyncClient` + `ASGITransport` for async integration tests. |
| uvicorn | >=0.38.0 | ASGI server | Production-grade. Already in use. Supports `ws="none"` for HTTP-only mode. |
| serpapi (Python) | >=0.1.5 | SerpApi client library | Official Python client. Provides `serpapi.search()`, `SerpResults`, and exception types. **Uses `requests` internally, not `httpx`** — critical for test mocking strategy. |

### Testing Framework

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| **pytest** | >=7.0 (existing) | Test runner and framework | Already in dev deps. Industry standard for Python. Powerful fixture system, parametrize support. No reason to switch. |
| **pytest-asyncio** | **>=0.24.0** (upgrade from 0.21) | Async test execution | Required because all MCP tool/resource calls are `async`. Upgrade to >=0.24 for stable `asyncio_mode = "auto"`. Version 0.21 had issues with auto mode. Auto mode eliminates `@pytest.mark.asyncio` boilerplate. |
| **responses** | **>=0.25.0** (new) | Mock `requests` library HTTP calls | **Critical discovery:** `serpapi` Python client uses `requests` (not `httpx`) internally. `responses` intercepts `requests` calls at the transport level, enabling realistic mocking of `serpapi.search()` without hitting the real API. `pytest-httpx` would NOT work here because `serpapi` doesn't use `httpx`. |
| **httpx** | >=0.25.0 (existing) | Async HTTP test client | Already a production dep. Use `httpx.AsyncClient` with `httpx.ASGITransport` for integration tests of the Starlette app (middleware, auth, routing). No new dependency. |
| **FastMCP Client** | (bundled, >=2.13) | In-memory MCP protocol testing | `Client(mcp_server)` with `FastMCPTransport` for zero-network testing. Canonical way to test MCP tools and resources. No HTTP server needed. |
| **FastMCP `run_server_async`** | (bundled, >=2.13) | HTTP transport integration testing | `fastmcp.utilities.tests.run_server_async` provides an in-process HTTP server context manager. Use for full Streamable HTTP transport tests with `StreamableHttpTransport`. |

### Type Checking & Linting

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| **mypy** | >=1.5.0 (existing) | Static type checker | Already configured with `disallow_untyped_defs = true`. Project constraint is mypy compliance — all functions need annotations. No alternative needed. |
| **black** | >=23.0 (existing) | Code formatter | Already configured with `line-length = 100`. Zero-config, consistent. |
| **isort** | >=5.12.0 (existing) | Import sorter | Already configured with `profile=black`. Keeps imports consistent. |
| **ruff** | **>=0.8.0** (new, replaces flake8) | Linter | **Replaces `flake8`.** Ruff is 10-100x faster, has zero-config flake8 compatibility, auto-fix, and 100+ more rules. The project lists flake8 in dev deps but has NO `.flake8` config file — ruff is a strict upgrade. Add ruff, remove flake8. |

### Schema Validation

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| **jsonschema** | **>=4.23.0** (new) | Validate engine JSON schemas at startup | Lightweight, standards-compliant JSON Schema validator. Use to validate each `engines/*.json` file against a schema dict at startup. No heavier framework needed — engine schemas are simple flat dicts with predictable keys (`params`, `common_params`, each containing `{type, options, required, description, group}`). |

## Testing Strategy Detail

### Layer 1: Unit Tests (FastMCP Client, in-memory)

```python
from fastmcp import Client, FastMCP

async def test_search_tool():
    mcp = create_test_server()  # Fresh instance with mocked serpapi
    async with Client(mcp) as client:
        result = await client.call_tool("search", {"params": {"q": "test"}, "mode": "compact"})
        assert result.data is not None
```

**Why:** In-memory transport bypasses HTTP entirely. Fastest, most isolated tests. Validates MCP tool/resource contract without network overhead.

**Best for:** Tool logic, resource loading, parameter validation, error handling.

### Layer 2: Integration Tests (Starlette TestClient)

```python
from starlette.testclient import TestClient

def test_healthcheck():
    app = create_starlette_app()
    client = TestClient(app)
    response = client.get("/healthcheck")
    assert response.status_code == 200

def test_api_key_missing():
    client = TestClient(create_starlette_app())
    response = client.post("/mcp", json={})
    assert response.status_code == 401
```

**Why:** Starlette's `TestClient` wraps ASGI synchronously. Perfect for testing middleware (auth, CORS, metrics) and HTTP routing without starting uvicorn.

**Best for:** `ApiKeyMiddleware` auth flows, healthcheck endpoint, CORS headers, 401/403 response formatting.

### Layer 3: Mocking SerpApi Calls

```python
import responses
import serpapi

@responses.activate
def test_search_with_mock():
    responses.add(
        responses.GET,
        "https://serpapi.com/search",
        json={"organic_results": [{"title": "Test"}]},
        status=200,
    )
    result = serpapi.search({"q": "test", "api_key": "fake"})
    assert result.as_dict()["organic_results"][0]["title"] == "Test"
```

**Why:** `serpapi` uses `requests` internally, not `httpx`. Using `responses` to intercept `requests` calls is more realistic than `unittest.mock.patch` because it tests the full call chain through the HTTP client layer. Catches issues like incorrect URL construction or header handling that `patch` would miss.

### Layer 4: HTTP Transport Tests (FastMCP `run_server_async`)

```python
from fastmcp import Client
from fastmcp.client.transports import StreamableHttpTransport
from fastmcp.utilities.tests import run_server_async

async def test_http_transport():
    mcp = create_test_server()
    async with run_server_async(mcp) as url:
        async with Client(transport=StreamableHttpTransport(url)) as client:
            result = await client.call_tool("search", {"params": {"q": "test"}})
            assert result.data is not None
```

**Why:** Validates the full Streamable HTTP transport path. Essential for catching middleware + MCP protocol interaction bugs. Slower than in-memory tests, so use sparingly — only for smoke/integration tests.

## Recommended pyproject.toml Configuration

```toml
[tool.pytest.ini_options]
asyncio_mode = "auto"
testpaths = ["tests"]

[tool.ruff]
line-length = 100
target-version = "py313"

[tool.ruff.lint]
select = ["E", "F", "W", "I"]  # Flake8-equivalent + isort
ignore = ["E501"]  # black handles line length
```

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| Mocking (SerpApi) | `responses` | `unittest.mock.patch("serpapi.search")` | `patch` is fragile — couples tests to internal call structure. `responses` intercepts at HTTP level, catches URL/header bugs. Use `patch` only for trivial unit tests where zero HTTP overhead is needed. |
| Mocking (SerpApi) | `responses` | `pytest-httpx` | `serpapi` Python client uses `requests`, not `httpx`. `pytest-httpx` only intercepts `httpx.AsyncClient`/`httpx.Client` calls. **Will not work.** |
| Linting | `ruff` | `flake8` | `flake8` is slower, requires separate config, and the project has no `.flake8` config. `ruff` is a superset of flake8 rules with auto-fix. Zero reason to stay on flake8. |
| Schema validation | `jsonschema` | `pydantic` models | `pydantic` is overkill for flat JSON validation at startup. `jsonschema` is 50 lines of code, no model classes needed. `pydantic` is available as a transitive dep but adds ceremony for no benefit here. |
| Schema validation | `jsonschema` | Manual `dict` key checks | Manual checks silently drift from reality. `jsonschema` is declarative, self-documenting, and catches all edge cases (missing keys, wrong types, unexpected values). |
| Async testing | `pytest-asyncio` (auto mode) | `anyio` + `anyio.from_thread.run` | `anyio` is great for Trio-based code, but this project uses standard `asyncio`. `pytest-asyncio` is the standard, and `auto` mode removes boilerplate. |
| HTTP test client | `httpx.AsyncClient + ASGITransport` | `Starlette TestClient` only | Both are valid. `TestClient` is simpler for sync middleware tests. `httpx.AsyncClient` is needed for async integration tests. **Use both** — TestClient for sync, AsyncClient for async. |
| Metrics | CloudWatch EMF (current) | OpenTelemetry | CloudWatch EMF is appropriate for AWS deployment. Adding OTel would be over-engineering for a single-service deployment. Stick with EMF. |

## Installation

```bash
# Add new dev dependencies
uv add --dev responses>=0.25.0 ruff>=0.8.0

# Upgrade pytest-asyncio for stable auto mode
uv add --dev "pytest-asyncio>=0.24.0"

# Add jsonschema for startup validation
uv add jsonschema>=4.23.0

# Remove deprecated dev dependency
uv remove --dev flake8
```

## Dependency Summary

### Keep (existing, verified)

| Package | Version | Type | Notes |
|---------|---------|------|-------|
| fastmcp | >=2.13.0.2 | production | Core framework. `Client` + `FastMCPTransport` for testing. |
| starlette | >=0.50.0 | production | ASGI framework. `TestClient` for middleware tests. |
| uvicorn | >=0.38.0 | production | ASGI server. |
| httpx | >=0.25.0 | production | Already available for `AsyncClient` tests. |
| python-dotenv | >=1.0.0 | production | Env loading. |
| serpapi | >=0.1.5 | production | SerpApi client. **Uses `requests` internally!** |
| beautifulsoup4 | >=4.12.0 | production | Engine schema generation (`build-engines.py`). |
| markdownify | >=0.14.1 | production | Engine schema generation (`build-engines.py`). |
| pytest | >=7.0 | dev | Test runner. |
| black | >=23.0 | dev | Formatter. |
| isort | >=5.12.0 | dev | Import sorter. |
| mypy | >=1.5.0 | dev | Type checker. |

### Upgrade

| Package | From | To | Reason |
|---------|------|----|--------|
| pytest-asyncio | >=0.21.0 | **>=0.24.0** | Stable `asyncio_mode = "auto"` support. Version 0.21 had issues. |

### Add (new)

| Package | Version | Type | Purpose |
|---------|---------|------|---------|
| responses | >=0.25.0 | dev | Mock `requests` calls for serpapi testing |
| jsonschema | >=4.23.0 | production | Validate engine JSON schemas at startup |
| ruff | >=0.8.0 | dev | Linter (replaces flake8) |

### Remove

| Package | Reason |
|---------|--------|
| flake8 | Replaced by `ruff`. No config file exists, ruff is a strict upgrade (faster, more rules, auto-fix). |

## Sources

- **FastMCP testing docs** (Context7, `/prefecthq/fastmcp`): in-memory Client testing, `run_server_async`, `FastMCPTransport` — HIGH confidence, verified against installed package
- **FastMCP Client API**: verified via `inspect` of installed v2.13.0.2 — `call_tool`, `list_tools`, `list_resources`, `read_resource`, `ping` all confirmed — HIGH confidence
- **FastMCP `http_app` signature**: verified via inspection — includes `json_response`, `stateless_http`, `transport`, `middleware` params — HIGH confidence
- **Starlette TestClient**: Context7 `/kludex/starlette` — `TestClient` for sync ASGI testing, `httpx.AsyncClient` with `ASGITransport` for async — HIGH confidence
- **pytest-asyncio auto mode**: Context7 `/pytest-dev/pytest-asyncio` — `asyncio_mode = "auto"` in pyproject.toml — HIGH confidence
- **`responses` library**: Context7 `/getsentry/responses` — `@responses.activate` decorator and `RequestsMock` context manager for mocking `requests` — HIGH confidence
- **`serpapi` uses `requests` internally**: verified by inspecting `serpapi.http.HTTPClient` source — uses `requests.Session` — HIGH confidence, source code verified
- **jsonschema**: well-established Python JSON Schema validation library, minimal dependency — HIGH confidence, ecosystem standard
- **ruff as flake8 replacement**: https://docs.astral.sh/ruff/ — 10-100x faster, auto-fix, superset of flake8 rules — HIGH confidence, widely adopted industry standard