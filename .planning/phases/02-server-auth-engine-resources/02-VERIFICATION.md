---
phase: 02-server-auth-engine-resources
verified: 2026-04-16T19:30:00Z
status: passed
score: 11/11 must-haves verified
gaps: []
human_verification:
  - test: "Start the server binary and verify it listens on configured port with startup banner"
    expected: "serpapi-mcp {version} (commit: {commit}, built: {date}) printed, logs engine count, responds to /health"
    why_human: "Requires running server process beyond what unit tests cover"
  - test: "Connect a real MCP client (e.g., Claude Desktop) and verify resource discovery"
    expected: "Client can list resources and read serpapi://engines and serpapi://engines/{engine}"
    why_human: "Requires live MCP protocol interaction with a real client"
  - test: "Verify CORS headers appear correctly in browser DevTools for OPTIONS and GET requests"
    expected: "Access-Control-Allow-Origin, Methods, Headers, Credentials headers present"
    why_human: "Browser DevTools inspection is a visual/manual verification"
---

# Phase 2: Server, Auth & Engine Resources Verification Report

**Phase Goal:** A running Go MCP server accepts authenticated connections and serves engine schemas
**Verified:** 2026-04-16T19:30:00Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | MCP server starts on configured port and responds to Streamable HTTP MCP protocol requests | ✓ VERIFIED | server.go: NewMCPServer creates mcp.Server + mcp.NewStreamableHTTPHandler; Run() starts HTTP server; `go build ./cmd/serpapi-mcp` compiles; 37/37 tests pass |
| 2 | Healthcheck endpoint at /health returns 200 OK | ✓ VERIFIED | server.go:84-93 registers /health handler returning JSON {"status":"healthy"}; TestHealthcheckEndpoint passes |
| 3 | CORS headers are present on responses (allow_origins=*) | ✓ VERIFIED | cors.go:corsMiddleware sets Access-Control-Allow-Origin/Methods/Headers/Credentials; cors_test.go verifies; integration tests confirm CORS on 401 responses and preflight |
| 4 | Server shuts down gracefully on SIGINT/SIGTERM | ✓ VERIFIED | main.go:54 signal.NotifyContext(SIGINT,SIGTERM); server.go:Run() waits for ctx.Done() then http.Server.Shutdown(10s timeout); TestGracefulShutdown passes |
| 5 | Requests without a valid API key are rejected with 401 and JSON error body | ✓ VERIFIED | auth.go:76-86 returns 401 with {"error":"Missing API key..."}; TestMissingAPIKey and TestIntegrationUnauthenticatedMCPReturns401 pass |
| 6 | API key extracted from URL path /{KEY}/mcp and path rewritten to /mcp | ✓ VERIFIED | auth.go:64-72 splits path, extracts key segment, rewrites r.URL.Path; TestPathBasedAuth and TestPathBasedAuthStripsKey pass |
| 7 | API key extracted from Authorization: Bearer {KEY} header | ✓ VERIFIED | auth.go:58-61 checks Authorization header for Bearer format; TestBearerHeaderAuth and TestBearerHeaderAuthWithPrefix pass |
| 8 | Auth middleware composed with MCP handler via http.Handler wrapping | ✓ VERIFIED | server.go:buildHandler chains corsMiddleware(corsCfg, authOrPassthrough(AuthDisabled, mux)); 6 integration tests verify chain behavior |
| 9 | serpapi://engines resource returns the list of all available engine names | ✓ VERIFIED | engines.go:registerEnginesIndex creates JSON resource with count, engines list, resource URIs; TestEnginesIndexResource verifies structure |
| 10 | serpapi://engines/{engine} resource returns per-engine parameter schema | ✓ VERIFIED | engines.go:registerEngineResource creates per-engine resources; TestEngineSchemaResource verifies google_light schema returned correctly |
| 11 | Server fails to start if engine JSON is corrupt or missing | ✓ VERIFIED | engines.go:LoadAndRegister returns errors on read/parse failures; main.go:47-50 os.Exit(1) on error; TestLoadAndRegister_MissingDir and TestLoadAndRegister_CorruptJSON pass |

