---
phase: 02-server-auth-engine-resources
plan: 01
subsystem: server
tags: [go, mcp, go-sdk, streamable-http, healthcheck, cors, graceful-shutdown]

# Dependency graph
requires:
  - phase: 01-project-foundation
    provides: "Go module with standard layout and go-sdk dependency"
provides:
  - "MCP server with Streamable HTTP transport at /mcp endpoint"
  - "Healthcheck endpoint at /health returning JSON status"
  - "CORS middleware with configurable origins (default allow_origins=[\"*\"])"
  - "Graceful shutdown on SIGINT/SIGTERM with 10s timeout"
  - "CLI flags (--host, --port, --cors-origins) with env var overrides"
affects: [03-engine-integration, 04-auth-middleware]

# Tech tracking
tech-stack:
  added: [go-sdk/mcp StreamableHTTPHandler, net/http.ServeMux, signal.NotifyContext]
  patterns: [corsMiddleware-wrapping-handler, flag+env-config, listener-based-startup, context-based-graceful-shutdown]

key-files:
  created: [internal/server/server.go, internal/server/cors.go, internal/server/server_test.go, internal/server/cors_test.go]
  modified: [cmd/serpapi-mcp/main.go]

key-decisions:
  - "Disabled SDK's DisableLocalhostProtection on StreamableHTTPOptions to allow non-localhost connections (MCP clients connect remotely)"
  - "Used net.Listen before http.Server.Serve for immediate port discovery (useful when Port=0)"
  - "CLI flags take precedence over env vars per D-07; both supported for --host, --port, --cors-origins"

patterns-established:
  - "internal/server/server.go: NewMCPServer creates mcp.Server + StreamableHTTPHandler, Run starts HTTP server with graceful shutdown"
  - "internal/server/cors.go: CORS middleware wraps http.Handler chain with configurable origins"
  - "cmd/serpapi-mcp/main.go: flag.Parse → envOr fallbacks → NewMCPServer → signal.NotifyContext → Run"

requirements-completed: [MCP-01, MCP-02, MCP-03, MCP-04]

# Metrics
duration: 8min
completed: 2026-04-16
---

# Phase 2 Plan 1: MCP Server with Streamable HTTP, Healthcheck & CORS Summary

**Go MCP server with Streamable HTTP transport, /health JSON endpoint, configurable CORS middleware, and graceful shutdown via signal.NotifyContext**

## Performance

- **Duration:** 8 min
- **Started:** 2026-04-16T14:46:47Z
- **Completed:** 2026-04-16T14:54:48Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- MCP server creates go-sdk mcp.Server with StreamableHTTPHandler (stateless, JSON response mode)
- Healthcheck at /health returns 200 OK with `{"status":"healthy","service":"SerpApi MCP Server"}`
- CORS middleware wraps entire handler chain with default allow_origins=["*"], handles OPTIONS preflight with 204
- Graceful shutdown on SIGINT/SIGTERM with 10-second timeout via context cancellation
- CLI flags (--host, --port, --cors-origins) with env var overrides (MCP_HOST, MCP_PORT, MCP_CORS_ORIGINS)

## Task Commits

Each task was committed atomically:

1. **Task 1: Create MCP server setup with Streamable HTTP transport and healthcheck** (TDD)
   - `6bb9818` (test) — RED: failing tests for NewMCPServer, healthcheck, graceful shutdown
   - `31e9a4d` (feat) — GREEN: server.go, cors.go, main.go implementation
2. **Task 2: Add CORS middleware with configurable origins**
   - `679a78a` (feat) — CORS tests and main.go cleanup

**Plan metadata:** pending (docs: complete plan)

## Files Created/Modified
- `internal/server/server.go` - NewMCPServer, MCPServer.Run, buildHandler, healthResponse
- `internal/server/cors.go` - CORSConfig, NewCORSConfig, corsMiddleware
- `internal/server/server_test.go` - Tests for NewMCPServer, healthcheck, graceful shutdown
- `internal/server/cors_test.go` - Tests for CORSConfig parsing, middleware headers, preflight
- `cmd/serpapi-mcp/main.go` - Entry point with CLI flags, env vars, signal handling, server.Run

## Decisions Made
- **Disabled SDK's localhost protection:** Set `DisableLocalhostProtection: true` on StreamableHTTPOptions because MCP clients may connect from non-localhost origins when the server is deployed; the Python server had wide-open CORS with no origin restrictions
- **Used net.Listen before Serve:** Started listening on a separate net.Listener before calling http.Server.Serve, which allows immediate port discovery (e.g., when Port=0 for testing or dynamic allocation)
- **CLI flags take precedence over env vars:** Per D-07, `flag.String("host", envOr("MCP_HOST", "0.0.0.0"), ...)` — env vars provide defaults, flags override

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] CORS middleware created in Task 1 (not Task 2)**
- **Found during:** Task 1 (MCP server implementation)
- **Issue:** server.go's buildHandler references corsMiddleware and NewCORSConfig from cors.go, which was planned for Task 2. The server would not compile without CORS support.
- **Fix:** Created cors.go alongside server.go in Task 1's GREEN phase so the code compiles and tests pass. Task 2 then added the dedicated CORS tests.
- **Files modified:** internal/server/cors.go (created in Task 1)
- **Verification:** All tests pass including healthcheck which goes through the CORS middleware
- **Committed in:** 31e9a4d (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** CORS middleware was simply created earlier than planned (Task 1 vs Task 2) because the server handler chain requires it. Task 2 then focused on adding comprehensive CORS tests. No scope creep.

## Issues Encountered
None — the CORS implementation dependency between tasks was resolved by creating cors.go in Task 1.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- MCP server running on configurable host:port with healthcheck and CORS
- Ready for Plan 02-02: API key auth middleware (path-based and Bearer header)
- Ready for Plan 02-03: Engine schema loading and MCP resource registration
- Server handler chain: CORS → mux (/mcp, /health) — auth middleware will insert between CORS and mux

## Self-Check: PASSED

All key files verified present on disk. All task commits verified in git history.

---
*Phase: 02-server-auth-engine-resources*
*Completed: 2026-04-16*