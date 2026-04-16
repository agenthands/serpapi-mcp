# Phase 2: Server, Auth & Engine Resources - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-16
**Phase:** 02-server-auth-engine-resources
**Areas discussed:** Auth middleware & key passing, Engine loading strategy, Configuration & flags, Healthcheck & CORS details

---

## Auth middleware & key passing

| Option | Description | Selected |
|--------|-------------|----------|
| Unified auth middleware | One middleware handles both header and path auth. Strips key from path, normalizes into context. | ✓ |
| Split SDK + custom | Use go-sdk RequireBearerToken for header auth, add separate path-stripping middleware. | |
| Header-only auth | Reject path-based auth entirely, only support Bearer header. Breaks backward compatibility. | |

**User's choice:** Unified auth middleware
**Notes:** Matches Python's single-middleware model, simpler handler chain.

**Key passing:**

| Option | Description | Selected |
|--------|-------------|----------|
| Custom context key | Store API key as string in context. Simple, explicit, no SDK coupling. | ✓ |
| Adapt into TokenInfo | Populate go-sdk's TokenInfo struct to use TokenInfoFromContext(). | |

**User's choice:** Custom context key
**Notes:** Decouples from SDK's token model, explicit and simple.

**Auth errors:**

| Option | Description | Selected |
|--------|-------------|----------|
| 401 JSON | Return 401 with JSON body like Python. Consistent with existing clients. | ✓ |
| 401 JSON + WWW-Authenticate | Also add WWW-Authenticate header. More HTTP-standard but unexpected for clients. | |

**User's choice:** 401 JSON
**Notes:** Consistent with Python server's error format.

---

## Engine loading strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Load all into memory | Validate and load all 107 schemas at startup. Fast serving, ~1-2 MB RAM, fail-fast. | ✓ |
| Validate startup, read disk per request | Validate at startup, re-read from disk per request. Less RAM, slower, needs runtime error handling. | |

**User's choice:** Load all into memory
**Notes:** Matches ENG-04 fail-fast requirement naturally, fast per-request serving.

**Resource registration:**

| Option | Description | Selected |
|--------|-------------|----------|
| All 107 as static resources | Register each engine individually via AddResource. Explicit, matches Python. | ✓ |
| Template with dynamic lookup | Use AddResourceTemplate with URI pattern. Fewer registrations, runtime lookup. | |

**User's choice:** All 107 as static resources
**Notes:** Matches Python's per-engine factory pattern exactly.

---

## Configuration & flags

| Option | Description | Selected |
|--------|-------------|----------|
| Env vars + CLI flags | Support both env vars and flags (--host, --port), flags take precedence. | ✓ |
| Env vars only | Only use MCP_HOST, MCP_PORT matching Python exactly. | |

**User's choice:** Env vars + CLI flags
**Notes:** Standard Go practice for server binaries.

**Flag parsing library:**

| Option | Description | Selected |
|--------|-------------|----------|
| stdlib flag package | No external dependency. Keeps go-sdk as only external dep. | ✓ |
| cobra + viper | More polished CLI but adds external dependencies. | |

**User's choice:** stdlib flag package
**Notes:** Keeps dependency surface minimal per Phase 1 D-18.

---

## Healthcheck & CORS details

**Healthcheck path:**

| Option | Description | Selected |
|--------|-------------|----------|
| /health only | Matches MCP-02 requirement. No backward-compatible alias. | ✓ |
| Both /health + /healthcheck | Supports new standard and backward compat. | |

**User's choice:** /health only
**Notes:** Clean, minimal, matches requirement spec.

**CORS policy:**

| Option | Description | Selected |
|--------|-------------|----------|
| Wide-open by default, flag to restrict | allow_origins=["*"] default, --cors-origins flag to tighten. | ✓ |
| Always open, no config | Hard-coded allow_origins=["*"], no escape hatch. | |

**User's choice:** Wide-open by default, flag to restrict
**Notes:** Works out of the box, configurable when needed.

---

## Agent's Discretion

- Middleware chain ordering (auth vs CORS position)
- Engine schema struct representation in Go (map vs typed struct)
- Graceful shutdown timeout value
- Error response JSON format details beyond the `error` key

## Deferred Ideas

None — discussion stayed within phase scope