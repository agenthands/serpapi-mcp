# Phase 5: Architecture Documentation - Context

**Gathered:** 2026-04-19
**Status:** Ready for planning

<domain>
## Phase Boundary

Contributors can understand the Go codebase structure, request flows, and system design from a single comprehensive ARCHITECTURE.md. This phase delivers: package layout documentation, request flow diagrams, subsystem designs (auth, engine loading, search, observability), and testing/CI strategy — all in ARCHITECTURE.md.

</domain>

<decisions>
## Implementation Decisions

### Document structure
- **D-01:** Package-first organization — one section per package (server, engines, search, middleware, main), each with purpose, key types, functions, and dependencies
- **D-02:** Cross-cutting sections after package sections — request flow, auth design, engine loading pipeline, search execution, observability, error handling
- **D-03:** Testing and CI/CD as final cross-cutting sections at the end of the document

### Diagram format
- **D-04:** Both Mermaid and ASCII diagrams for every flow — Mermaid renders on GitHub, ASCII works everywhere (terminal, editors, plain text)
- **D-05:** Four flows require diagrams: HTTP request flow (CORS → correlation → auth → MCP → SerpApi → response), engine loading pipeline, search tool execution path, startup sequence

### Level of detail
- **D-06:** Package-level overview — exported types with one-line descriptions, key function signatures, dependency list. No line-by-line walkthroughs.
- **D-07:** Include small code snippets showing how packages wire together (e.g., main.go initialization flow). No full file reproductions.
- **D-08:** Contributor reads ARCHITECTURE.md for the map, then jumps to source files for implementation details

### Testing docs scope
- **D-09:** Describe test suite structure: which packages have tests, unit vs integration distinction, how to run (make test, make test-race, make cover), current coverage target
- **D-10:** NOT a full contributor guide — no "how to add new tests" patterns or mocking guides. That belongs in CONTRIBUTING.md (deferred to v2).

### Agent's Discretion
- Exact section headings and ordering within the package-first structure
- Which types/functions to highlight per package (only the key ones, not exhaustive)
- Wording and prose style
- Whether to include a table-of-contents and what it links to
- Exact code snippet selection for wiring examples
- Whether to include a dependency diagram in addition to the 4 flow diagrams

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Context
- `.planning/PROJECT.md` — Vision, constraints, key decisions, current state (shipped v1.0, 4 phases, 10 plans, 3,500 lines of Go, 86% test coverage)
- `.planning/REQUIREMENTS.md` — ARCH-01 through ARCH-07 requirements for this phase
- `.planning/ROADMAP.md` — Phase 5 goal, success criteria, and plan outline

### Reference Implementation
- `cmd/serpapi-mcp/main.go` — Entry point, flag parsing, initialization flow, env var helpers
- `internal/server/server.go` — MCPServer struct, Config, buildHandler chain, Run method
- `internal/server/auth.go` — Auth middleware, path stripping, context key, APIKeyFromContext
- `internal/server/cors.go` — CORSConfig, corsMiddleware, preflight handling
- `internal/middleware/correlation.go` — Correlation ID generation (crypto/rand), middleware, CorrelationIDFromContext
- `internal/engines/engines.go` — LoadAndRegister, engine schema validation, resource registration, EngineNames, RequiredParams
- `internal/search/search.go` — RegisterSearchTool, callSearchTool, toolError, compact mode, SerpApi HTTP calls
- `internal/search/validation.go` — ValidateEngine, ValidateMode, ValidateRequiredParams

### Build & CI
- `.github/workflows/ci.yml` — CI pipeline: golangci-lint, go vet, go test -race
- `Makefile` — Make targets: test, test-race, cover, lint, vet
- `.goreleaser.yml` — Release config: 5 platforms, tag-triggered, SHA256 checksums
- `go.mod` — Module path, Go version, sole external dep (go-sdk v1.5.0)

### Test Files
- `cmd/serpapi-mcp/main_test.go` — Main package tests
- `internal/engines/engines_test.go` — Engine loading and validation tests
- `internal/middleware/correlation_test.go` — Correlation middleware tests
- `internal/search/search_test.go` — Search tool unit tests
- `internal/search/validation_test.go` — Validation logic tests
- `internal/server/auth_test.go` — Auth middleware tests
- `internal/server/cors_test.go` — CORS middleware tests
- `internal/server/server_test.go` — Server integration tests

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- Complete Go codebase with 4 internal packages (server, engines, search, middleware) + cmd entry point — all source files are the primary documentation source
- 7 test files (2,498 total lines) covering unit and integration tests — existing test structure can be documented directly
- Makefile with standard targets — already provides the "how to run tests" content
- CI workflow with 3 clear steps — simple pipeline to document
- goreleaser config with 5 platforms — distribution strategy to document

### Established Patterns
- Handler chain ordering: CORS → correlation → auth → mux — established in Phases 2-3, must be documented accurately
- Engine loading: all 107 schemas at startup, fail-fast validation, individual MCP resource registration — established in Phase 2
- Search execution: validate → construct URL → HTTP call → status handling → compact mode → response — established in Phase 3
- Error format: flat JSON `{"error": "<code>", "message": "<desc>"}` with IsError=true — established in Phase 3
- Observability: 32-char hex correlation IDs from crypto/rand, slog structured logging — established in Phase 3

### Integration Points
- ARCHITECTURE.md will be a new top-level file at repo root
- README.md (Phase 6) will link to ARCHITECTURE.md
- INSTALL.md and USAGE.md (Phase 7) may reference architecture concepts

</code_context>

<specifics>
## Specific Ideas

- Both Mermaid and for each diagram — contributor choice of how to consume
- Four specific flow diagrams needed: HTTP request flow, engine loading pipeline, search tool execution, startup sequence
- Package-level depth is the right scope — contributors can read the actual source for implementation details
- Testing section should cover structure and how-to-run, not a contributor guide for writing tests

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 05-architecture-documentation*
*Context gathered: 2026-04-19*