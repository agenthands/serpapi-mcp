# Feature Research

**Domain:** Go MCP server — SerpApi search gateway rewrite
**Researched:** 2026-04-15
**Confidence:** HIGH

## Executive Summary

The Go rewrite must faithfully port existing capabilities while fixing critical deficiencies in the Python implementation: error handling that violates the MCP spec (string prefixes instead of `IsError` flag), zero input validation, no request correlation, no graceful shutdown, and fragile error extraction. Beyond the port, production Go MCP servers add structured logging (slog), composable middleware chains, proper MCP-compliant error responses, startup fail-fast validation, and multi-platform static binary distribution via GoReleaser. The Go MCP SDK ecosystem is mature — mark3labs/mcp-go (8.6k stars) provides streamable HTTP transport, tool/resource middleware, hooks, session management, and proper error typing out of the box.

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete or untrustworthy.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **MCP server with Streamable HTTP transport** | Core capability. The server IS an MCP server; no transport = no server. | Low | mark3labs/mcp-go provides `server.NewStreamableHTTPServer()`. ~20 lines to set up. Official go-sdk provides `mcp.NewStreamableHTTPHandler()`. Either SDK makes this trivial. |
| **API key auth middleware** | Existing clients depend on `/{KEY}/mcp` path and `Authorization: Bearer` header auth. Removing breaks backward compatibility. | Medium | Python mutates `request.scope["path"]` to strip key — brittle. In Go, use HTTP middleware before MCP handler or trpc-mcp-go's `WithHTTPContextFunc()` to extract auth into context. Path rewriting is cleaner with Go's `http.ServeMux` patterns. |
| **Search tool** | This IS the product. Single `search` tool routing to all SerpApi engines is the core value proposition. | Low | Direct HTTP call to SerpApi REST API (`GET https://serpapi.com/search`). No Go SDK for SerpApi — use `net/http` directly with `encoding/json`. Simpler than Python's `serpapi` client. |
| **Engine parameter schemas as MCP resources** | Agents discover available engines and parameters via `serpapi://engines` and `serpapi://engines/<engine>`. Existing clients depend on these URIs. | Low | Load `engines/*.json` at startup, register as `mcp.NewResource()` with mark3labs/mcp-go. Engine index resource + per-engine resources. ~100 engines. |
| **Complete and compact response modes** | Existing clients use `mode=compact` to reduce response size. Removing breaks these clients. | Low | Compact mode removes 5 metadata fields (`search_metadata`, `search_parameters`, `search_information`, `pagination`, `serpapi_pagination`). Simple JSON field filtering in Go. |
| **Proper MCP error responses** | MCP spec requires `IsError: true` flag on `CallToolResult` for errors. Python returns `"Error: ..."` strings — clients can't distinguish errors from results programmatically. **This is the #1 bug to fix.** | Low | mark3labs/mcp-go: `mcp.NewToolResultError("message")` sets `IsError=true`. Official go-sdk: `result.SetError(err)`. Both handle this natively. Python's string-prefix approach violates the spec. |
| **Input validation** | Python accepts any engine name, any mode value, any parameter — forwards garbage to SerpApi, produces confusing errors. Production servers validate before proxying. | Medium | Validate engine name against loaded schemas (reject unknown engines). Validate `mode` (only `"complete"` or `"compact"`). Validate required params per engine schema. Return structured MCP errors with actionable messages. |
| **Structured logging with request correlation IDs** | Production debugging requires correlated logs. Python has basic `logging` + CloudWatch EMF, no correlation IDs. Impossible to trace a request from auth → tool call → SerpApi response → error. | Medium | Use Go's `log/slog` (stdlib, Go 1.21+). Structured JSON logging with key-value pairs. Add correlation ID middleware that injects `slog.Attr` into context. Every log line includes `request_id`, `engine`, `mode`. Essential for operating in production. |
| **Startup validation (fail fast)** | Python warns on missing `engines/` but starts anyway, serving broken resources. Production servers must validate dependencies before accepting traffic. | Low | At startup: verify `engines/` directory exists, all JSON files parse, required fields present. Exit with clear error if validation fails. No more silent degradation. |
| **Graceful shutdown** | Python's `uvicorn.run()` doesn't handle SIGTERM/SIGINT for connection draining. In Go, `signal.NotifyContext()` + `http.Server.Shutdown()` is standard practice. Production deployments require clean shutdown. | Medium | Go pattern: `ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` → `server.Shutdown(ctx)`. Drains in-flight requests. mark3labs/mcp-go examples show this pattern. |
| **Healthcheck endpoint** | Existing clients and load balancers depend on `GET /healthcheck`. Required for deployment health probes. | Low | Simple HTTP handler returning `{"status":"healthy","timestamp":"..."}`. Not an MCP endpoint — separate from the MCP handler. Mount alongside MCP handler on Go's `http.ServeMux`. |
| **CORS support** | Browser-based MCP clients require CORS headers. Python enables permissive CORS. Removing breaks browser clients. | Low | Use standard Go CORS middleware (e.g., `rs/cors` package or manual header injection). Same permissive policy as Python: all origins, all methods, all headers. |
| **Multi-platform static binary builds** | Project explicitly ships static binaries instead of containers. Go's primary advantage over Python. Without cross-platform builds, the Go rewrite loses its deployment story. | Medium | GoReleaser config: `CGO_ENABLED=0`, build for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`. GitHub Release integration. ~40 lines of `.goreleaser.yml`. |
| **Test suite** | Python has zero tests. Go has no excuse to ship without tests — testing is first-class in Go. No tests = no confidence in rewrite correctness. | Medium | Go `testing` package + `testify` for assertions. Test auth middleware, search tool (valid/invalid engine, valid/invalid mode), resource loading, compact mode, error handling. Use `httptest.NewServer` for HTTP-level integration tests. |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Typed tool input schema** | mark3labs/mcp-go supports `mcp.WithString()`, `mcp.WithNumber()`, `mcp.Required()`, `mcp.Enum()` on tool definitions. Generates JSON Schema automatically. Python defines `params: dict[str, Any]` — no structure. Typed schema lets agents discover valid parameters without reading engine resources. | Low | Define `search` tool with explicit parameter schemas: `engine` (string, enum of known engines), `mode` (string, enum `"complete"|"compact"`), `q` (string, description). The `params` dict still needed for engine-specific params, but common params can be typed. |
| **MCP Completions** | Auto-completion for engine names and parameter values. Agents discover valid options interactively instead of reading all engine schemas. mark3labs/mcp-go supports `server.WithCompletions()`. | Medium | Implement completion provider for engine names (filter from loaded engines list). Wire engine schema data into `CompleteResourceArgument` handler. Directly improves agent UX. |
| **Tool handler middleware chain** | Composable middleware for auth, logging, metrics, recovery. mark3labs/mcp-go: `server.WithToolHandlerMiddleware()`. Python has no equivalent — all cross-cutting concerns mixed into handler code. | Low | Add recovery middleware (`server.WithRecovery()`) to prevent panics from crashing the server. Add logging middleware that logs tool name, duration, engine, error status. Cleanly separates concerns. |
| **Request hooks / lifecycle events** | mark3labs/mcp-go `server.WithHooks()` provides lifecycle callbacks: `OnRequest`, `OnToolCall`, `OnError`, etc. Enables observability without polluting handler code. Python has no equivalent. | Low | Register hooks for: request start/end, tool call timing, error tracking. Feed into slog structured logging. Zero business logic in hooks — pure observability. |
| **Version info in server capabilities** | MCP spec includes `ServerInfo.Name` and `ServerInfo.Version`. Embedding build version via `-ldflags` at compile time gives deterministic versioning per binary. | Low | Go pattern: `var version = "dev"` set via `-ldflags "-X main.version=..."`. GoReleaser does this automatically. Report in MCP `initialize` response and `/healthcheck`. |
| **Engine list caching in memory** | Python reads engine files on every resource read (via closure, but the closure reads the file each time). Go can parse once at startup and serve from memory. Faster, less I/O. | Low | Load all `engines/*.json` into `map[string]json.RawMessage` at startup. Serve from map in resource handlers. Already a natural Go pattern. |
| **SerpApi HTTP client with timeouts** | Python's `serpapi` client uses `requests` with default timeouts (no limit). In Go, configure `http.Client{Timeout: 30s}` to prevent hanging connections from blocking goroutines. | Low | Create a shared `http.Client` with reasonable timeout (30s) and connection pool settings. Pass API key as query parameter. No Go SerpApi SDK needed — just HTTP + JSON. |
| **Configurable engine data path** | Python hardcodes `ENGINES_DIR` relative to `server.py`. Go binary has no reliable "relative to source" path. Need CLI flag / env var for engine data directory. | Low | `--engines-dir` flag / `ENGINES_DIR` env var. Default to `./engines` (relative to working directory). Validate at startup. Required for binary distribution where binary and data aren't co-located. |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| **OAuth2 authentication flow** | "Modern APIs use OAuth." | Adding PKCE, token exchange, refresh, scope management is massive complexity (1000+ lines). SerpApi uses API keys. MCP clients expect simple key auth. OAuth would be over-engineering for a stateless proxy. | Keep API key auth (path + header). Matches SerpApi's model. Simple, well-understood, zero config. |
| **Result caching** | "Reduce API calls and latency." | SerpApi handles server-side caching. Client-side caching serves stale results, adds invalidation complexity, TTL management, memory pressure for large result sets. A stateless proxy should not cache. | Let SerpApi cache. Clients cache if they want. Stateless = simpler. |
| **Server-side rate limiting** | "Protect against abuse." | SerpApi enforces per-key rate limits. Duplicating adds latency, config surface (`--rate-limit`, `--burst`, etc.), and incorrect limits if SerpApi's limits change. The proxy isn't the authority on rate limits. | Rely on SerpApi's rate limiting. Return SerpApi's 429 errors to clients with proper `IsError` responses. |
| **WebSocket/SSE transport** | "Support all MCP transports." | Project explicitly scopes to Streamable HTTP only. SSE is legacy. WebSocket adds complexity for bidirectional framing. Both require different handler patterns. Supporting multiple transports doubles the testing surface. | Streamable HTTP only. It's the current MCP standard. If needed later, add as separate phase. |
| **CloudWatch EMF metrics** | "Parity with Python server's metrics." | EMF is AWS-specific. Go binaries are deployment-agnostic (that's the point). Embedding CloudWatch logic in the binary contradicts the static binary distribution model. Would require AWS SDK dependency. | Use structured slog logging. Ship logs to whatever backend (CloudWatch, Datadog, Loki) via log aggregation, not code. Metrics can be added later if a specific platform is chosen. |
| **Container/Docker distribution** | "Docker is standard deployment." | Project explicitly ships static binaries. Docker adds image build, registry, tag management, platform-specific images. The Go binary IS the deployment artifact. Docker negates Go's single-binary advantage. | Ship static binaries. Let users containerize if they want — the binary is self-contained. |
| **MCP Resource subscriptions** | "Notify clients when engine schemas change." | Engine schemas change only across deployments, not during runtime. Adding subscribe/`listChanged` for resources adds session state, notification channels, and event routing — complexity with zero practical benefit for this server. | Reload/restart server when schemas change (deployment event, not runtime event). |
| **MCP Sampling** | "Let the server request LLM completions for query refinement." | Creates latency loops (client → server → LLM → server → SerpApi → client). The server is a stateless search proxy — it should not orchestrate LLM calls. Query composition is the client's job. | Skip sampling. Server returns search results; client/agent decides what to do with them. |
| **MCP Prompts** | "Pre-built prompt templates for common search patterns." | The Python research identified this as a differentiator, but for the Go rewrite it adds scope. Prompts require authored content, maintenance, testing. Not core to a search proxy. Agent frameworks have their own prompt management. | Defer to v1.x. Focus on core server correctness first. Prompts can be added without breaking changes. |
| **Database / persistent storage** | "Store analytics, API key mappings, session state." | Stateless proxy design. Adding a database creates operational overhead (migrations, backups, connection pooling), deployment complexity, and single point of failure. Search results are ephemeral. | Stateless. Log everything via slog. No server-side storage. |
| **Custom MCP tasks** | "Async long-running search operations." | MCP `tasks` are experimental. SerpApi calls complete in <5s. Adding task management (polling, cancellation, storage) is complexity for a problem that doesn't exist. | Keep synchronous `search` tool. If async is needed later, add as separate phase. |

## Feature Dependencies

```
Core MCP server (transport) ──────────────────────────────────┐
    │                                                          │
    ├──→ Auth middleware ──→ Search tool ──→ Compact mode      │
    │         │                    │                             │
    │         │                    ├──→ Proper error responses   │
    │         │                    ├──→ Input validation        │
    │         │                    │         │                   │
    │         │                    │         └──requires──→ Engine schemas │
    │         │                    │                             │
    │         │                    └──→ SerpApi HTTP client      │
    │         │                                                  │
    ├──→ Engine schemas (resources) ──→ Startup validation      │
    │         │                                                  │
    │         └──→ Engine list caching                          │
    │                                                            │
    ├──→ Healthcheck endpoint (independent)                     │
    ├──→ CORS middleware (independent)                           │
    ├──→ Graceful shutdown (independent)                        │
    │                                                            │
    ├──→ Structured logging (slog) ──→ Request correlation IDs  │
    │         │                                                  │
    │         └──→ Tool handler middleware                      │
    │                   │                                        │
    │                   └──→ Request hooks                      │
    │                                                            │
    ├──→ Test suite (independent, but validates everything)     │
    │                                                            │
    ├──→ GoReleaser config (independent)                        │
    │                                                            │
    └──→ Version embedding (independent)                        │

Critical path: Core server → Auth middleware → Search tool → Error handling
              Engine schemas → Input validation
              Structured logging → Correlation IDs
```

### Dependency Notes

- **Auth middleware requires core server setup**: Auth is HTTP middleware that wraps the MCP handler; must be configured during server initialization.
- **Search tool requires both auth middleware AND engine schemas**: Needs API key from auth context to make SerpApi calls, and engine schema data for validation.
- **Input validation requires engine schemas**: Can't validate engine names or required params without knowing what engines exist and what params they accept.
- **Structured logging enables request correlation IDs**: Correlation IDs are injected via logging middleware — without slog structured logging, correlation is just string concatenation.
- **Tool handler middleware enhances all tool calls**: Logging, recovery, and timing middleware wrap tool handlers — not strictly dependent on structured logging but enhanced by it.
- **Test suite validates everything**: Not on the critical path in terms of feature dependencies, but ships with each feature. No feature should merge without tests.
- **GoReleaser is fully independent**: Build/release pipeline is orthogonal to server features. Can set up at any point.

## MVP Definition

### Launch With (v1.0)

Minimum viable product — faithful port with critical Python bugs fixed.

- [x] Core MCP server with Streamable HTTP — product doesn't exist without it
- [x] API key auth middleware — existing clients break without it
- [x] Search tool — core value proposition
- [x] Engine parameter schemas as resources — agents can't use search without discovery
- [x] Complete and compact response modes — existing clients depend on compact mode
- [x] Proper MCP error responses (`IsError` flag) — **#1 bug fix from Python**, spec violation
- [x] Input validation (engine names, mode, required params) — **#2 bug fix**, prevents garbage-in-garbage-out
- [x] Healthcheck endpoint — deployment health probes depend on it
- [x] CORS support — browser clients depend on it
- [x] Startup validation — fail fast on missing/corrupt engine schemas
- [x] Structured logging with correlation IDs — production requirement for debugging
- [x] Graceful shutdown — production requirement for deployments
- [x] Test suite — no confidence in rewrite without tests
- [x] Multi-platform static binary builds — Go's primary advantage, project's deployment model
- [x] Legacy Python code moved to `legacy/` — preserve reference implementation

### Add After Validation (v1.x)

Features to add once core is working and validated.

- [ ] Typed tool input schema — improves agent discovery, low effort after search tool works
- [ ] Engine list caching — optimization, natural after startup validation works
- [ ] Configurable engine data path — needed for flexible binary deployment
- [ ] SerpApi HTTP client with timeouts — hardening, prevents hanging connections
- [ ] Version info embedding — polish, enables `--version` flag
- [ ] Tool handler middleware chain — composable recovery/logging, clean architecture
- [ ] Request hooks — lifecycle observability without handler pollution

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] MCP Completions — improves agent UX but medium effort, requires wiring schema data
- [ ] MCP Prompts — authored content, maintenance, not core to a search proxy
- [ ] Custom transport support (SSE) — only if users demand it, unlikely

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Core MCP server (Streamable HTTP) | HIGH | LOW | P1 |
| API key auth middleware | HIGH | MEDIUM | P1 |
| Search tool | HIGH | LOW | P1 |
| Engine schemas as resources | HIGH | LOW | P1 |
| Proper MCP error responses | HIGH | LOW | P1 |
| Complete/compact modes | HIGH | LOW | P1 |
| Input validation | HIGH | MEDIUM | P1 |
| Healthcheck endpoint | MEDIUM | LOW | P1 |
| CORS support | MEDIUM | LOW | P1 |
| Startup validation | HIGH | LOW | P1 |
| Structured logging + correlation | HIGH | MEDIUM | P1 |
| Graceful shutdown | MEDIUM | MEDIUM | P1 |
| Test suite | HIGH | MEDIUM | P1 |
| Multi-platform binary builds | HIGH | MEDIUM | P1 |
| Typed tool input schema | MEDIUM | LOW | P2 |
| Engine list caching | LOW | LOW | P2 |
| Configurable engine data path | MEDIUM | LOW | P2 |
| SerpApi HTTP client timeouts | MEDIUM | LOW | P2 |
| Version info embedding | LOW | LOW | P2 |
| Tool handler middleware chain | MEDIUM | LOW | P2 |
| Request hooks | MEDIUM | LOW | P3 |
| MCP Completions | MEDIUM | MEDIUM | P3 |
| MCP Prompts | LOW | LOW | P3 |