**Score:** 11/11 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/server/server.go` | NewMCPServer, Run, buildHandler functions (min 50 lines) | ✓ VERIFIED | 162 lines — NewMCPServer, Run, buildHandler, SetEngineCount, healthResponse |
| `internal/server/cors.go` | CORSConfig, NewCORSConfig, corsMiddleware (min 20 lines) | ✓ VERIFIED | 55 lines — CORSConfig struct, NewCORSConfig, corsMiddleware |
| `internal/server/auth.go` | API key auth middleware with path+Bearer (min 40 lines) | ✓ VERIFIED | 92 lines — authMiddleware, APIKeyFromContext, authOrPassthrough, context key type |
| `internal/engines/engines.go` | LoadAndRegister, resource registration (min 80 lines) | ✓ VERIFIED | 165 lines — LoadAndRegister, registerEnginesIndex, registerEngineResource, validation |
| `internal/server/auth_test.go` | Auth tests (min 50 lines) | ✓ VERIFIED | 328 lines — 9 unit tests + 6 integration tests |
| `internal/engines/engines_test.go` | Engine loading tests (min 40 lines) | ✓ VERIFIED | 386 lines — 9 tests including real 107-engine test |
| `cmd/serpapi-mcp/main.go` | Entry point with flags, env vars, engines.LoadAndRegister | ✓ VERIFIED | 95 lines — CLI flags, env vars, signal handling, LoadAndRegister call |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| cmd/serpapi-mcp/main.go | internal/server/server.go | server.NewMCPServer, server.Run | ✓ WIRED | main.go:43 NewMCPServer(cfg, version), main.go:57 mcpServer.Run(ctx) |
| cmd/serpapi-mcp/main.go | internal/engines/engines.go | engines.LoadAndRegister | ✓ WIRED | main.go:46 LoadAndRegister(mcpServer.MCPServer, enginesDir, logger) |
| internal/server/server.go | go-sdk/mcp | mcp.NewServer, mcp.NewStreamableHTTPHandler | ✓ WIRED | server.go:45 mcp.NewServer, server.go:55 mcp.NewStreamableHTTPHandler |
| internal/server/server.go | internal/server/cors.go | corsMiddleware in buildHandler | ✓ WIRED | server.go:99 corsMiddleware(corsCfg, authenticated) |
| internal/server/server.go | internal/server/auth.go | authOrPassthrough in buildHandler | ✓ WIRED | server.go:96 authOrPassthrough(s.Config.AuthDisabled, mux) |
| internal/engines/engines.go | go-sdk/mcp | server.AddResource, mcp.Resource | ✓ WIRED | engines.go:121 srv.AddResource, engines.go:146 srv.AddResource |
| internal/engines/engines.go | engines/*.json | os.ReadDir, os.ReadFile | ✓ WIRED | engines.go:31 os.ReadDir(enginesDir), engines.go:58 os.ReadFile(filePath) |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| engines.go | schemas map | engines/*.json files on disk | ✓ 107 real engine schemas loaded | ✓ FLOWING |
| engines.go | engineNames []string | sorted from schemas map | ✓ used in index resource | ✓ FLOWING |
| auth.go | apiKey string | Authorization header or URL path | ✓ stored in context for downstream | ✓ FLOWING |
| server.go | handler chain | CORS → Auth → mux wiring | ✓ HTTP requests flow through all layers | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Binary compiles | `go build ./cmd/serpapi-mcp` | Success (exit 0) | ✓ PASS |
| Vet passes | `go vet ./...` | Success (exit 0) | ✓ PASS |
| All tests pass | `go test ./internal/... -count=1 -timeout 30s` | 37/37 PASS | ✓ PASS |
| Engine count matches | `ls engines/*.json \| wc -l` | 107 | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| MCP-01 | 02-01 | Streamable HTTP transport using go-sdk StreamableHTTPHandler | ✓ SATISFIED | server.go:55 mcp.NewStreamableHTTPHandler with Stateless:true, JSONResponse:true |
| MCP-02 | 02-01 | Healthcheck endpoint at /health returning 200 OK | ✓ SATISFIED | server.go:84-93 /health handler; TestHealthcheckEndpoint passes |
| MCP-03 | 02-01 | CORS support matching Python server behavior | ✓ SATISFIED | cors.go:corsMiddleware sets all required headers; TestCORSMiddleware* passes |
| MCP-04 | 02-01 | Graceful shutdown on SIGINT/SIGTERM | ✓ SATISFIED | main.go:54 signal.NotifyContext; server.go:143-161 shutdown with 10s timeout; TestGracefulShutdown passes |
| AUTH-01 | 02-02 | API key extraction from URL path /{KEY}/mcp | ✓ SATISFIED | auth.go:64-72 path parsing and rewrite; TestPathBasedAuth passes |
| AUTH-02 | 02-02 | API key extraction from Authorization: Bearer header | ✓ SATISFIED | auth.go:58-61 Bearer header extraction; TestBearerHeaderAuth passes |
| AUTH-03 | 02-02 | Auth middleware composed via http.Handler wrapping | ✓ SATISFIED | server.go:96 authOrPassthrough wraps mux in buildHandler chain |
| ENG-01 | 02-03 | Engine schemas loaded from engines/*.json at startup | ✓ SATISFIED | engines.go:LoadAndRegister reads from enginesDir; TestLoadAndRegister_RealEnginesDir loads all 107 |
| ENG-02 | 02-03 | Engine list resource at serpapi://engines | ✓ SATISFIED | engines.go:registerEnginesIndex; TestEnginesIndexResource verifies JSON structure |
| ENG-03 | 02-03 | Per-engine schema resource at serpapi://engines/{engine} | ✓ SATISFIED | engines.go:registerEngineResource; TestEngineSchemaResource verifies google_light |
| ENG-04 | 02-03 | Startup validation fail-fast on corrupt/missing JSON | ✓ SATISFIED | engines.go:30-75 validates all files; returns errors; main.go:47-50 os.Exit(1) on error |
| ENG-05 | 02-03 | Engine generation remains via build-engines.py | ✓ SATISFIED | build-engines.py still at repo root; Go code only consumes JSON files |

No orphaned requirements found. All 12 requirement IDs from plans are accounted for in REQUIREMENTS.md and verified in the codebase.

### Anti-Patterns Found

No anti-patterns detected. No TODO/FIXME/placeholder comments, no empty implementations, no hardcoded empty data, no console.log-only handlers found in any key files.

### Human Verification Required

1. **Live server startup**
   - **Test:** Run `./serpapi-mcp --engines-dir ./engines` and verify startup message and port
   - **Expected:** Logs startup with port and engine count of 107; /health returns 200
   - **Why human:** Requires running server process beyond unit test scope

2. **MCP client resource discovery**
   - **Test:** Connect an MCP client and list resources, then read serpapi://engines and serpapi://engines/google_light
   - **Expected:** Client sees 108 resources (1 index + 107 engines); per-engine schema returned
   - **Why human:** Requires live MCP protocol interaction with a real client

3. **CORS behavior in browser**
   - **Test:** Send OPTIONS and GET requests from browser DevTools
   - **Expected:** All CORS headers present on both authenticated and unauthenticated responses
   - **Why human:** Browser DevTools inspection is visual/manual

### Gaps Summary

No gaps found. All 11 observable truths verified, all 12 requirements satisfied, all key links wired, all tests passing. The phase goal — "A running Go MCP server accepts authenticated connections and serves engine schemas" — is fully achieved.

---

_Verified: 2026-04-16T19:30:00Z_
_Verifier: the agent (gsd-verifier)_