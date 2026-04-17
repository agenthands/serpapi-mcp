---
phase: 04-testing-release
plan: 03
subsystem: testing, ci, infra
tags: [go, golangci-lint, ci, coverage, refactor, tdd]

# Dependency graph
requires:
  - phase: 04-testing-release
    provides: existing test suite and CI workflow
provides:
  - CI workflow with correct Go 1.25.x version, push trigger, and module caching
  - Makefile lint and vet targets matching CI steps
  - Dynamic engine count test (no hardcoding)
  - Extracted run() function for cmd/serpapi-mcp testability
  - 83.8% coverage for cmd/serpapi-mcp package (up from ~40%)
affects: [ci, makefile, cmd/serpapi-mcp, internal/engines]

# Tech tracking
tech-stack:
  added: []
  patterns: [extracted-run-function-for-testability, flag-NewFlagSet-for-test-isolation, dynamic-directory-count-in-tests]

key-files:
  created: []
  modified:
    - .github/workflows/ci.yml
    - Makefile
    - internal/engines/engines_test.go
    - cmd/serpapi-mcp/main.go
    - cmd/serpapi-mcp/main_test.go

key-decisions:
  - "run() accepts ctx, args, stdout, stderr for full testability instead of reading os.Args directly"
  - "flag.NewFlagSet replaces global flag package so tests can parse custom args without side effects"
  - "Dynamic os.ReadDir count in engine tests eliminates brittle hardcoded 107/108 values"

patterns-established:
  - "Extracted run() pattern: main() as trivial 2-line wrapper, run() holds all testable logic"
  - "FlagSet isolation: use flag.NewFlagSet in extracted functions for test-safe arg parsing"
  - "Dynamic directory counting: use os.ReadDir + filter by .json extension instead of hardcoding counts"

requirements-completed: [TEST-03, TEST-04]

# Metrics
duration: 2min
completed: 2026-04-17
---

# Phase 04 Plan 03: Fix CI Version, Coverage & Test Quality Summary

**CI Go 1.25.x fix with push trigger/caching, Makefile lint/vet targets, dynamic engine count test, and 83.8% cmd/serpapi-mcp coverage via extracted run() function**

## Performance

- **Duration:** 2 min
- **Started:** 2026-04-17T16:51:45Z
- **Completed:** 2026-04-17T16:53:40Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Fixed critical CI Go version mismatch (1.23.x → 1.25.x matching go.mod)
- Added push trigger and module caching to CI workflow
- Added Makefile lint and vet targets matching CI steps
- Replaced hardcoded engine count (107/108) with dynamic os.ReadDir-based counting
- Extracted testable run() from main(), achieving 83.8% cmd/serpapi-mcp coverage
- Added integration tests for startup/shutdown and bad engines dir error paths

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix CI Go version, add Makefile lint/vet targets, fix hardcoded engine count** - `f930bd9` (fix)
2. **Task 2: Extract run() from main() and add integration tests for 83.8% coverage** - `bba724c` (feat)

## Files Created/Modified
- `.github/workflows/ci.yml` - Go 1.25.x, cache: true, push trigger on main
- `Makefile` - Added lint and vet targets
- `internal/engines/engines_test.go` - Dynamic engine count using os.ReadDir + strings.HasSuffix
- `cmd/serpapi-mcp/main.go` - Extracted run(ctx, args, stdout, stderr) with flag.NewFlagSet
- `cmd/serpapi-mcp/main_test.go` - Added TestRunStartupWithGracefulShutdown and TestRunBadEnginesDir

## Decisions Made
- run() accepts ctx, args, stdout, stderr for full testability instead of reading os.Args directly — allows integration tests without manipulating process state
- flag.NewFlagSet replaces global flag package so tests can parse custom args without side effects — avoids test pollution between parallel runs
- Dynamic os.ReadDir count in engine tests eliminates brittle hardcoded 107/108 values — survives engine additions/removals

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All Phase 04 verification gaps closed
- CI now matches go.mod Go version, triggers on push and PR
- cmd/serpapi-mcp coverage well above 70% threshold (83.8%)
- Engine tests are maintainable (dynamic count, no hardcoding)
- No blockers or concerns

---
*Phase: 04-testing-release*
*Completed: 2026-04-17*

## Self-Check: PASSED

- All 5 modified files exist on disk
- Both task commits found in git log (f930bd9, bba724c)
- All tests pass with race detector
- cmd/serpapi-mcp coverage confirmed at 83.8%