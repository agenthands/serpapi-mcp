---
phase: 02-server-auth-engine-resources
plan: 03
subsystem: engines
tags: [go, mcp, go-sdk, engine-schemas, resource-registration, validation]

# Dependency graph
requires:
  - phase: 02-server-auth-engine-resources
    provides: "MCP server with Streamable HTTP transport, MCPServer.MCPServer exposed as *mcp.Server"
provides:
  - "Engine loading with startup validation (107 engine JSON schemas)"
  - "serpapi://engines resource listing all engines with count, names, and resource URIs"
  - "serpapi://engines/{name} per-engine resources serving full JSON schemas"
  - "Fail-fast on corrupt or missing engine JSON (ENG-04)"
  - "--engines-dir CLI flag with ENGINES_DIR env var override"
affects: [04-search-tool, 05-observability]

# Tech tracking
tech-stack:
  added: [os.ReadDir, os.ReadFile, regexp engine-name-validation, json.RawMessage in-memory storage]
  patterns: [LoadAndRegister-entry-point, fail-fast-validation, per-engine-resource-handler-closures, sorted-engine-names]

key-files:
  created: [internal/engines/engines.go, internal/engines/engines_test.go]
  modified: [cmd/serpapi-mcp/main.go, internal/server/server.go]

key-decisions:
  - "LoadAndRegister takes enginesDir as parameter (not hardcoded) for testability and CLI/ENV override"
  - "Engine name validation uses [a-z0-9_]+ regex, skipping invalid filenames with warning log"
  - "Engine field must match filename stem, returning error on mismatch (stricter than Python version)"
  - "MCPServer.SetEngineCount method added for startup logging with engine count"

patterns-established:
  - "internal/engines/engines.go: LoadAndRegister loads, validates, and registers — single entry point"
  - "Resource handler closures capture pre-serialized JSON string to avoid re-serialization on each read"
  - "cmd/serpapi-mcp/main.go: fail-fast on LoadAndRegister error (os.Exit(1))"

requirements-completed: [ENG-01, ENG-02, ENG-03, ENG-04, ENG-05]

# Metrics
duration: 19min
completed: 2026-04-16
---

# Phase 2 Plan 3: Engine Loading & MCP Resource Registration Summary

**107 engine JSON schemas loaded at startup with validation, serpapi://engines index resource, and per-engine serpapi://engines/{name} resources via go-sdk AddResource**

## Performance

- **Duration:** 19 min
- **Started:** 2026-04-16T15:06:57Z
- **Completed:** 2026-04-16T15:25:54Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- LoadAndRegister loads all 107 engine JSON schemas, validates each, registers MCP resources
- serpapi://engines resource returns JSON with count, engines list, resource URIs, and schema info
- serpapi://engines/{name} resources return per-engine full JSON schemas
- Fail-fast validation: corrupt JSON, missing directory, engine field mismatch all return errors
- Engine filename validation: only [a-z0-9_]+.json accepted, invalid names skipped with warning
- --engines-dir CLI flag with ENGINES_DIR env var override, startup log includes engine count
- Comprehensive test suite using go-sdk InMemoryTransport for real MCP resource reads
- Integration test verifying all 107 real engines load successfully

## Task Commits

Each task was committed atomically:

1. **Task 1: Create engine loading with startup validation and MCP resource registration** (TDD)
   - `a84d637` (feat) — engines.go, engines_test.go, main.go, server.go

2. **Task 2: Integration test — full server startup with engines and healthcheck**
   - `6798b8b` (test) — Integration tests in server_test.go and engines_test.go

## Files Created/Modified
- `internal/engines/engines.go` — LoadAndRegister, registerEnginesIndex, registerEngineResource with validation
- `internal/engines/engines_test.go` — 9 tests covering valid/missing/corrupt/invalid/empty/mismatch scenarios plus real 107-engine test
- `cmd/serpapi-mcp/main.go` — Added --engines-dir flag, ENGINES_DIR env var, LoadAndRegister call, SetEngineCount
- `internal/server/server.go` — Added engineCount field, SetEngineCount method, updated startup log

## Decisions Made
- **LoadAndRegister takes enginesDir as parameter** — not hardcoded, enabling CLI flag and env var override with clean testability
- **Engine field must match filename stem** — stricter than Python version which didn't validate this; returns error on mismatch for data integrity
- **SetEngineCount method on MCPServer** — allows main.go to set the engine count after LoadAndRegister, server.go uses it in startup log rather than hardcoding 0
- **Resource handler closures capture pre-serialized JSON** — avoids re-serialization on every resource read; the JSON is serialized once during LoadAndRegister

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None — the go-sdk AddResource API matched the plan's interface expectations perfectly.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Engine schemas loaded and accessible via MCP resources, ready for search tool parameter validation
- serpapi://engines/{name} resources can be used to discover valid parameters per engine
- CLI flag --engines-dir allows custom engine directories for testing or deployment

## Self-Check: PASSED

All key files verified present on disk. All task commits verified in git history.

---
*Phase: 02-server-auth-engine-resources*
*Completed: 2026-04-16*