**Priority key:**
- P1: Must have for launch — faithful port + critical bug fixes + production necessities
- P2: Should have — polish and hardening, add when core is stable
- P3: Nice to have — future consideration, doesn't block launch

## What the Python Server Did WRONG

These are the specific deficiencies identified in the Python implementation that the Go rewrite must fix:

| Python Problem | Impact | Go Fix |
|----------------|--------|--------|
| Error responses are string prefixes (`"Error: ..."`) | MCP clients can't distinguish errors from results; violates MCP spec | Use `mcp.NewToolResultError()` which sets `IsError: true` on `CallToolResult` |
| String-matching HTTP status codes (`"429" in str(e)`) | Fragile, breaks if error message format changes | Parse SerpApi HTTP response status codes directly from Go's `http.Response.StatusCode` |
| `extract_error_response()` traverses exception chain (depth 10) | Deeply coupled to Python library internals, fragile | Read HTTP response body directly from Go's `http.Response.Body` — no exception chain traversal |
| No input validation | Garbage requests forwarded to SerpApi, confusing error messages | Validate engine name against known schemas, validate mode, validate required params before API call |
| No request correlation IDs | Can't trace requests across logs; impossible to debug production issues | Inject correlation ID via context + slog; include in all log entries |
| Module-level side effects on import | 100+ engine files read on import; test isolation impossible | Load engines in `main()` or `init()`, not at package level; inject via dependency |
| `get_http_request()` for API key access | Breaks in-process testing; ties tool to HTTP transport | Pass API key through Go context (`context.WithValue`), not by reaching into HTTP request |
| No graceful shutdown | SIGTERM kills in-flight requests; deployment causes errors | `signal.NotifyContext()` + `http.Server.Shutdown()` for connection draining |
| Silent startup failures | Missing `engines/` directory → warning only, broken resources served | Validate at startup, exit with clear error if dependencies missing |
| Path rewriting for auth (`request.scope["path"]`) | Brittle Starlette-specific hack, doesn't survive middleware | Use Go's `http.StripPrefix()` or route the MCP handler at the correct path |
| No test suite | Zero confidence in any change; no regression protection | Go `testing` package + testify; test every feature |
| CloudWatch EMF for metrics | AWS-specific; doesn't fit static binary distribution model | `log/slog` structured logging; let log aggregation handle metrics |

