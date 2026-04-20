---
phase: 07-installation-usage
verified: 2026-04-20T12:00:00Z
status: passed
score: 10/10 must-haves verified
re_verification: false
---

# Phase 7: Installation & Usage Verification Report

**Phase Goal:** Users can install, configure, run, integrate, and troubleshoot the server using complete operational guides
**Verified:** 2026-04-20
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User can download a pre-built binary for their platform and run it | ✓ VERIFIED | INSTALL.md lines 16–71: platform-specific download commands for all 5 targets (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) with curl/PowerShell, tar/zip, and chmod for each |
| 2 | User can install via go install and verify the version | ✓ VERIFIED | INSTALL.md lines 73–105: `go install github.com/agenthands/serpapi-mcp@latest`, PATH setup, `serpapi-mcp --version`, upgrade with version tag |
| 3 | User can build from source with go build or goreleaser | ✓ VERIFIED | INSTALL.md lines 107–167: `go build` command with ldflags for version injection, goreleaser snapshot and release builds, verification and upgrade steps |
| 4 | Each install method has platform-specific instructions for all 5 supported platforms | ✓ VERIFIED | Binary download: 5 explicit subsections with per-platform commands; go install: cross-platform (Go handles this); build from source: goreleaser covers all 5 platforms per `.goreleaser.yml` (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) |
| 5 | User can start the server with CLI flags or env vars by reading USAGE.md | ✓ VERIFIED | USAGE.md lines 6–58: CLI flags table (all 5 flags with env vars and defaults), starting examples with defaults/CLI flags/env vars, version check, .env file note |
| 6 | User can configure API key auth (path or header) by reading USAGE.md | ✓ VERIFIED | USAGE.md lines 60–98: Path-based auth (recommended) with hosted/self-hosted URLs, header-based auth with Bearer token, auth priority order (header > path), auth-disabled mode, /health exemption |
| 7 | User can connect an MCP client using provided config JSON by reading USAGE.md | ✓ VERIFIED | USAGE.md lines 118–226: Config JSON for Claude Desktop, Cursor, VS Code Copilot, and Windsurf — each with hosted and self-hosted URL variants; transport note about Streamable HTTP only |
| 8 | User can discover engines and parameters via MCP resources by reading USAGE.md | ✓ VERIFIED | USAGE.md lines 228–300: Engine index resource `serpapi://engines` with JSON example, per-engine resource `serpapi://engines/<engine>`, search tool examples (basic, specific engine, compact, with location), compact mode field removal list |
| 9 | User can look up any error response and understand its cause by reading USAGE.md | ✓ VERIFIED | USAGE.md lines 302–395: Complete error catalog with exact JSON format — Missing API key, missing_api_key, invalid_engine, invalid_mode, missing_params, rate_limited, invalid_api_key, forbidden, search_error — each with cause and fix |
| 10 | User can troubleshoot common issues using the symptom-cause-fix table | ✓ VERIFIED | USAGE.md lines 397–441: 9-row troubleshooting table (connection refused, 401, invalid key, engine not found, missing params, rate limit, engine schemas missing, empty results, CORS), debug tips with curl examples, correlation ID documentation |

**Score:** 10/10 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `INSTALL.md` | Complete installation guide (min 80 lines, contains "go install") | ✓ VERIFIED | 176 lines, contains "go install", covers all 3 methods with all 5 platforms |
| `USAGE.md` | Complete usage guide (min 150 lines, contains "serpapi://engines") | ✓ VERIFIED | 446 lines, contains "serpapi://engines", covers all 6 usage requirements |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| INSTALL.md | .goreleaser.yml | Platform targets and archive naming | ✓ WIRED | `linux-amd64`, `darwin-arm64` patterns match goreleaser `name_template`; `.tar.gz`/`.zip` format overrides match |
| INSTALL.md | go.mod | Go version requirement and module path | ✓ WIRED | `go 1.25+` and `github.com/agenthands/serpapi-mcp` match go.mod |
| INSTALL.md | cmd/serpapi-mcp/main.go | Version flag and CLI entry point | ✓ WIRED | `--version` flag and `serpapi-mcp` binary name match main.go |
| USAGE.md | cmd/serpapi-mcp/main.go | CLI flags and env var reference | ✓ WIRED | All 5 CLI flags (`--host`, `--port`, `--cors-origins`, `--auth-disabled`, `--engines-dir`) and 5 env vars (`MCP_HOST`, `MCP_PORT`, `MCP_CORS_ORIGINS`, `MCP_AUTH_DISABLED`, `ENGINES_DIR`) with exact defaults match main.go |
| USAGE.md | internal/server/auth.go | Auth methods and error messages | ✓ WIRED | Bearer header priority, path-based `/{KEY}/mcp` pattern, exact error message `"Missing API key. Use path format /{API_KEY}/mcp or Authorization: Bearer {API_KEY} header"` all match |
| USAGE.md | internal/search/search.go | Search tool description and error codes | ✓ WIRED | Default engine `google_light`, default mode `complete`, error codes (search_error, missing_api_key, invalid_engine, rate_limited, invalid_api_key, forbidden), compact mode field removal list all match |
| USAGE.md | internal/search/validation.go | Validation error messages | ✓ WIRED | Exact error strings for `invalid_engine`, `invalid_mode`, `missing_params` match validation.go |
| USAGE.md | README.md | Consistent quickstart snippets and links | ✓ WIRED | USAGE.md next-steps link to README.md; README.md links to USAGE.md for "detailed configuration options", "authentication details", and "full parameter reference" |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|-------------------|--------|
| N/A — documentation phase | N/A | N/A | N/A | SKIPPED |

