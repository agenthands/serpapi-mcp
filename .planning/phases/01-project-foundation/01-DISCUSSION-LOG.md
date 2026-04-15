# Phase 1: Project Foundation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-15
**Phase:** 1-Project Foundation
**Areas discussed:** Legacy archival scope, CI workflow design, goreleaser config details, Go module setup

---

## Legacy Archival Scope

| Option | Description | Selected |
|--------|-------------|----------|
| Everything Python to legacy/ | Move src/, pyproject.toml, uv.lock, .python-version, Dockerfile, smithery.yaml, copilot/ to legacy/. Keep build-engines.py at root and engines/ at root. server.json stays (registry metadata). | ✓ |
| Only Python source to legacy/ | Move src/ and build-engines.py to legacy/, leave deployment configs at root | |
| Just src/server.py to legacy/ | Move src/ only, gradual transition | |

**User's choice:** Everything Python to legacy/
**Notes:** Clean break preferred. `build-engines.py` stays at root because CI needs it. `server.json` stays — it's MCP registry metadata, not Python-specific.

## CI Workflow Design

| Option | Description | Selected |
|--------|-------------|----------|
| Default golangci-lint | Most common linters enabled, can tighten later | ✓ |
| Maximum strictness | Enable all golangci-lint linters from the start | |
| Relaxed / minimal | Minimal linter set (errcheck, gosimple, govet, ineffassign, staticcheck) | |

| Option | Description | Selected |
|--------|-------------|----------|
| PRs only | CI runs on pull requests to main | ✓ |
| PRs + push to main | Extra safety net but doubles CI minutes | |

| Option | Description | Selected |
|--------|-------------|----------|
| Single version (1.23) | Test on Go 1.23.x only; goreleaser handles cross-platform | ✓ |
| Matrix (1.22 + 1.23) | Future-proofs but doubles CI time | |

**User's choice:** Default golangci-lint, PRs only, single Go 1.23.x
**Notes:** Simple CI preferred. goreleaser cross-compilation means single Go version in CI is sufficient.

## Goreleaser Config Details

| Option | Description | Selected |
|--------|-------------|----------|
| Standard Go naming | serpapi-mcp-{version}-{os}-{arch}.tar.gz | ✓ |
| Short naming | serpapi-mcp-{os}-{arch} without version | |

| Option | Description | Selected |
|--------|-------------|----------|
| tar.gz + .zip | .tar.gz for Linux/macOS, .zip for Windows | ✓ |
| tar.gz only | Simpler but Windows users expect .zip | |

| Option | Description | Selected |
|--------|-------------|----------|
| Tag-triggered releases | Actual releases on git tag push; snapshots in CI | ✓ |
| Auto-release on main push | Ships every merge to main | |

| Option | Description | Selected |
|--------|-------------|----------|
| Yes, include checksums | SHA256 checksums.txt alongside binaries | ✓ |
| No checksums | Simpler but can't verify integrity | |

**User's choice:** Standard naming, tar.gz + .zip, tag-triggered, SHA256 checksums included
**Notes:** All industry-standard goreleaser defaults selected.

## Go Module Setup

| Option | Description | Selected |
|--------|-------------|----------|
| github.com/agenthands/serpapi-mcp | Matches GitHub repo, standard Go convention | ✓ |
| Short name only | Works locally but breaks go install from remote | |

| Option | Description | Selected |
|--------|-------------|----------|
| Go 1.23+ | Latest stable, has all needed features | ✓ |
| Go 1.22+ | Broader compat but misses 1.23 niceties | |

| Option | Description | Selected |
|--------|-------------|----------|
| Domain packages | internal/server, internal/search, internal/engines | ✓ |
| Flat internal/ | Fewer directories but gets unwieldy | |

**User's choice:** github.com/agenthands/serpapi-mcp module path, Go 1.23+, domain-based internal packages
**Notes:** Domain packages mirror the logical structure of the Python server (auth, search, engine resources).

## the agent's Discretion

- Exact `.golangci.yml` configuration within default settings
- goreleaser `.goreleaser.yml` internals (hooks, env vars, builds section)
- `.gitignore` updates — replace Python entries with Go-appropriate entries
- Whether to add a `Makefile` or `Taskfile.yml` for common commands

## Deferred Ideas

None — discussion stayed within phase scope