---
phase: 07-installation-usage
plan: 02
subsystem: documentation
tags: [usage, configuration, auth, mcp-clients, engines, errors, troubleshooting]

# Dependency graph
requires:
  - phase: 07-installation-usage
    plan: 01
    provides: INSTALL.md for cross-linking
provides:
  - USAGE.md covering configuration, auth, MCP client integration, engine discovery, error reference, and troubleshooting
affects: [end-users, mcp-client-integrators]

# Tech tracking
tech-stack:
  added: []
  patterns: [cli-flags-env-vars-reference, mcp-client-config-json, error-catalog-with-json-examples, symptom-cause-fix-table]

key-files:
  created: [USAGE.md]
  modified: []

key-decisions:
  - "Path-based auth documented as recommended method per D-08"
  - "Bearer header priority over path auth documented per auth.go"
  - "Both hosted and self-hosted URLs provided for each MCP client per D-08"
  - "Streamable HTTP only transport noted per D-09"
  - "All error responses use exact JSON format from source code per D-10"
  - "Troubleshooting table uses symptom-cause-fix format per D-11"

patterns-established:
  - "Error catalog: each error type has JSON example, cause, and fix"
  - "MCP client config: hosted and self-hosted variants for each client"
  - "CLI reference table: flag, env var, default, description"

requirements-completed: [USE-01, USE-02, USE-03, USE-04, USE-05, USE-06]

# Metrics
duration: 5min
completed: 2026-04-20
---

# Phase 7: Installation & Usage Plan 2 Summary

**USAGE.md covering all 6 usage requirements: configuration, auth, MCP client integration, engine discovery, error reference, and troubleshooting**

## Performance

- **Duration:** 5 min
- **Started:** 2026-04-20T10:47:41Z
- **Completed:** 2026-04-20T10:53:11Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- Comprehensive USAGE.md with CLI flags/env vars reference table matching main.go exactly
- Complete API key authentication documentation covering path-based (recommended), header-based, auth-disabled mode, and Bearer header priority
- MCP client integration config for Claude Desktop, Cursor, VS Code Copilot, and Windsurf with both hosted and self-hosted URL patterns
- Engine discovery and search tool documentation with example payloads for basic, specific engine, compact, and location searches
- Full error reference catalog with exact JSON format from auth.go, search.go, and validation.go source code
- Troubleshooting symptom-cause-fix table with debug tips including curl examples and correlation ID documentation

## Task Commits

Each task was committed atomically:

1. **Task 1: Write USAGE.md** - `f06003c` (docs)

**Plan metadata:** pending

## Files Created/Modified
- `USAGE.md` - Complete usage guide (445 lines): configuration, auth, MCP clients, engines, errors, troubleshooting

## Decisions Made
- Path-based auth documented as recommended method per D-08 context decision
- Both hosted and self-hosted URL variants provided for each MCP client
- Streamable HTTP only transport noted (no SSE/WebSocket per D-09)
- Error responses use exact strings from source code for accuracy
- Troubleshooting table uses symptom-cause-fix markdown format per D-11
- Debug tips include curl examples and correlation ID documentation

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- USAGE.md complete, README.md links to it now resolve
- Phase 07 is complete when 07-01 (INSTALL.md) also ships
- Ready for milestone completion once both documentation plans are done

---
*Phase: 07-installation-usage*
*Completed: 2026-04-20*
## Self-Check: PASSED
