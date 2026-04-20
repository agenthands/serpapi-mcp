---
status: testing
phase: 07-installation-usage
source: [07-01-SUMMARY.md, 07-02-SUMMARY.md]
started: 2026-04-20T13:00:00.000Z
updated: 2026-04-20T13:03:00.000Z
---

## Current Test

number: 4
name: Install Prerequisites and Version Info
expected: |
  INSTALL.md lists prerequisites (Go 1.25+ for go install/build, SerpApi key), version verification command, and upgrade instructions per method
awaiting: user response

## Tests

### 1. Binary Download Instructions
expected: INSTALL.md has copy-pasteable download commands for all 5 platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) with curl/tar or PowerShell syntax
result: pass

### 2. Go Install Method
expected: INSTALL.md documents `go install github.com/agenthands/serpapi-mcp@latest`, notes GOPATH/bin, and shows version check with `serpapi-mcp --version`
result: issue
reported: "go install github.com/agenthands/serpapi-mcp@latest fails — module found but does not contain package. The correct command should be `go install github.com/agenthands/serpapi-mcp/cmd/serpapi-mcp@latest` since main package is under cmd/serpapi-mcp"
severity: blocker

### 3. Build from Source
expected: INSTALL.md covers both `go build` and `goreleaser build --snapshot` methods with ldflags for versioned builds
result: issue
reported: "goreleaser v2 rejects .goreleaser.yml version 0 config with 'only version: 2 configuration files are supported, yours is version 0'. The goreleaser section should note the config version requirement or the .goreleaser.yml needs updating to version: 2 format."
severity: major

### 4. Install Prerequisites and Version Info
expected: INSTALL.md lists prerequisites (Go 1.25+ for go install/build, SerpApi key), version verification command, and upgrade instructions per method
result: [pending]

### 5. CLI Flags Reference
expected: USAGE.md has a table with all 5 CLI flags (--host, --port, --cors-origins, --auth-disabled, --engines-dir), their env var equivalents (MCP_HOST, MCP_PORT, MCP_CORS_ORIGINS, MCP_AUTH_DISABLED, ENGINES_DIR), and correct defaults
result: [pending]

### 6. API Key Authentication
expected: USAGE.md documents path-based auth /{KEY}/mcp (recommended), Bearer header auth, auth-disabled mode for testing, and notes Bearer header takes priority over path
result: [pending]

### 7. MCP Client Integration Configs
expected: USAGE.md provides working JSON config snippets for Claude Desktop, Cursor, VS Code Copilot, and Windsurf with both hosted and self-hosted URL patterns
result: [pending]

### 8. Engine Discovery and Search
expected: USAGE.md documents serpapi://engines resource index, per-engine resources, search tool with example payloads (basic, specific engine, compact mode, location)
result: [pending]

### 9. Error Reference Catalog
expected: USAGE.md lists all error codes (invalid_engine, invalid_mode, missing_params, missing_api_key, rate_limited, invalid_api_key, forbidden, search_error, 401 auth) with exact JSON format from source code
result: [pending]

### 10. Troubleshooting Table
expected: USAGE.md has symptom-cause-fix table covering connection refused, auth errors, invalid key, engine not found, missing params, rate limiting, CORS issues, and debug tips with curl examples
result: [pending]

### 11. Cross-Document Links
expected: README.md links to INSTALL.md and USAGE.md; INSTALL.md links to USAGE.md; USAGE.md links to INSTALL.md, README.md, and ARCHITECTURE.md — all links resolve to existing files
result: [pending]

## Summary

total: 11
passed: 1
issues: 2
pending: 8
skipped: 0
blocked: 0

## Gaps

- truth: "go install github.com/agenthands/serpapi-mcp@latest successfully installs the serpapi-mcp binary"
  status: failed
  reason: "User reported: go install github.com/agenthands/serpapi-mcp@latest fails — module found but does not contain package. The correct command needs /cmd/serpapi-mcp suffix since main package is under cmd/serpapi-mcp"
  severity: blocker
  test: 2
  root_cause: ""
  artifacts: []
  missing: []
- truth: "goreleaser build --snapshot --clean works as documented in INSTALL.md"
  status: failed
  reason: "User reported: goreleaser v2 rejects .goreleaser.yml version 0 config — 'only version: 2 configuration files are supported, yours is version: 0'. Config needs updating or docs need a note."
  severity: major
  test: 3
  root_cause: ""
  artifacts: []
  missing: []