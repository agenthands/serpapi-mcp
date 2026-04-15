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

### Active

- [ ] Go MCP server with Streamable HTTP transport (markmcclain/mcp-go-sdk or similar)
- [ ] API key authentication middleware (path `/{KEY}/mcp` and `Authorization: Bearer` header)
- [ ] Search tool delegating to SerpApi HTTP API
- [ ] Engine parameter schemas served as MCP resources (`serpapi://engines` and `serpapi://engines/<engine>`)
- [ ] Complete and compact response modes
- [ ] Healthcheck endpoint
- [ ] CORS support
- [ ] Proper MCP-compliant error responses (ToolError, not string prefixes)
- [ ] Input validation: reject invalid engine names, missing required params, invalid mode
- [ ] Startup validation of engine schemas (fail fast on corrupt/missing JSON)
- [ ] Structured logging with request correlation IDs
- [ ] Multi-platform static binary builds (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
- [ ] Legacy Python code moved to `legacy/` directory

### Out of Scope

- WebSocket/SSE transport — Streamable HTTP only
- Multi-tenant API key management — each request carries its own key
- Result caching — SerpApi handles caching
- Rate limiting on the server side — SerpApi handles rate limiting per API key
- Docker/container deployment — shipping static binaries instead
- CloudWatch EMF metrics — Go binary won't use AWS-specific metrics format
- `build-engines.py` code generation — will be replaced with Go-native schema generation or converted

## Context

- **Existing codebase**: Python MCP server (`src/server.py`) using FastMCP on Starlette + uvicorn — being replaced
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
| Go rewrite | Static binaries, no runtime dependencies, type-safe, better performance for concurrent MCP clients | — Pending |
| Multi-platform binaries | Simplifies deployment, no Docker needed, matches Go's cross-compilation strength | — Pending |
| Faithful port + improvements | Keeps API compatibility while fixing known issues (error handling, validation, logging) | — Pending |
| Archive Python to `legacy/` | Preserve reference implementation for comparison during rewrite | — Pending |

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

*Last updated: 2026-04-15 after v1.0 milestone definition*