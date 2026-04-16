# SerpApi MCP Server

## What This Is

A Go-based Model Context Protocol (MCP) server exposing SerpApi's multi-engine search to AI agents. Runs as a streamable HTTP service with API key auth (path or header), provides a single `search` tool with per-engine parameter schemas as MCP resources, and ships as static binaries for multiple platforms.

## Core Value

AI agents can search any SerpApi-supported engine through a single, authenticated MCP endpoint with structured parameter discovery and proper MCP-compliant error handling.

## Requirements

### Validated

(Moving from Python validation — these capabilities carry forward)
- ✓ MCP server with Streamable HTTP transport — existing design
- ✓ API key authentication (path-based and header-based Bearer) — existing design
- ✓ Single `search` tool routing to all SerpApi engines — existing design
- ✓ Engine parameter schemas served as MCP resources — existing design
- ✓ Complete and compact response modes — existing design
- ✓ Healthcheck endpoint — existing design
- ✓ Engine schema generation from SerpApi playground — existing design

### Validated in Phase 02: Server, Auth & Engine Resources

- ✓ Go MCP server with Streamable HTTP transport (modelcontextprotocol/go-sdk) — serves /mcp and /health endpoints
- ✓ API key authentication middleware (path `/{KEY}/mcp` and `Authorization: Bearer` header) — CORS → Auth → mux handler chain
- ✓ Engine parameter schemas served as MCP resources (`serpapi://engines` and `serpapi://engines/<engine>`) — 107 engines loaded at startup
- ✓ Healthcheck endpoint — /health returns JSON status
- ✓ CORS support — configurable origins, handles preflight
- ✓ Startup validation of engine schemas (fail fast on corrupt/missing JSON) — LoadAndRegister validates all schemas before serving

### Active

- [ ] Search tool delegating to SerpApi HTTP API
- [ ] Complete and compact response modes
- [ ] Proper MCP-compliant error responses (ToolError, not string prefixes)
- [ ] Input validation: reject invalid engine names, missing required params, invalid mode
- [ ] Structured logging with request correlation IDs

### Validated in Phase 01: Project Foundation

- ✓ Multi-platform static binary builds (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) — goreleaser configured, 5 platform targets
- ✓ Legacy Python code moved to `legacy/` directory — archived, out of Go build path
- ✓ Go module initialized with standard layout — `modelcontextprotocol/go-sdk` is sole external dependency
- ✓ CI workflow runs on PR to main — golangci-lint, go vet, go test

### Out of Scope

- WebSocket/SSE transport — Streamable HTTP only
- Multi-tenant API key management — each request carries its own key
- Result caching — SerpApi handles caching
- Rate limiting on the server side — SerpApi handles rate limiting per API key
- Docker/container deployment — shipping static binaries instead
- CloudWatch EMF metrics — Go binary won't use AWS-specific metrics format
- `build-engines.py` code generation — will be replaced with Go-native schema generation or converted

## Context

- **Existing codebase**: Python MCP server archived in `legacy/src/server.py` — reference for Go rewrite
- **Engine schemas**: 100+ engines as JSON in `engines/` — will be consumed by Go server
- **Previous roadmap**: 4-phase Python hardening roadmap (tests → types → errors → validation) is superseded by this Go rewrite
- **Deployment shift**: Moving from AWS Copilot + Docker to multi-platform static binaries
- **Compatibility**: Must maintain API compatibility with existing MCP clients (path auth, header auth, same resource URIs)

## Constraints

- **Language**: Go (rewriting from Python)
- **Transport**: Streamable HTTP only (MCP spec)
- **Auth model**: API key in URL path or Authorization header
- **Engine schemas**: `engines/*.json` consumed at runtime; generation approach TBD (Go-native or converted)
- **Distribution**: Static binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- **No Docker**: No container images — just binaries
- **API compatibility**: Must work with existing MCP clients expecting `/{KEY}/mcp` path and `serpapi://engines` resource URIs

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go rewrite | Static binaries, no runtime dependencies, type-safe, better performance for concurrent MCP clients | ✓ Phase 01: Go module initialized, builds cleanly |
| Multi-platform binaries | Simplifies deployment, no Docker needed, matches Go's cross-compilation strength | ✓ Phase 01: goreleaser configured for 5 platforms |
| Faithful port + improvements | Keeps API compatibility while fixing known issues (error handling, validation, logging) | ✓ Phase 02: Auth, engine resources, healthcheck, CORS all match Python server behavior |
| Archive Python to `legacy/` | Preserve reference implementation for comparison during rewrite | ✓ Phase 01: All Python files in legacy/ |

## Current Milestone: v1.0 Go Rewrite

**Goal:** Rewrite the SerpApi MCP server in Go with improved error handling, validation, and multi-platform binary distribution.

**Target features:**
- Faithful port of all existing capabilities (search tool, engine resources, auth, healthcheck, CORS, compact/complete modes)
- Improved error handling (proper MCP error responses, not string prefixes)
- Input validation (engine names, parameters, mode)
- Structured logging with request correlation IDs
- Startup validation of engine schemas
- Multi-platform static binary releases (no Docker)
- Legacy Python code archived in `legacy/`

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---

*Last updated: 2026-04-16 after Phase 02 completion*