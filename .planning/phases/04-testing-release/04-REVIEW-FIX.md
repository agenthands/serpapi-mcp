---
status: all_fixed
phase: 04-testing-release
iteration: 1
findings_in_scope: 4
fixed: 4
skipped: 0
scope: critical_warning
---

# Code Review Fix Report: Phase 04 (testing-release)

**Scope:** critical_warning (High + Critical) | **Iteration:** 1 | **Date:** 2026-04-17

## Findings Fixed

| ID | Severity | Finding | Fix |
|----|----------|---------|-----|
| H-01 | High | Global mutable state in `search` package — data race on `serpapiBaseURL` | Replaced `var serpapiBaseURL` with `var serpapiBaseURLResolver = func() string`. Tests swap the function variable instead of writing a shared global, eliminating the race. |
| H-02 | High | Per-request `http.Client` defeats connection pooling | Added package-level `var serpAPIClient = &http.Client{}`. `callSearchTool` now uses `http.NewRequestWithContext(ctx, ...)` with `context.WithTimeout(30s)`. Shared transport reuses keep-alive connections. |
| M-01 | Medium | Hand-rolled `contains()` reimplements `strings.Contains` | Replaced with `strings.Contains` in `containsJSON` helper. Deleted the manual character-by-character function. |
| L-05 | Low | Orphaned comment fragment in `auth.go` | Restored full doc comment: `// authMiddleware validates the API key from the Authorization header or URL path` |

## Files Changed

| File | Change |
|------|--------|
| `internal/search/search.go:22-26` | `serpapiBaseURL` → `serpapiBaseURLResolver` (func var); added `serpAPIClient` shared client; `context.WithTimeout` per-request |
| `internal/search/search_test.go:58-80` | Test helpers swap resolver + restore via `t.Cleanup` |
| `internal/search/search_test.go:340-342,401-403,468-470` | Standalone tests use resolver swap pattern |
| `internal/search/validation_test.go:139-141` | Validation test uses resolver swap pattern |
| `internal/server/auth.go:45` | Restored complete `authMiddleware` doc comment |
| `internal/server/server_test.go:104-111` | Replaced `contains()` with `strings.Contains` |

## Skipped Findings

None. All in-scope findings fixed.

## Verification

- `go test -race ./...` — all packages pass
- `go vet ./...` — clean