## Competitor Feature Analysis

| Feature | mark3labs/mcp-go | Official go-sdk | trpc-mcp-go | Our Approach |
|---------|------------------|-----------------|--------------|--------------|
| Streamable HTTP | ✅ `NewStreamableHTTPServer` | ✅ `NewStreamableHTTPHandler` | ✅ built-in | Use mark3labs/mcp-go for maturity |
| Tool middleware | ✅ `WithToolHandlerMiddleware` | ✅ higher-order functions | — partial | Chain: recovery → logging → handler |
| Resource middleware | ✅ `WithResourceHandlerMiddleware` | ✅ higher-order functions | — | Logging for resource reads |
| Error handling | ✅ `NewToolResultError` | ✅ `SetError/GetError` | ✅ `NewErrorResult` | Use SDK's native error type |
| Session management | ✅ full session API | ✅ implicit | ✅ sessions | Stateless for this server |
| Hooks | ✅ `WithHooks` | — not seen | — not seen | Request lifecycle observability |
| Tool filtering | ✅ `WithToolFilter` | — not seen | ✅ `WithToolListFilter` | Could filter by engine availability |
| Recovery | ✅ `WithRecovery` | — not seen | — | Essential for production |
| Completions | ✅ `WithCompletions` | ✅ in ServerCapabilities | — | P3 differentiator |
| Auth context | Manual (context values) | Manual (middleware) | ✅ `WithHTTPContextFunc` | HTTP middleware → context |
| OAuth | — not in scope | ✅ `auth` package | — | Not needed (API key model) |

## Sources

- mark3labs/mcp-go documentation (Context7, `/mark3labs/mcp-go`): server creation, streamable HTTP, resources, tools, middleware, hooks, sessions, completions — HIGH confidence
- Official go-sdk documentation (Context7, `/websites/pkg_go_dev_github_com_modelcontextprotocol_go-sdk`): StreamableHTTPHandler, error handling, middleware, jsonschema validation — HIGH confidence
- trpc-mcp-go documentation (Context7, `/trpc-group/trpc-mcp-go`): HTTP context extraction, tool filtering, graceful shutdown — HIGH confidence
- GoReleaser documentation (Context7, `/goreleaser/goreleaser`): multi-platform binary builds, CGO_ENABLED=0, GitHub releases — HIGH confidence
- Existing Python codebase analysis: `src/server.py` error handling patterns, auth middleware, resource loading — HIGH confidence
- PROJECT.md: stated requirements, constraints, out-of-scope decisions — HIGH confidence
- Go standard library: `log/slog`, `net/http`, `os/signal`, `context` — HIGH confidence (stdlib)

---
*Feature research for: Go MCP server rewrite (SerpApi)*
*Researched: 2026-04-15*