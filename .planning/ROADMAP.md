# Roadmap: SerpApi MCP Server

## Overview

Documentation milestone covering the full user and contributor journey: from understanding the Go codebase internals (ARCHITECTURE.md), to discovering the project (README.md), to installing and running it (INSTALL.md + USAGE.md). Phases deliver complete documents in dependency order — architecture first (no external dependencies), then README (links forward to install/usage), then combined installation & usage (the operational guides README points to).

## Milestones

- ✅ **v1.0 Go Rewrite MVP** - Phases 1-4 (shipped 2026-04-17)
- 🚧 **v1.1 Documentation** - Phases 5-7 (in progress)

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

<details>
<summary>✅ v1.0 Go Rewrite MVP (Phases 1-4) - SHIPPED 2026-04-17</summary>

### Phase 1: Project Foundation
**Goal**: Go project scaffolded with CI, builds, and legacy Python archived
**Plans**: 2 plans

Plans:
- [x] 01-01: Go module init, standard layout, legacy archive, goreleaser
- [x] 01-02: CI workflow (golangci-lint, go vet, go test), GitHub Actions

### Phase 2: Server, Auth & Engine Resources
**Goal**: Running MCP server with auth, engine resources, and healthcheck
**Plans**: 3 plans

Plans:
- [x] 02-01: Streamable HTTP server, CORS, graceful shutdown, /health
- [x] 02-02: API key auth middleware (path + header), context-based key passing
- [x] 02-03: Engine schema loading, validation, MCP resource registration

### Phase 3: Search, Validation & Observability
**Goal**: Search tool with validation, error handling, and request tracing
**Plans**: 2 plans

Plans:
- [x] 03-01: Search tool, MCP-compliant errors, input validation
- [x] 03-02: Correlation ID middleware, structured logging

### Phase 4: Testing & Release
**Goal**: Full test coverage and release-ready project
**Plans**: 2 plans

Plans:
- [x] 04-01: Unit tests, integration tests, coverage targets
- [x] 04-02: CI hardening, Makefile, goreleaser release

</details>

### 🚧 v1.1 Documentation (In Progress)

**Milestone Goal:** Comprehensive documentation for both users (MCP client integrators) and contributors (codebase developers), covering architecture, testing flows, installation, and usage.

- [x] **Phase 5: Architecture Documentation** - ARCHITECTURE.md with package layout, request flows, subsystem designs, and testing/CI strategy (2026-04-19)
- [ ] **Phase 6: README** - Project overview, quickstart, and navigation links to all detailed docs
- [ ] **Phase 7: Installation & Usage** - INSTALL.md and USAGE.md covering all platforms, configuration, MCP client integration, and troubleshooting

## Phase Details

### Phase 5: Architecture Documentation
**Goal**: Contributors can understand the Go codebase structure, request flows, and system design from a single comprehensive document
**Depends on**: Nothing (continues from v1.0 Phase 4 — existing shipped codebase)
**Requirements**: ARCH-01, ARCH-02, ARCH-03, ARCH-04, ARCH-05, ARCH-06, ARCH-07
**Success Criteria** (what must be TRUE):
  1. Contributor can read ARCHITECTURE.md and identify every package, its purpose, and its dependencies
  2. Contributor can trace an HTTP request from entry to SerpApi response using included diagrams
  3. Contributor can explain engine schema loading, search execution, observability design, and CI/CD pipeline after reading ARCHITECTURE.md
  4. Contributor can understand the test suite structure (unit vs integration, how to run, coverage targets) from ARCHITECTURE.md
**Plans**: 2 plans

Plans:
- [x] 05-01: Package layout, startup & request flow diagrams (ARCH-01, ARCH-02)
- [x] 05-02: Subsystem designs, testing & CI/CD sections (ARCH-03–ARCH-07)

### Phase 6: README
**Goal**: Any visitor can quickly understand what the project does and get started in under 5 minutes
**Depends on**: Nothing (entry point doc; links to INSTALL.md and USAGE.md work whether or not those files exist yet)
**Requirements**: READ-01, READ-02, READ-03
**Success Criteria** (what must be TRUE):
  1. Visitor can read README.md and understand what the server does and who it's for
  2. Visitor can follow the quickstart guide and have a running server in under 5 minutes
  3. Visitor can find links to all detailed documentation files (ARCHITECTURE.md, INSTALL.md, USAGE.md)
**Plans**: 1 plan

Plans:
- [ ] 06-01: Write new README.md (README overview, dual-path quickstart, brief auth & search, doc links, license)

### Phase 7: Installation & Usage
**Goal**: Users can install, configure, run, integrate, and troubleshoot the server using complete operational guides
**Depends on**: Phase 6 (README links to INSTALL.md and USAGE.md)
**Requirements**: INST-01, INST-02, INST-03, USE-01, USE-02, USE-03, USE-04, USE-05, USE-06
**Success Criteria** (what must be TRUE):
  1. User can install pre-built binaries, via go install, or from source by following INSTALL.md
  2. User can configure and run the server with env vars, CLI flags, and auth settings by following USAGE.md
  3. User can connect an MCP client and discover/perform searches using the tool by following USAGE.md
  4. User can interpret error responses and troubleshoot common issues using USAGE.md
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 5 → 6 → 7

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Foundation | v1.0 | 2/2 | Complete | 2026-04-17 |
| 2. Server, Auth & Engine Resources | v1.0 | 3/3 | Complete | 2026-04-17 |
| 3. Search, Validation & Observability | v1.0 | 2/2 | Complete | 2026-04-17 |
| 4. Testing & Release | v1.0 | 2/2 | Complete | 2026-04-17 |
| 5. Architecture Documentation | v1.1 | 2/2 | Complete | 2026-04-19 |
| 6. README | v1.1 | 0/? | Not started | - |
| 7. Installation & Usage | v1.1 | 0/? | Not started | - |