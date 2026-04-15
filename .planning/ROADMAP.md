# Roadmap: SerpApi MCP Server

## Overview

Harden the existing SerpApi MCP server from a functional-but-fragile state to production-ready reliability. The journey starts with establishing a test suite (which requires refactoring module-level side effects), then locks in type safety and CI enforcement, overhauls error handling to follow MCP protocol, and finally adds input validation and observability so invalid inputs fail fast and request flows are traceable.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Test Suite & Module Refactoring** - Establish test coverage and refactor module-level side effects for test isolation
- [ ] **Phase 2: Type Annotations & CI Hardening** - Add full type annotations and enforce quality gates in CI
- [ ] **Phase 3: Error Handling & MCP Protocol Compliance** - Replace string-based error handling with MCP-compliant exceptions and async-safe patterns
- [ ] **Phase 4: Input Validation & Observability** - Validate inputs against engine schemas and add request traceability

## Phase Details

### Phase 1: Test Suite & Module Refactoring
**Goal**: Server code is testable and covered by a working test suite
**Depends on**: Nothing (first phase)
**Requirements**: TEST-01, TEST-02, TEST-03, TEST-04, TEST-05, TEST-06, TEST-07, TEST-08, ERR-04, CI-02, CI-04
**Success Criteria** (what must be TRUE):
  1. All core tests pass via `pytest` — unit tests for search tool, engine resources, compact mode, and CloudWatch metrics
  2. API key authentication is verified via Starlette TestClient integration tests (including path-based and header-based key extraction)
  3. Healthcheck endpoint responds correctly in integration tests
  4. Engine registration is a callable `register_engines()` function instead of module-level side effects, enabling clean test isolation
  5. CI runs pytest on every push to validate tests stay green
**Plans**: TBD

Plans:
- [ ] 01-01: Refactor module-level side effects and fix mutable defaults
- [ ] 02-01: Write unit tests for search tool and error handling paths
- [ ] 02-02: Write integration tests for auth middleware and healthcheck
- [ ] 02-03: Write unit tests for engine resources, compact mode, and metrics

### Phase 2: Type Annotations & CI Hardening
**Goal**: Codebase passes mypy strict and CI enforces all quality gates
**Depends on**: Phase 1
**Requirements**: TYPE-01, TYPE-02, TYPE-03, CI-01, CI-03
**Success Criteria** (what must be TRUE):
  1. All functions in server.py and build-engines.py have complete type annotations
  2. mypy passes with zero errors in `disallow_untyped_defs` strict mode on the entire codebase
  3. CI pipeline runs ruff (replacing flake8), mypy, and pytest on every PR — all three must pass for merge
**Plans**: TBD

Plans:
- [ ] 04-01: Add type annotations to server.py and build-engines.py
- [ ] 04-02: Replace flake8 with ruff and add mypy + ruff to CI

### Phase 3: Error Handling & MCP Protocol Compliance
**Goal**: Errors follow MCP protocol correctly and async patterns are safe
**Depends on**: Phase 2
**Requirements**: ERR-01, ERR-02, ERR-03, ERR-05, OBS-03
**Success Criteria** (what must be TRUE):
  1. Search tool errors return MCP-compliant `isError=true` responses via `ToolError` instead of string-prefixed error text
  2. Error handling uses proper exception types and status codes instead of string matching on exception messages
  3. SerpApi search calls run asynchronously via `asyncio.to_thread()` without blocking the event loop
  4. Uncaught exceptions are handled by FastMCP's `ErrorHandlingMiddleware` with consistent error responses across all error paths
**Plans**: TBD

Plans:
- [ ] 05-01: Replace string-prefix errors with ToolError exceptions and proper exception type checking
- [ ] 05-02: Wrap blocking serpapi calls in asyncio.to_thread and add ErrorHandlingMiddleware

### Phase 4: Input Validation & Observability
**Goal**: Invalid inputs are caught early and request flows are traceable
**Depends on**: Phase 3
**Requirements**: VAL-01, VAL-02, VAL-03, OBS-01, OBS-02
**Success Criteria** (what must be TRUE):
  1. Search tool rejects invalid engine names and missing required parameters with clear error messages
  2. Mode parameter accepts only "complete" or "compact" at the tool input schema level (not in function body)
  3. Server validates all engine JSON files at startup and fails fast on malformed or missing schemas
  4. Every request includes a correlation ID visible in logs for cross-request traceability
  5. Search tool logs use FastMCP Context (`ctx.info()`/`ctx.debug()`) instead of raw `logger.info(json.dumps(...))` calls
**Plans**: TBD

Plans:
- [ ] 06-01: Add engine schema validation and input parameter validation
- [ ] 06-02: Add correlation IDs and migrated context logging

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Test Suite & Module Refactoring | 0/4 | Not started | - |
| 2. Type Annotations & CI Hardening | 0/2 | Not started | - |
| 3. Error Handling & MCP Protocol Compliance | 0/2 | Not started | - |
| 4. Input Validation & Observability | 0/2 | Not started | - |