# Phase 1: Test Suite & Module Refactoring - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-15
**Phase:** 01-test-suite-module-refactoring
**Areas discussed:** Testing approach, Module refactoring, Test organization, CI design

---

## Testing Approach

| Option | Description | Selected |
|--------|-------------|----------|
| Dual-layer (FastMCP Client + Starlette TestClient) | Unit tests via in-memory Client, integration tests via HTTP TestClient | ✓ |
| FastMCP Client only | All tests via in-memory transport, no HTTP | |
| Starlette TestClient only | All tests via HTTP, no in-memory transport | |
| httpx AsyncClient | Full async HTTP testing | |

**User's choice:** Dual-layer (recommended) — auto-selected
**Notes:** Research confirmed `get_http_request()` requires HTTP context for search tool, so TestClient is mandatory for auth tests. FastMCP Client provides faster, simpler unit tests for other features.

---

## Module Refactoring Scope

| Option | Description | Selected |
|--------|-------------|----------|
| Minimal: extract functions + create_app() factory | Register engines as callable function, create_app() factory, keep single file | ✓ |
| Moderate: split into package with modules | Separate auth, metrics, engines, search into modules under src/ | |
| Full: restructure into package with sub-packages | Full package structure with each concern in own module | |

**User's choice:** Minimal refactoring (recommended) — auto-selected
**Notes:** Architecture research confirms flat structure is correct for this scale (one tool, one resource family). Over-engineering into packages adds complexity without value.

---

## Test Organization

| Option | Description | Selected |
|--------|-------------|----------|
| tests/ with conftest.py | Standard pytest project layout, shared fixtures in conftest | ✓ |
| Inline tests in src/ | Tests alongside source files | |
| Hybrid (unit in src/, integration in tests/) | Mixed approach | |

**User's choice:** tests/ with conftest.py (recommended) — auto-selected
**Notes:** Standard Python project layout. conftest.py provides shared fixtures for FastMCP Client, TestClient, and SerpApi mocking.

---

## CI Design

| Option | Description | Selected |
|--------|-------------|----------|
| New PR workflow (pytest + mypy + ruff) | Separate from existing format check | ✓ |
| Extend existing format_check.yml | Add all checks to one workflow | |
| Multiple specialized workflows | Separate workflows for each check | |

**User's choice:** New PR workflow (recommended) — auto-selected
**Notes:** Keeps existing format_check.yml unchanged. New workflow runs all quality gates on PRs.

---

## Agent's Discretion

- Test fixture details (mock setup, number of test cases per feature)
- Assertion style (assert vs pytest.raises patterns)
- conftest.py fixture naming conventions

## Deferred Ideas

None — discussion stayed within phase scope.