# Phase 5: Architecture Documentation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-19
**Phase:** 05-architecture-documentation
**Areas discussed:** Document structure, Diagram format, Level of detail, Testing docs scope

---

## Document structure

| Option | Description | Selected |
|--------|-------------|----------|
| Package-first | One section per package, then cross-cutting sections for flows/auth/observability/testing/CI | ✓ |
| Flow-first | Start with request flow diagram, reference packages inline | |
| Hybrid | Overview + flow + package sections + cross-cuts | |

**User's choice:** Package-first
**Notes:** Package sections first, then cross-cutting sections (request flow, auth, engine loading, search execution, observability, error handling). Testing and CI/CD as final sections.

## Diagram format

| Option | Description | Selected |
|--------|-------------|----------|
| Mermaid only | Renders on GitHub, flowcharts and sequence diagrams | |
| ASCII only | Works everywhere (terminal, editor), harder to maintain | |
| Both Mermaid + ASCII | Maximum compatibility, doubles maintenance effort | ✓ |

**User's choice:** Both Mermaid + ASCII
**Notes:** Four specific flows need diagrams: HTTP request flow, engine loading pipeline, search tool execution path, startup sequence.

## Level of detail

| Option | Description | Selected |
|--------|-------------|----------|
| Package-level overview | Key types, function signatures, dependencies. No line-by-line walkthroughs | ✓ |
| Function-level deep dives | Every exported function with params, returns, behavior | |

**User's choice:** Package-level overview
**Notes:** Include small code snippets showing package wiring (e.g., main.go initialization). No full file reproductions.

## Testing docs scope

| Option | Description | Selected |
|--------|-------------|----------|
| Structure + how to run | Suite structure, unit vs integration, make targets, coverage target | ✓ |
| Full contributor guide | Everything above + how to add tests, mocking patterns, where to put tests | |

**User's choice:** Structure + how to run
**Notes:** Not a full contributor guide — "how to add tests" belongs in CONTRIBUTING.md (deferred to v2).

## Agent's Discretion

- Exact section headings and ordering within the package-first structure
- Which types/functions to highlight per package
- Wording and prose style
- Whether to include a table-of-contents
- Exact code snippet selection for wiring examples
- Whether to include a dependency diagram in addition to the 4 flow diagrams

## Deferred Ideas

None — discussion stayed within phase scope