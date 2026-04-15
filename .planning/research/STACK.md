# Technology Stack

**Project:** SerpApi MCP Server (Go Rewrite)
**Researched:** 2026-04-15
**Mode:** Ecosystem (Stack dimension for Go MCP server)
**Previous stack (reference only):** FastMCP on Starlette + uvicorn, Python 3.13 — being replaced

## Recommended Stack

### Core Language & Runtime

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.25+ | Language runtime | Required by modelcontextprotocol/go-sdk v1.4.1+ (uses `http.CrossOriginProtection`); provides enhanced ServeMux, mature slog, range-over-func; Go 1.25 was released Aug 2025 — stable toolchain as of April 2026 |

### MCP SDK

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| modelcontextprotocol/go-sdk | v1.5.0+ | MCP server & client SDK | Official MCP org SDK maintained with Google; stable v1.x API with backward compatibility guarantees; `NewStreamableHTTPHandler` returns `http.Handler` for natural middleware composition; `getServer func(*http.Request) *Server` pattern enables per-request API key injection; typed tool handlers with generics (`AddTool[In, Out]`); auto JSON schema generation from Go struct `jsonschema` tags; built-in DNS rebinding protection and cross-origin request validation; actively tracks MCP spec (supports 2025-11-25) |

### HTTP Transport

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| net/http | stdlib | HTTP server, routing, middleware chain | Go 1.22+ enhanced ServeMux supports method-based routing (`GET /healthcheck`); only 2 routes needed (MCP endpoint + healthcheck) — no framework justified; official SDK returns `http.Handler` which composes directly with `http.HandleFunc` patterns |
| CORS middleware | hand-written (~20 lines) | CORS headers | Policy is trivial: allow all origins, credentials, methods, headers. Writing a middleware function is simpler and zero-dependency vs pulling in `rs/cors` for 3 header sets. No origin-matching logic needed. |

### Logging

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| log/slog | stdlib (Go 1.21+) | Structured JSON logging with request correlation | Zero dependencies; `slog.Logger` is goroutine-safe; JSON handler built-in; this is a network-IO-bound proxy service — zerolog/zap's allocation benchmarks are irrelevant when each request spends milliseconds waiting for SerpApi; `slog.With("request_id", id)` gives correlation IDs without custom code |

### Build & Release

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| goreleaser | v2.x | Cross-compilation, release automation, GitHub Releases | De-facto standard for Go binary distribution; handles all 5 target platforms with `CGO_ENABLED=0`; generates checksums, archives, changelogs; one-command `goreleaser release` workflow; no manual `GOOS=GOARCH=` scripts to maintain |
| golangci-lint | v1.64+ | Static analysis aggregator | Runs 50+ linters in parallel; replaces individual tools (staticcheck, go vet, errcheck, gosec, etc.); CI-ready with `.golangci.yml` config; faster than running linters separately |

### Testing

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| testing | stdlib | Unit tests, table-driven tests | No framework needed; Go's `testing` package is sufficient; table-driven test pattern for engine validation and error cases |
| net/http/httptest | stdlib | HTTP handler tests | `httptest.NewServer` and `httptest.NewRecorder` for testing middleware (auth, CORS) and MCP handler integration against real HTTP requests without network |
| gotest.tools/v3 | v3.5+ | Assertion helpers | `assert.DeepEqual`, `assert.NilError`, `assert.Check` provide clearer failure messages than `t.Fatal()` without the opinionated framework overhead of testify; minimal dependency surface |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| encoding/json | stdlib | JSON marshaling/unmarshaling | All JSON handling: engine schemas, search results, MCP responses; `json.Marshal` and `json.Unmarshal` are sufficient — no third-party JSON library needed |
| os / path/filepath | stdlib | File I/O | Loading engine JSON schemas from `engines/` directory at startup; `filepath.Glob("engines/*.json")` for discovery |
| net/http | stdlib | SerpApi HTTP client | Call SerpApi search API directly via `http.Client` — no Go SerpApi client library needed (the Python `serpapi` package just wraps HTTP calls) |
| context | stdlib | Request scoping, cancellation, API key propagation | `context.WithValue` for per-request API key; standard Go pattern; timeout propagation to SerpApi calls |

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| MCP SDK | modelcontextprotocol/go-sdk | **mark3labs/mcp-go** v0.48.0 | **Pre-v1 — no API stability guarantee.** Community-maintained, not official; `NewStreamableHTTPServer.Start()` wraps HTTP lifecycle making middleware composition harder (our auth middleware would need `WithHTTPContextFunc` workarounds); acknowledged as inspiration by official SDK's README; risk of spec divergence as official SDK evolves; 8.6k stars demonstrate strong community but don't guarantee API stability |
| MCP SDK | modelcontextprotocol/go-sdk | **trpc-group/trpc-mcp-go** | Niche SDK from Tencent's tRPC ecosystem; 73 code snippets vs 811 for official SDK; brings tRPC dependency tree; no compelling advantage for our use case; lower community adoption and documentation |
| HTTP Framework | net/http stdlib | **go-chi/chi** v5 | Unnecessary — we have 2 routes (MCP endpoint + healthcheck); Go 1.22 ServeMux handles method-based routing; chi adds a dependency for features we don't use (middleware groups, URL parameter extraction, route groups) |
| HTTP Framework | net/http stdlib | **gin-gonic/gin** | Massive overkill — gin's radix-tree router and middleware chain are designed for REST APIs with dozens of endpoints; we have 2 routes; gin would add ~100KB dependency for zero benefit; gin's `context.Context` replacement breaks stdlib patterns |
| HTTP Framework | net/http stdlib | **labstack/echo** | Same reasoning as gin — framework overhead not justified for 2-endpoint server; echo's `echo.Context` is non-standard |
| Logging | log/slog | **rs/zerolog** | Zerolog's zero-allocation advantage is irrelevant for a network-IO-bound proxy that spends 50-500ms per SerpApi call; adds dependency for no measurable benefit; slog produces identical JSON output; slog is stdlib — no version skew risk |
| Logging | log/slog | **uber-go/zap** | Same as zerolog — CPU-bound benchmarking advantage doesn't apply; dependency adds coupling; slog has identical capabilities for our use case |
| CORS | hand-written middleware | **rs/cors** | 3 header sets with no complex origin-matching logic; 20-line middleware function vs dependency; rs/cors adds configuration complexity we don't need |
| Testing | gotest.tools/v3 | **stretchr/testify** | testify's suite pattern and assert/require dichotomy add more API surface than needed; gotest.tools is composable, lighter, and designed for stdlib-first Go testing; testify's `assert` vs `require` distinction creates inconsistent test patterns |
| Build | goreleaser v2 | **manual go build scripts** | goreleaser handles checksums, archives, GitHub Release integration, and Homebrew taps automatically; manual scripts would reimplement all of this poorly and break across platforms |
| Build | goreleaser v2 | **goreleaser v1** (deprecated) | v2 has better cross-compilation, cleaner YAML, and is the current release; v1 is no longer maintained |

## Architecture Integration

How the stack composes into the server architecture:

```
Request Flow:
  HTTP Request
    → net/http ServeMux (route: /{KEY}/mcp, /mcp, /healthcheck)
    → API Key Middleware (extract key from path/header, rewrite path)
    → CORS Middleware (add Access-Control-Allow-* headers)
    → Request Logging Middleware (slog with request_id, method, path)
    → StreamableHTTPHandler (modelcontextprotocol/go-sdk)
      → MCP Server (tools and resources registered)
        → search tool handler (calls SerpApi HTTP API)
        → resource handlers (serve engine JSON schemas)
```

### Key Composition Pattern: Per-Request API Key Injection

The official SDK's `NewStreamableHTTPHandler` takes a `getServer func(*http.Request) *mcp.Server`. This is the natural injection point for the existing Python auth model — resolve the API key *before* the MCP server sees the request:

```go
handler := mcp.NewStreamableHTTPHandler(
    func(r *http.Request) *mcp.Server {
        apiKey := r.Header.Get("X-API-Key") // set by our middleware
        return newMCPServer(apiKey)          // server with key baked in
    },
    &mcp.StreamableHTTPOptions{
        // Stateless: true — matches Python's stateless_http=True
    },
)
```

**Why this is better than mark3labs/mcp-go's context approach:**
- Auth failures return `401` at the HTTP layer, not as MCP errors
- The API key is resolved before the MCP server even exists
- No `context.WithValue` needed for the key between middleware and tool handler
- Matches the existing Python design where `request.state.api_key` is set by middleware

### Middleware Chain Pattern

