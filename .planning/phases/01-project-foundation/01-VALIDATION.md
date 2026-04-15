---
phase: 1
slug: project-foundation
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-04-15
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none (Go uses file conventions) |
| **Quick run command** | `go build ./cmd/serpapi-mcp` |
| **Full suite command** | `go build ./cmd/serpapi-mcp && go vet ./... && go test ./...` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go build ./cmd/serpapi-mcp`
- **After every plan wave:** Run `go build ./cmd/serpapi-mcp && go vet ./... && go test ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | SETUP-01 | build | `go build ./cmd/serpapi-mcp` | ✅ | ⬜ pending |
| 01-01-01 | 01 | 1 | SETUP-02 | file | `test -f cmd/serpapi-mcp/main.go && test -d internal/server` | ✅ | ⬜ pending |
| 01-01-02 | 01 | 1 | SETUP-03 | file | `test -d legacy/src && ! test -d src` | ✅ | ⬜ pending |
| 01-02-01 | 02 | 2 | SETUP-04 | file | `test -f .github/workflows/ci.yml` | ✅ | ⬜ pending |
| 01-02-02 | 02 | 2 | SETUP-05 | build | `goreleaser release --snapshot --clean` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/server/server.go` — stub package (package server, no exported funcs yet)
- [ ] `internal/search/search.go` — stub package (package search, no exported funcs yet)
- [ ] `internal/engines/engines.go` — stub package (package engines, no exported funcs yet)
- [ ] `cmd/serpapi-mcp/main.go` — minimal main() that compiles

*These are created as part of Plan 01 Task 1, not a separate Wave 0.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| goreleaser produces 5 platform binaries | SETUP-05 | Requires goreleaser CLI installed | `goreleaser release --snapshot --clean` and check dist/ for 5 platform archives |

*All other phase behaviors have automated verification.*

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending