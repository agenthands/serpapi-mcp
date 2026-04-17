# Milestones

## v1.0 Go Rewrite MVP (Shipped: 2026-04-17)

**Phases completed:** 4 phases, 10 plans, 20 tasks

**Key accomplishments:**

- Go module initialized with go-sdk dependency and standard layout; Python code archived to legacy/
- GitHub Actions CI with golangci-lint/go vet/go test on PRs, goreleaser for 5-platform static binary builds
- Go MCP server with Streamable HTTP transport, /health JSON endpoint, configurable CORS middleware, and graceful shutdown via signal.NotifyContext
- API key auth middleware with Bearer header and path-based /{KEY}/mcp extraction, context-based key passing, and CORS-preflight-compatible handler chain
- 107 engine JSON schemas loaded at startup with validation, serpapi://engines index resource, and per-engine serpapi://engines/{name} resources via go-sdk AddResource
- SerpApi search tool with complete/compact modes, MCP-compliant error handling (429/401/403/5xx→IsError), and API key extraction from request context
- Input validation (engine/mode/required params) before HTTP calls + crypto/rand correlation ID middleware for end-to-end request tracing
- Extended unit test coverage for search tool edge cases, validation boundaries, engine accessors, and CLI env helpers — 18 new test functions across 4 files
- 7 new integration tests for auth edge cases and healthcheck, race-free verification, 81.2% coverage, Makefile with test/test-race/cover targets, CI with -race flag
- CI Go 1.25.x fix with push trigger/caching, Makefile lint/vet targets, dynamic engine count test, and 83.8% cmd/serpapi-mcp coverage via extracted run() function

---
