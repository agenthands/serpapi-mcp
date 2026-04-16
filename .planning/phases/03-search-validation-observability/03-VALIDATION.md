---
phase: 03
slug: search-validation-observability
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-16
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none (standard go test) |
| **Quick run command** | `go test ./internal/search/... ./internal/engines/... -count=1 -timeout=30s` |
| **Full suite command** | `go test ./... -count=1 -timeout=60s` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/search/... -count=1 -timeout=30s`
- **After every plan wave:** Run `go test ./... -count=1 -timeout=60s`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | SRCH-01,02,03,04 | unit | `go test ./internal/search/... -run TestSearch -count=1` | ❌ W0 | ⬜ pending |
| 03-01-02 | 01 | 1 | SRCH-05,06,07 | unit | `go test ./internal/search/... -run TestSearchError -count=1` | ❌ W0 | ⬜ pending |
| 03-02-01 | 02 | 2 | VAL-01,02,03 | unit | `go test ./internal/search/... -run TestValidation -count=1` | ❌ W0 | ⬜ pending |
| 03-02-02 | 02 | 2 | OBS-01,02 | unit | `go test ./internal/middleware/... -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/search/search_test.go` — stubs for SRCH-01 through SRCH-07
- [ ] `internal/search/validation_test.go` — stubs for VAL-01 through VAL-03
- [ ] `internal/middleware/correlation_test.go` — stubs for OBS-01, OBS-02

*If none: "Existing infrastructure covers all phase requirements."*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Full end-to-end search with real SerpApi key | SRCH-01 | Requires valid API key and network | Run server with key, call search tool via MCP client, verify result returned |
| Startup log with engine count | OBS-03 | Already covered in Phase 2 | Server log shows "engines_loaded" count |

*All other phase behaviors have automated verification via mocked HTTP responses.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending