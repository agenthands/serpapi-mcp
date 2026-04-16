---
phase: 03-search-validation-observability
plan: 02
subsystem: validation
tags: [input-validation, correlation-id, structured-logging, middleware, crypto-rand]

# Dependency graph
requires:
  - phase: 03-search-validation-observability
    provides: "Search tool handler with SerpApi HTTP client, complete/compact modes, EngineNames() and RequiredParams() accessors"
provides:
  - "Input validation functions: ValidateEngine, ValidateMode, ValidateRequiredParams"
  - "Correlation ID middleware: CorrelationIDMiddleware, CorrelationIDFromContext"
  - "Updated handler chain: CORS → correlation → auth → mux"
  - "Structured logging with correlation_id in all search log entries"
affects: [search, middleware, server, observability]

# Tech tracking
tech-stack:
  added: []
  patterns: ["validation-before-HTTP pattern (fast errors, no quota waste)", "crypto/rand correlation IDs for request tracing", "client-provided X-Correlation-ID header support"]

key-files:
  created: [internal/search/validation.go, internal/search/validation_test.go, internal/middleware/correlation.go, internal/middleware/correlation_test.go]
  modified: [internal/search/search.go, internal/server/server.go]

key-decisions:
  - "Validation runs before any SerpApi HTTP call to avoid wasting API quota on invalid requests"
  - "32-char hex correlation IDs generated with crypto/rand (not UUID) for simplicity and security"
  - "Client-provided X-Correlation-ID header honored for distributed tracing across services"

patterns-established:
  - "Validation pattern: ValidateEngine → ValidateMode → ValidateRequiredParams → HTTP call (ordered, fast-fail)"
  - "Middleware chain pattern: CORS → correlation → auth → mux (correlation before auth so auth logs also have IDs)"
  - "Correlation ID logging: all slog entries in search handler include correlation_id field"

requirements-completed: [VAL-01, VAL-02, VAL-03, OBS-01, OBS-02, OBS-03]

# Metrics
duration: 41min
completed: 2026-04-16
---

# Phase 3 Plan 2: Input Validation & Correlation IDs Summary

**Input validation (engine/mode/required params) before HTTP calls + crypto/rand correlation ID middleware for end-to-end request tracing**

## Performance

- **Duration:** 41 min
- **Started:** 2026-04-16T17:18:07Z
- **Completed:** 2026-04-16T17:59:43Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Invalid engine names rejected with error listing all 107 available engines
- Invalid mode values rejected with clear "must be 'complete' or 'compact'" message
- Missing required parameters per engine schema rejected with param name
- Validation runs before SerpApi HTTP calls (fast errors, no quota waste)
- Correlation ID middleware generates unique 32-char hex IDs per request
- Client-provided X-Correlation-ID headers honored for distributed tracing
- All search handler log entries include correlation_id field
- Handler chain updated: CORS → correlation → auth → mux

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement input validation for engine names, mode, and required parameters** - `fed0b24` (feat)
2. **Task 2: Add correlation ID middleware and structured logging throughout search flow** - `bba6b81` (feat)

## Files Created/Modified
- `internal/search/validation.go` - ValidateEngine, ValidateMode, ValidateRequiredParams functions
- `internal/search/validation_test.go` - 7 validation tests + 1 HTTP prevention integration test
- `internal/middleware/correlation.go` - CorrelationIDMiddleware and CorrelationIDFromContext
- `internal/middleware/correlation_test.go` - 6 correlation ID tests
- `internal/search/search.go` - Validation integration, correlation ID in log entries
- `internal/server/server.go` - Correlation middleware in handler chain

## Decisions Made
- Validation runs before any SerpApi HTTP call to avoid wasting API quota on invalid requests — fast-fail pattern
- Used 32-char hex IDs from crypto/rand for correlation IDs (not UUID) — simpler, no external dependency, cryptographically unique
- Client-provided X-Correlation-ID header is honored — enables distributed tracing across services
- Correlation middleware placed before auth in handler chain — auth-related logs also benefit from correlation IDs

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Added engine schema loading to existing search tests**
- **Found during:** Task 1 (validation integration)
- **Issue:** Existing search tests didn't load engine schemas; validation checks would fail since EngineNames() returned nil
- **Fix:** Added engine schema loading to setupTestServer() in search_test.go using slog.Default() logger
- **Files modified:** internal/search/search_test.go
- **Verification:** All 15 search tests pass (8 original + 7 new validation)
- **Committed in:** fed0b24 (Task 1 commit)

**2. [Rule 3 - Blocking] Fixed nil logger crash in LoadAndRegister during tests**
- **Found during:** Task 1 (test development)
- **Issue:** validation_test.go passed nil logger to engines.LoadAndRegister, causing nil pointer dereference when LoadAndRegister tried to log
- **Fix:** Changed to pass slog.Default() instead of nil
- **Files modified:** internal/search/validation_test.go
- **Verification:** Tests pass without panics
- **Committed in:** fed0b24 (Task 1 commit)

---

**Total deviations:** 2 auto-fixed (1 missing critical, 1 blocking)
**Impact on plan:** Both fixes essential for test correctness. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 03 complete: search tool + validation + observability all implemented
- Ready for Phase 04 (testing phase) — comprehensive integration and E2E tests
- All requirements VAL-01 through VAL-03 and OBS-01 through OBS-03 fulfilled

---
*Phase: 03-search-validation-observability*
*Completed: 2026-04-16*

## Self-Check: PASSED