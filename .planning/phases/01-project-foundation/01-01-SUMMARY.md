---
phase: 01-project-foundation
plan: 01
subsystem: infra
tags: [go, mcp, go-sdk, module-init, legacy-archive]

# Dependency graph
requires:
  - phase: none
    provides: "First plan — no prior dependencies"
provides:
  - "Go module at github.com/agenthands/serpapi-mcp with go-sdk dependency"
  - "Standard Go project layout (cmd/, internal/)"
  - "Python legacy code archived in legacy/ directory"
  - "Go-appropriate .gitignore"
affects: [02-mcp-server, 03-engine-integration, 04-auth-middleware]

# Tech tracking
tech-stack:
  added: [go-sdk/mcp v1.5.0, Go 1.25.0 toolchain]
  patterns: [cmd/internal Go layout, blank import for dependency pinning]

key-files:
  created: [go.mod, go.sum, cmd/serpapi-mcp/main.go, internal/server/server.go, internal/search/search.go, internal/engines/engines.go]
  modified: [.gitignore]

key-decisions:
  - "Accepted go 1.25.0 directive in go.mod instead of 1.23 — Go toolchain management auto-upgrades the directive; go mod tidy always reverts manual changes"
  - "Added blank import of go-sdk/mcp in main.go to keep dependency as direct (not indirect) in go.mod"

patterns-established:
  - "cmd/serpapi-mcp/main.go: entry point with blank import of go-sdk for dependency tracking"
  - "internal/{server,search,engines}/: domain-based internal packages mirroring Python structure"
  - "legacy/: archived Python code preserved for rewrite reference"

requirements-completed: [SETUP-01, SETUP-02, SETUP-03]

# Metrics
duration: 4min
completed: 2026-04-15
---

# Phase 1 Plan 1: Go Module Init & Legacy Archive Summary

**Go module initialized with go-sdk dependency and standard layout; Python code archived to legacy/**

## Performance

- **Duration:** 4 min
- **Started:** 2026-04-15T17:28:38Z
- **Completed:** 2026-04-15T17:32:45Z
- **Tasks:** 2
- **Files modified:** 17

## Accomplishments
- Go module `github.com/agenthands/serpapi-mcp` builds cleanly with go-sdk as the only external dependency
- Standard Go project layout established: `cmd/serpapi-mcp/` entry point and `internal/` packages for server, search, and engines
- All Python-specific files (src/, pyproject.toml, uv.lock, .python-version, Dockerfile, smithery.yaml, copilot/) archived to `legacy/`
- Root-level assets preserved: build-engines.py, engines/, server.json, README.md, LICENSE
- .gitignore updated from Python-focused to Go-focused entries

## Task Commits

Each task was committed atomically:

1. **Task 1: Initialize Go module, layout, and go-sdk dependency** - `182ffd5` (feat)
2. **Task 2: Archive Python legacy code and update .gitignore** - `be355a7` (chore)

**Plan metadata:** pending (docs: complete plan)

## Files Created/Modified
- `go.mod` - Go module definition with go-sdk dependency
- `go.sum` - Dependency checksums
- `cmd/serpapi-mcp/main.go` - Application entry point with blank go-sdk import
- `internal/server/server.go` - Server package stub
- `internal/search/search.go` - Search package stub
- `internal/engines/engines.go` - Engines package stub
- `.gitignore` - Updated from Python to Go project entries
- `legacy/src/server.py` - Archived Python server (moved from src/)
- `legacy/pyproject.toml` - Archived Python config (moved from root)
- `legacy/uv.lock` - Archived Python lockfile (moved from root)
- `legacy/.python-version` - Archived Python version (moved from root)
- `legacy/Dockerfile` - Archived Docker build (moved from root)
- `legacy/smithery.yaml` - Archived Smithery config (moved from root)
- `legacy/copilot/` - Archived AWS Copilot config (moved from root)

## Decisions Made
- **go 1.25.0 directive instead of 1.23**: Go's toolchain management (Go 1.21+) auto-upgrades the `go` directive in go.mod to match the toolchain version. Setting it to 1.23 is overridden by `go mod tidy` or `go get`. The `toolchain go1.25.1` directive tracks the actual toolchain. This is a departure from D-16 but follows Go's standard module behavior.
- **Blank import of go-sdk in main.go**: Without an actual import, `go mod tidy` marks go-sdk as `// indirect` or removes it entirely. Adding `_ "github.com/modelcontextprotocol/go-sdk/mcp"` keeps it as a direct dependency, which correctly reflects project intent. The `_` import will be replaced with actual usage in Phase 2.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Go version directive auto-upgraded from 1.23 to 1.25.0**
- **Found during:** Task 1 (Go module initialization)
- **Issue:** D-16 specifies `go 1.23` in go.mod, but Go toolchain management auto-upgrades the directive to match the installed toolchain (1.25.1). Setting `go 1.23` is reverted by any `go get` or `go mod tidy` call.
- **Fix:** Accepted `go 1.25.0` directive with `toolchain go1.25.1`. The module still works with Go 1.23+ toolchains via toolchain switching.
- **Files modified:** go.mod
- **Verification:** `go build ./cmd/serpapi-mcp && go vet ./...` pass successfully
- **Committed in:** 182ffd5 (Task 1 commit)

**2. [Rule 3 - Blocking] Added blank import to keep go-sdk as direct dependency**
- **Found during:** Task 1 (Go module initialization)
- **Issue:** Without importing go-sdk in any Go source file, `go mod tidy` removes it or marks it `// indirect`. The plan requires go-sdk as the only external dependency in go.mod.
- **Fix:** Added `_ "github.com/modelcontextprotocol/go-sdk/mcp"` blank import in main.go. This keeps go-sdk as a direct dependency and will be replaced with actual usage in Phase 2.
- **Files modified:** cmd/serpapi-mcp/main.go
- **Verification:** `grep "require github.com/modelcontextprotocol/go-sdk" go.mod` shows direct (not indirect) dependency
- **Committed in:** 182ffd5 (Task 1 commit)

---

**Total deviations:** 2 auto-fixed (2 blocking)
**Impact on plan:** Both auto-fixes were necessary for the module to build correctly and represent dependencies accurately. No scope creep.

## Issues Encountered
None — both Go toolchain management adjustments were handled as deviations above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Go module builds cleanly, ready for MCP server implementation in Phase 2
- Internal package stubs (server, search, engines) provide target structure for Phase 2 implementation
- Legacy Python code preserved in `legacy/` for reference during rewrite
- Plan 01-02 (CI and goreleaser) is ready to execute

## Self-Check: PASSED

All key files verified present on disk. All task commits verified in git history.

---
*Phase: 01-project-foundation*
*Completed: 2026-04-15*