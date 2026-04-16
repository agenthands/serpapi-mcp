---
phase: 02-server-auth-engine-resources
plan: 02
subsystem: auth
tags: [go, http.Handler, auth-middleware, bearer-token, path-based-auth, context]

# Dependency graph
requires:
  - phase: 02-server-auth-engine-resources
    plan: 01
    provides: "MCP server with Streamable HTTP transport, /health endpoint, CORS middleware"
provides:
  - "Auth middleware extracting API key from Bearer header and URL path /{KEY}/mcp"
  - "API key passed to downstream handlers via context (APIKeyFromContext helper)"
  - "401 JSON error body matching Python server format"
  - "Auth-disabled mode for testing (Config.AuthDisabled, --auth-disabled flag)"
  - "CORS → Auth → mux handler chain"
affects: [03-engine-integration, 04-search-tool]

# Tech tracking
tech-stack:
  added: []
  patterns: [http.Handler-wrapping-auth, context-based-key-passing, authOrPassthrough-conditional-middleware]

key-files:
  created: [internal/server/auth.go, internal/server/auth_test.go]
  modified: [internal/server/server.go, cmd/serpapi-mcp/main.go]

key-decisions:
  - "CORS → Auth → mux handler ordering so OPTIONS preflight bypasses auth before CORS responds with headers"
  - "authOrPassthrough helper function for Config.AuthDisabled conditional middleware wrapping"
  - "authOrPassthrough is unexported (not exported) since it's internal server wiring, not a public API"

patterns-established:
  - "internal/server/auth.go: authMiddleware wraps http.Handler, extracts Bearer header (priority) then path-based key, stores in context via custom contextKey type"
  - "internal/server/server.go: buildHandler chains CORS → authOrPassthrough(AuthDisabled, mux) — AuthDisabled flag for testing"

requirements-completed: [AUTH-01, AUTH-02, AUTH-03]

# Metrics
duration: 6min
completed: 2026-04-16
---

# Phase 2 Plan 2: Auth Middleware Summary

**API key auth middleware with Bearer header and path-based /{KEY}/mcp extraction, context-based key passing, and CORS-preflight-compatible handler chain**

## Performance

- **Duration:** 6 min
- **Started:** 2026-04-16T15:03:54Z
- **Completed:** 2026-04-16T15:10:12Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Auth middleware extracts API key from Authorization: Bearer header (priority over path-based per D-01)
- Auth middleware extracts API key from URL path /{KEY}/mcp and rewrites path to /mcp (per D-03)
- 401 JSON error response matching Python server's format: `{"error":"Missing API key. Use path format /{API_KEY}/mcp or Authorization: Bearer {API_KEY} header"}` (per D-04)
- Context-based API key passing via custom context key type (per D-02), with APIKeyFromContext helper
- Handler chain: CORS → Auth → mux — OPTIONS preflight bypasses auth before CORS responds with 204
- AuthDisabled config flag for testing (CLI --auth-disabled, env MCP_AUTH_DISABLED)
- 6 integration tests verifying full handler chain behavior

## Task Commits

Each task was committed atomically:

1. **Task 1: Create auth middleware with path-based and Bearer header key extraction** (TDD)
   - `a4c1492` (test) — RED: failing tests for Bearer header, path-based, missing key, health exempt, priority, invalid format
   - `b7b7459` (feat) — GREEN: auth.go implementation with authMiddleware, APIKeyFromContext, authErrorResponse
2. **Task 2: Wire auth middleware into server handler chain**
   - `3150e26` (feat) — Auth wired into buildHandler, AuthDisabled flag, integration tests, envBoolOr helper

**Plan metadata:** pending (docs: complete plan)

## Files Created/Modified
- `internal/server/auth.go` — authMiddleware, APIKeyFromContext, authOrPassthrough, context key type, 401 JSON error
- `internal/server/auth_test.go` — 9 unit tests + 6 integration tests for auth middleware and handler chain
- `internal/server/server.go` — Config.AuthDisabled field, buildHandler updated with authOrPassthrough in chain
- `cmd/serpapi-mcp/main.go` — --auth-disabled flag, MCP_AUTH_DISABLED env var, envBoolOr helper, strings import

## Decisions Made
- **CORS → Auth → mux handler ordering:** CORS outermost so OPTIONS preflight requests get CORS headers and 204 before auth could reject them. This matches the plan's deliberation and standard Go middleware patterns.
- **authOrPassthrough unexported:** The conditional middleware helper is internal server wiring, not a public API surface. Config.AuthDisabled controls it.
- **envBoolOr accepts "1"/"true"/"yes":** Standard Go convention for boolean environment variables, matching common CLI patterns.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None — the TDD flow worked cleanly. RED tests failed with compilation errors (expected), GREEN implementation passed all tests on first run.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Auth middleware fully functional, ready for search tool to extract API key via APIKeyFromContext
- Handler chain: CORS → Auth → mux — ready for engine resources to be added at /mcp endpoint
- AuthDisabled flag available for future integration testing without needing API keys

## Self-Check: PASSED

All key files verified present on disk. All task commits verified in git history.

---
*Phase: 02-server-auth-engine-resources*
*Completed: 2026-04-16*