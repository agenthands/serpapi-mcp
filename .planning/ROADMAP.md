# Roadmap: SerpApi MCP Server — Go Rewrite

## Overview

Rewrite the Python MCP server in Go: scaffold the project, build the MCP server with auth and engine resources, deliver the search tool with validation and observability, then harden with tests and verify the build pipeline. The result is a stateless, static-binary Go server that drops in as a replacement for the Python version with improved error handling and input validation.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3, 4): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Project Foundation** - Go module, layout, legacy archival, CI, and release tooling
- [ ] **Phase 2: Server, Auth & Engine Resources** - Running MCP server with authentication and engine schema discovery
- [ ] **Phase 3: Search, Validation & Observability** - Search tool with validated inputs, proper errors, and structured logging
- [ ] **Phase 4: Testing & Release** - Full test suite and release build verification

## Phase Details

### Phase 1: Project Foundation
**Goal**: The Go project is scaffolded, builds cleanly, and CI runs on every PR
**Depends on**: Nothing (first phase)
**Requirements**: SETUP-01, SETUP-02, SETUP-03, SETUP-04, SETUP-05
**Success Criteria** (what must be TRUE):
  1. `go build ./cmd/serpapi-mcp` succeeds producing a binary
  2. `golangci-lint run && go vet ./...` passes cleanly
  3. Legacy Python code is archived in `legacy/` and out of the Go build path
  4. `goreleaser release --snapshot` produces binaries for 5 platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
**Plans**: 2 plans

Plans:
- [ ] 01-01: Go module init, standard layout, and legacy archival
- [ ] 01-02: CI workflow and goreleaser configuration

### Phase 2: Server, Auth & Engine Resources
**Goal**: A running Go MCP server accepts authenticated connections and serves engine schemas
**Depends on**: Phase 1
**Requirements**: MCP-01, MCP-02, MCP-03, MCP-04, AUTH-01, AUTH-02, AUTH-03, ENG-01, ENG-02, ENG-03, ENG-04, ENG-05
**Success Criteria** (what must be TRUE):
  1. MCP server starts on the configured port and responds to Streamable HTTP MCP protocol requests
  2. Requests without a valid API key (via URL path or Bearer header) are rejected with an auth error
  3. `serpapi://engines` resource returns the list of all available engine names to MCP clients
  4. `serpapi://engines/{engine}` resource returns the parameter schema for a specific engine to MCP clients
  5. Server fails to start if engine JSON is corrupt or missing; `/health` returns 200 OK; CORS headers are present on responses
**Plans**: 3 plans

Plans:
- [ ] 02-01: MCP server with Streamable HTTP transport, healthcheck, CORS, and graceful shutdown
- [ ] 02-02: API key auth middleware (path-based and Bearer header)
- [ ] 02-03: Engine resource loading, schema serving, and startup validation

### Phase 3: Search, Validation & Observability
**Goal**: AI agents can search any SerpApi engine through the MCP tool with validated inputs and proper error handling
**Depends on**: Phase 2
**Requirements**: SRCH-01, SRCH-02, SRCH-03, SRCH-04, SRCH-05, SRCH-06, SRCH-07, VAL-01, VAL-02, VAL-03, OBS-01, OBS-02, OBS-03
**Success Criteria** (what must be TRUE):
  1. Search tool calls SerpApi and returns results in complete or compact mode
  2. SerpApi rate limits (429), auth errors (401/403), and server errors (5xx) return MCP-compliant error responses (IsError: true, not string prefixes)
  3. Invalid engine names, modes, and missing required parameters are rejected with clear error messages
  4. Log entries include correlation IDs for end-to-end request tracing
  5. Server logs a startup confirmation message with port and engine count
**Plans**: 2 plans

Plans:
- [ ] 03-01: Search tool with SerpApi HTTP client, response modes, and proper error handling
- [ ] 03-02: Input validation and structured logging with correlation IDs

### Phase 4: Testing & Release
**Goal**: The server is fully tested and ready for production release
**Depends on**: Phase 3
**Requirements**: TEST-01, TEST-02, TEST-03, TEST-04, TEST-05, TEST-06
**Success Criteria** (what must be TRUE):
  1. Unit tests for the search tool pass, mocking SerpApi HTTP responses (including compact mode)
  2. Unit tests for engine resource loading and schema retrieval pass
  3. Integration tests for auth middleware (path and header key extraction) and healthcheck endpoint pass
  4. Input validation tests confirm invalid engine names, invalid modes, and missing parameters are properly rejected
**Plans**: 2 plans

Plans:
- [ ] 04-01: Unit tests for search tool, compact mode, and engine resources
- [ ] 04-02: Integration tests for auth middleware, healthcheck, and input validation

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Project Foundation | 0/2 | Not started | - |
| 2. Server, Auth & Engine Resources | 0/3 | Not started | - |
| 3. Search, Validation & Observability | 0/2 | Not started | - |
| 4. Testing & Release | 0/2 | Not started | - |