# Phase 3: Search, Validation & Observability - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-16
**Phase:** 03-search-validation-observability
**Areas discussed:** Error responses

---

## Error responses

| Option | Description | Selected |
|--------|-------------|----------|
| IsError + JSON detail | Return IsError:true in CallToolResult with structured JSON error body | ✓ |
| IsError + plain text message | Return IsError:true with human-readable text message | |
| Match Python format | Return "Error: ..." string prefix (not MCP-compliant) | |

**User's choice:** IsError + JSON detail
**Notes:** User confirmed recommended option — MCP-compliant errors with structured JSON body

### Error JSON detail level

| Option | Description | Selected |
|--------|-------------|----------|
| Flat: error + message | Same flat structure for all errors: {"error": "...", "message": "..."} | ✓ |
| Enriched: add status + engine | Include HTTP status code and request-specific fields | |
| Minimal: error code only | Error code only, client maps to messages | |

**User's choice:** Flat: error + message
**Notes:** Consistent and simple — client always parses the same shape

### SerpApi HTTP error mapping granularity

| Option | Description | Selected |
|--------|-------------|----------|
| Match Python + generic | 429, 401, 403, plus generic catch-all | ✓ |
| Granular HTTP status codes | Per-status-code differentiation (429, 401, 403, 5xx, timeout, network) | |

**User's choice:** Match Python + generic
**Notes:** Three specific cases cover real scenarios; catch-all handles everything else

---

## Areas not discussed (selected but deferred by user)

- Compact mode — carried Python's 5-field list as default (D-08)
- Validation strictness — requirements clear from VAL-01/02/03 (D-09/10/11)
- Correlation IDs — standard Go/slog patterns apply (D-12)

---

## Agent's Discretion

- Exact HTTP client timeout values for SerpApi calls
- Correlation ID generation method (UUID, random hex, etc.)
- How correlation IDs propagate through request context
- Compact mode field removal implementation (iterate-and-delete vs. rebuild-without)
- Whether validation errors also use flat error JSON format or a different shape
- Error response for missing API key at tool level (fallback if auth middleware misses)

## Deferred Ideas

None — discussion stayed within phase scope