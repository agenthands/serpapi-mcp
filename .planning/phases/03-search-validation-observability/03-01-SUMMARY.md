---
phase: 03-search-validation-observability
plan: 01
subsystem: search
tags: [serpapi, mcp, http-client, error-handling, compact-mode]

# Dependency graph
requires:
  - phase: 02-server-auth-engine-resources
    provides: "MCPServer, auth middleware with APIKeyFromContext, engine resource loading"
provides:
  - "Search tool handler with SerpApi HTTP client, complete/compact modes, error mapping"
  - "EngineNames() and RequiredParams() accessor functions for validation"
  - "ContextWithAPIKey helper for testability"
affects: [search, validation, observability]

# Tech tracking
tech-stack:
  added: []
  patterns: ["toolError flat JSON error body for MCP-compliant errors", "serpapiBaseURL package var for test override", "httptest mock server for SerpApi test doubles"]

key-files:
  created: [internal/search/search.go, internal/search/search_test.go]
  modified: [internal/engines/engines.go, internal/server/auth.go, cmd/serpapi-mcp/main.go]

key-decisions:
  - "toolError helper returns flat JSON {error, message} body with IsError=true instead of using SetError string prefix"
  - "serpapiBaseURL as package-level var (not env var) for test override -- simpler, no env pollution"
  - "Added ContextWithAPIKey helper to server package for test injection of API keys"

patterns-established:
  - "MCP tool error pattern: toolError(code, message) returns CallToolResult with flat JSON body and IsError=true"
  - "MCP tool registration pattern: RegisterXTool(srv *mcp.Server, logger *slog.Logger) with srv.AddTool"
  - "Test context setup: server.ContextWithAPIKey(ctx, key) to inject API keys in test contexts"

requirements-completed: [SRCH-01, SRCH-02, SRCH-03, SRCH-04, SRCH-05, SRCH-06, SRCH-07]

# Metrics
duration: 13min
completed: 2026-04-16
---

# Phase 3 Plan 1: Search Tool Implementation Summary

**SerpApi search tool with complete/compact modes, MCP-compliant error handling (429/401/403/5xx→IsError), and API key extraction from request context**

## Performance

- **Duration:** 13 min
- **Started:** 2026-04-16T17:00:01Z
- **Completed:** 2026-04-16T17:13:56Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Search tool handler calls SerpApi HTTP API with complete/compact response modes
- All HTTP error codes (429/401/403/5xx) mapped to MCP-compliant IsError=true responses with flat JSON error bodies
- API key extracted from request context via APIKeyFromContext for per-request authentication
- Default engine set to google_light, default mode to complete
- EngineNames() and RequiredParams() accessor functions added for Plan 02 validation
- ContextWithAPIKey helper added to server package for testability

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement search tool handler** - `34436c9` (feat)
2. **Task 2: Register search tool on MCP server** - `d2794ba` (feat)

## Files Created/Modified
- `internal/search/search.go` - Search tool handler with SerpApi client, mode handling, error mapping
- `internal/search/search_test.go` - 8 test functions covering all 7 behavior specs plus default engine test
- `internal/engines/engines.go` - Added EngineNames(), RequiredParams() accessor functions and schema cache stores
- `internal/server/auth.go` - Added ContextWithAPIKey helper for test context injection
- `cmd/serpapi-mcp/main.go` - Search tool registration in main initialization

## Decisions Made
- Used toolError helper returning flat JSON {error, message} body with IsError=true instead of SetError string prefix — matches Python server's error JSON format
- Used serpapiBaseURL as package-level var (not env var) for test override — simpler, no env pollution, test can mutate directly
- Added ContextWithAPIKey helper to server package — unexported contextKey type prevents external test injection without it

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Added ContextWithAPIKey helper to server package**
- **Found during:** Task 1 (test development)
- **Issue:** apiKeyContextKey is unexported in server package; tests cannot inject API keys into context without a helper
- **Fix:** Added ContextWithAPIKey(ctx, key) function to server/auth.go that wraps context.WithValue with the unexported key
- **Files modified:** internal/server/auth.go
- **Verification:** Tests use ContextWithAPIKey to set API keys; all 8 tests pass
- **Committed in:** 34436c9 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 missing critical)
**Impact on plan:** Minor addition essential for testability. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Search tool registered and callable via MCP server
- Ready for Plan 02: input validation (engine name, required params, mode validation) using EngineNames() and RequiredParams() accessors
- Observability (structured logging with request correlation IDs) will be added in Plan 02

---
*Phase: 03-search-validation-observability*
*Completed: 2026-04-16*

## Self-Check: PASSED