*This is a documentation phase. Data-flow trace is not applicable — there are no components rendering dynamic data.*

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| SKIP | N/A | N/A | SKIPPED (documentation phase — no runnable entry points produced) |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| INST-01 | 07-01 | User can install pre-built binaries for all 5 supported platforms | ✓ SATISFIED | INSTALL.md lines 16–71: download commands for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64 |
| INST-02 | 07-01 | User can install via `go install` from source | ✓ SATISFIED | INSTALL.md lines 73–105: go install command, PATH setup, version check, upgrade instructions |
| INST-03 | 07-01 | User can build from source with goreleaser or `go build` | ✓ SATISFIED | INSTALL.md lines 107–167: go build with ldflags, goreleaser snapshot/release builds |
| USE-01 | 07-02 | User can configure and run the server (env vars, CLI flags, port/host binding) | ✓ SATISFIED | USAGE.md lines 6–58: CLI flags table, starting examples, version check, env vars |
| USE-02 | 07-02 | User can configure API key authentication (path-based and header-based) | ✓ SATISFIED | USAGE.md lines 60–98: path auth, header auth, priority order, auth-disabled mode, /health exemption |
| USE-03 | 07-02 | User can integrate the server with an MCP client (connection URL, transport config) | ✓ SATISFIED | USAGE.md lines 118–226: Claude Desktop, Cursor, VS Code Copilot, Windsurf configs with hosted+ self-hosted URLs |
| USE-04 | 07-02 | User can discover available engines and their parameters via MCP resources | ✓ SATISFIED | USAGE.md lines 228–300: serpapi://engines index, per-engine resources, search examples |
| USE-05 | 07-02 | User can interpret error responses and understand common failure modes | ✓ SATISFIED | USAGE.md lines 302–395: complete error catalog with JSON examples, causes, and fixes |
| USE-06 | 07-02 | User can troubleshoot common issues (auth failures, missing engines, connection errors) | ✓ SATISFIED | USAGE.md lines 397–441: 9-row symptom-cause-fix table, curl debug examples, correlation IDs |

No orphaned requirements found — all 9 requirement IDs are covered by plans and verified in artifacts.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none) | — | — | — | — |

No TODO, FIXME, PLACEHOLDER, stub, empty implementation, or hardcoded empty data patterns found in INSTALL.md or USAGE.md.

### Human Verification Required

### 1. Cross-document link resolution

**Test:** Click all markdown links in README.md, INSTALL.md, and USAGE.md
**Expected:** All links resolve to existing sections/files
**Why human:** Link resolution in rendered markdown requires visual browser verification

### 2. MCP client config JSON validity

**Test:** Copy each MCP client config JSON block from USAGE.md into the respective client's config file
**Expected:** Each client accepts the config and connects to the server
**Why human:** JSON schema correctness for each client requires running the client application — cannot verify format programmatically without each client's schema validator

### 3. Installation instructions end-to-end

**Test:** Follow binary download, go install, and build-from-source instructions on a clean machine
**Expected:** A working `serpapi-mcp --version` output in all three cases
**Why human:** Requires actually running the commands on each supported platform

### Gaps Summary

No gaps found. All 10 observable truths are verified, both artifacts (INSTALL.md and USAGE.md) are substantive and well-wired to the codebase, all 9 requirement IDs (INST-01 through INST-03, USE-01 through USE-06) are satisfied, and no anti-patterns were detected.

INSTALL.md provides complete, platform-specific installation instructions for all 3 methods across all 5 supported platforms, with accurate goreleaser naming, Go version requirements, and version verification commands that match the source code.

USAGE.md provides comprehensive operational documentation with CLI flags/env vars exactly matching main.go, authentication methods matching auth.go (including Bearer header priority), MCP client configs for 4 clients with both hosted and self-hosted URLs, engine discovery matching the serpapi://engines resource URIs, a complete error catalog with JSON examples matching source code, and a 9-row troubleshooting table with curl debug examples.

---

_Verified: 2026-04-20_
_Verifier: the agent (gsd-verifier)_