```go
mux := http.NewServeMux()

// Healthcheck (no auth required)
mux.HandleFunc("GET /healthcheck", healthcheckHandler)

// MCP endpoint (auth required)
mcpHandler := mcp.NewStreamableHTTPHandler(getServer, &mcp.StreamableHTTPOptions{})
mux.Handle("/{key}/mcp", chain(mcpHandler, apiKeyMiddleware, corsMiddleware, loggingMiddleware))
```

## Go Version Rationale

The official SDK **requires Go 1.25+** as of v1.4.1 (March 2026). This is because v1.4.1 added cross-origin request protection using `http.CrossOriginProtection`, a new API introduced in Go 1.25.

**This is not a constraint — it's aligned with current Go releases.** Go 1.25 was released August 2025; by April 2026, Go 1.25 (or 1.26) is the stable toolchain.

Benefits of targeting Go 1.25:
- Enhanced `http.ServeMux` with method-based routing patterns (`GET /healthcheck`)
- `log/slog` is mature and proven (3+ years in stdlib)
- `iter` package and range-over-func
- `http.CrossOriginProtection` for MCP security
- `synctest` for testing concurrent/deterministic code
- `os.Root` for safe filesystem access (if needed for engine schema validation)

## goreleaser Configuration

```yaml
# .goreleaser.yaml
builds:
  - main: ./cmd/serpapi-mcp
    binary: serpapi-mcp
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: checksums.txt

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
```

## Installation

```bash
# Initialize Go module
go mod init github.com/agenthands/serpapi-mcp

# Core dependency
go get github.com/modelcontextprotocol/go-sdk/mcp@v1.5.0

# Dev tooling
go install github.com/goreleaser/goreleaser/v2@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Test helper
go get gotest.tools/v3@latest
```

## Dependency Summary

### Production Dependencies

| Package | Version | Purpose | From |
|---------|---------|---------|------|
| modelcontextprotocol/go-sdk/mcp | v1.5.0+ | MCP server SDK | go get |

### Zero External Production Dependencies

Everything else is stdlib: `net/http`, `encoding/json`, `log/slog`, `os`, `path/filepath`, `context`, `time`. No SerpApi Go client library needed — we call SerpApi's HTTP API directly via `net/http`, which is what the Python `serpapi` package does internally anyway.

### Dev Dependencies

| Package | Version | Purpose | Install |
|---------|---------|---------|---------|
| goreleaser | v2.x | Build & release | go install |
| golangci-lint | v1.64+ | Static analysis | go install |
| gotest.tools/v3 | v3.5+ | Test assertions | go get (dev) |

## Sources

- **modelcontextprotocol/go-sdk v1.5.0** — https://github.com/modelcontextprotocol/go-sdk — **HIGH confidence**: Context7 verified (`/websites/pkg_go_dev_github_com_modelcontextprotocol/go-sdk`), pkg.go.dev API docs confirmed, GitHub releases page verified v1.5.0 released Apr 7, 2026, 4.4k stars, 22 releases, official MCP org + Google collaboration
- **mark3labs/mcp-go v0.48.0** — https://github.com/mark3labs/mcp-go — **HIGH confidence**: Context7 verified (`/mark3labs/mcp-go`), evaluated and not recommended; 8.6k stars, pre-v1 API, acknowledged as inspiration by official SDK
- **trpc-group/trpc-mcp-go** — https://github.com/trpc-group/trpc-mcp-go — **MEDIUM confidence**: Context7 verified (`/trpc-group/trpc-mcp-go`), evaluated and not recommended; niche ecosystem
- **GoReleaser** — https://goreleaser.com — **HIGH confidence**: Context7 verified (`/goreleaser/goreleaser`), official docs confirmed cross-compilation config structure
- **Go slog** — https://pkg.go.dev/log/slog — **HIGH confidence**: stdlib, verified
- **Go net/http CrossOriginProtection** — modelcontextprotocol/go-sdk v1.4.1 release notes — **HIGH confidence**: verified from GitHub releases page
- **gotest.tools/v3** — https://github.com/gotestyourself/gotest.tools — **MEDIUM confidence**: widely used in Go ecosystem, not in Context7
- **Go 1.25 release** — Go release policy (two newest versions) + official SDK Go 1.25 requirement — **HIGH confidence**: verified from official SDK go.mod and release notes