# System Design and Implementation Plan

## Phased Rollout

The project is delivered in five named phases. Each phase has explicit acceptance criteria and ships a working artifact.

### Phase A — Foundation

**Goal** Establish the Go project skeleton, CI pipeline, and quality baseline.

- Go module + `internal/`, `cmd/`, `pkg/` layout
- CI: build + vet + test + lint
- 80% unit test coverage on `internal/`

### Phase B — Coordinator

**Goal** Stand up the 5-layer flow end-to-end without external AI.

- Router → Planner → Scheduler → Reviewer → Writer → Quality Gate
- InProc A2A bus
- Stub host drivers for all four backends
- Single-writer lock enforced by tests

### Phase C — Real Drivers

**Goal** Wire Codex and OpenCode as real drivers; keep Claude-Code and DSH as stubs.

- `pkg/runtime/runtime.go::Runtime` trait
- Codex driver with stdin/stdout protocol
- OpenCode driver with HTTP protocol
- Coverage ≥ 90% on `host/codex` and `host/opencode`

### Phase D — RAG + Memory

**Goal** Inject knowledge and memory into every role's context.

- BM25 indexer over `docs/` and `examples/`
- Memory store with facts / pitfalls / lessons
- Director wiring so every node sees relevant context

### Phase E — External Protocols

**Goal** Expose the platform over gRPC, MCP, ACP, and A2A.

- gRPC server for the TUI
- MCP server for external tools
- ACP endpoint for standard Agent clients
- A2A HTTP endpoint for cross-instance mesh

## Quality Gates per Phase

| Phase | Gate |
|---|---|
| A | `make all` green; coverage ≥ 80% |
| B | End-to-end test of full pipeline passes |
| C | Real driver produces identical artifact shape to stub |
| D | RAG demo shows context retrieved and used |
| E | All four protocols respond to a smoke ping |

## Risks and Mitigations

| Risk | Mitigation |
|---|---|
| Real driver version drift | Pin driver CLI version; surface mismatch in quality gate |
| BM25 noise degrading context | Tunable score threshold per role |
| A2A message loss in container mode | Idempotent retries + dead-letter queue |
| Quality gate flaky tests | `-race` + fixed time zones + no network in unit tests |

See [Quality Constraints](/docs/quality-constraints) for the project's red-line thresholds.