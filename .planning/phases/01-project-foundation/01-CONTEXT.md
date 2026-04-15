# Phase 1: Project Foundation - Context

**Gathered:** 2026-04-15
**Status:** Ready for planning

<domain>
## Phase Boundary

The Go project is scaffolded, builds cleanly, and CI runs on every PR. This phase delivers: Go module initialization with standard layout, legacy Python code archival, golangci-lint CI on PRs, and goreleaser multi-platform binary build verification.

</domain>

<decisions>
## Implementation Decisions

### Legacy Archival Scope
- **D-01:** All Python-specific files move to `legacy/` in one clean break: `src/`, `pyproject.toml`, `uv.lock`, `.python-version`, `Dockerfile`, `smithery.yaml`, `copilot/` directory
- **D-02:** `build-engines.py` stays at repo root — it's needed in CI to regenerate `engines/*.json` and is not Go code, so it lives alongside (not inside) the Go project
- **D-03:** `server.json` stays at repo root — it's MCP registry metadata (not Python-specific), referencing the hosted endpoint
- **D-04:** `engines/` directory stays at repo root — consumed by both the Python legacy server and the Go server
- **D-05:** `.venv/` is gitignored and doesn't need archival; the `.gitignore` should be updated for Go artifacts instead of Python-only entries

### CI Workflow Design
- **D-06:** golangci-lint with default settings — most common linters enabled, can tighten later
- **D-07:** CI triggers on pull requests to main only (not on push to main)
- **D-08:** Single Go version in CI: 1.23.x — goreleaser cross-compiles for all target platforms regardless
- **D-09:** CI steps: golangci-lint run, go vet ./..., go test ./... (even if no tests exist yet, the step should be present and passing)

### Goreleaser Configuration
- **D-10:** Standard Go naming convention: `serpapi-mcp-{version}-{os}-{arch}` (e.g., `serpapi-mcp-1.0.0-linux-amd64.tar.gz`)
- **D-11:** Archive formats: `.tar.gz` for Linux/macOS, `.zip` for Windows
- **D-12:** Tag-triggered releases — actual releases only on git tag push; CI runs `goreleaser release --snapshot` for build verification
- **D-13:** SHA256 checksums file included in releases
- **D-14:** Five target platforms: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64

### Go Module Setup
- **D-15:** Module path: `github.com/agenthands/serpapi-mcp` — matches GitHub repo, standard Go convention
- **D-16:** Minimum Go version: 1.23 in go.mod (go directive)
- **D-17:** Domain-based internal packages: `internal/server`, `internal/search`, `internal/engines` — groups by domain, mirrors Python server's logical structure
- **D-18:** Only external dependency: `modelcontextprotocol/go-sdk` (per SETUP-01)

### the agent's Discretion
- Exact `.golangci.yml` configuration within default settings — agent picks which default linters to enable
- goreleaser `.goreleaser.yml` internals (hooks, env vars, builds section) — standard config is fine
- `.gitignore` updates — replace Python entries with Go-appropriate entries, keep what's still relevant
- Whether to add a `Makefile` or `Taskfile.yml` for common commands — agent decides based on convention

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Context
- `.planning/PROJECT.md` — Vision, constraints, key decisions for the Go rewrite
- `.planning/REQUIREMENTS.md` — SETUP-01 through SETUP-05 requirements for this phase
- `.planning/ROADMAP.md` — Phase 1 goal, success criteria, and plan outline

### Legacy Reference
- `src/server.py` — Current Python MCP server (being archived, but needed as reference for Go rewrite)
- `build-engines.py` — Python engine schema generator (stays at root, needed in CI)
- `pyproject.toml` — Python project config (being archived, contains dependency list for reference)
- `Dockerfile` — Python Docker build (being archived, documents previous deployment approach)
- `copilot/manifest.yml` — AWS Copilot service manifest (being archived, documents previous deployment)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `engines/*.json` (107 engine schemas): Consumed at runtime by both Python and Go servers. Format is stable — Go server will read the same JSON files.
- `build-engines.py`: Remains at root for CI; produces `engines/*.json` from SerpApi playground scrape.

### Established Patterns
- Python server uses FastMCP + Starlette + uvicorn (Streamable HTTP transport). Go version will use `modelcontextprotocol/go-sdk` with StreamableHTTPHandler — same protocol, different SDK.
- Auth pattern: API key from URL path `/{KEY}/mcp` or `Authorization: Bearer` header. Must maintain this contract.
- Engine resource URIs: `serpapi://engines` and `serpapi://engines/{engine}`. Must maintain these resource paths.

### Integration Points
- Go server entry: `cmd/serpapi-mcp/main.go`
- Internal packages: `internal/server`, `internal/search`, `internal/engines`
- Engine data: `engines/` directory at repo root (consumed at runtime)
- CI: Replace `.github/workflows/format_check.yml` (Python) and `.github/workflows/deploy.yml` (AWS Copilot) with Go CI workflow

</code_context>

<specifics>
## Specific Ideas

- Clean break from Python — everything Python-specific moves to `legacy/` at once, no gradual migration
- `server.json` stays because it's registry metadata pointing at `mcp.serpapi.com`, not a Python artifact
- CI should be simple: lint → vet → test on PR, goreleaser snapshot for build verification. No deployment workflow in this phase.
- Goreleaser uses standard goreleaser patterns — no custom hooks or unusual config needed

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-project-foundation*
*Context gathered: 2026-04-15*