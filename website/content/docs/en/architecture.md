# Technical Architecture Design

## Goals

The architecture must satisfy five goals simultaneously:

1. **Separation of concerns** — orchestration never writes code, execution never decides
2. **Container-as-role** — every role is an independent, replaceable unit
3. **Deterministic verification** — quality gate runs as a hard machine check
4. **Local-first** — code never leaves the container, no API keys held
5. **Protocol-as-contract** — A2A, ACP, MCP, gRPC are stable interfaces

## Five-Layer Flow

```
Router → Planner → Scheduler → [Reviewer (parallel) | Writer (serial)] → Quality Gate
                       │
                       ├── Knowledge (BM25 RAG)
                       └── Memory (facts / pitfalls / lessons)
```

| Layer | Responsibility | Implementation |
|---|---|---|
| Router | Classify the user request and pick a workflow | Embedding-based intent classifier |
| Planner | Build a DAG of writer / reviewer nodes | Topological sort on role dependencies |
| Scheduler | Dispatch reviewers in parallel, writers under a single-writer lock | Go scheduler with mutex-guarded writer queue |
| Knowledge | Inject BM25-retrieved context into each node | Local index built at `knowledge index` time |
| Memory | Persist facts / pitfalls / lessons across runs | Append-only JSON store under `~/.aicodingagentteam/memory/` |
| Quality Gate | Run lint, vet, test; emit structured failures | Hard-coded commands, machine-readable output |

## A2A Protocol

A2A is the only allowed inter-role communication channel. Its contract is defined in `docs/spec/a2a-protocol.md` and validated by JSON schema in CI:

- **AgentCard** — role identity, capabilities, supported operations
- **Task** — unit of work passed between roles
- **Result** — output, including success/failure verdict
- **ProgressEvent** — incremental status during long-running tasks

The transport switches between InProc (development) and Redis Pub/Sub (containerized) without changing the message schema.

## Host Drivers

The platform never executes LLM calls itself. Instead, each role delegates to a host CLI:

| Backend | Driver | Status |
|---|---|---|
| Codex | Real CLI invocation | Real driver |
| OpenCode | Real CLI invocation | Real driver |
| Claude-Code | Stub | Stub (interface only) |
| DeepSeek-DSH | Stub | Stub (interface only) |

All drivers implement the `pkg/runtime/runtime.go::Runtime` trait; no driver can fake capabilities it does not have.

## Quality Gate

The quality gate is intentionally not negotiable:

- `golangci-lint run ./...` must pass with 0 new warnings
- `go vet ./...` must pass
- `go test ./... -race` must pass
- Coverage must not regress below the project's baseline

If a model self-evaluates a failure as success, the gate treats the run as failed.

## Container Topology

Each role runs in its own container. They communicate only over A2A. Scaling one role does not affect the others. Replacing a role's implementation does not require rebuilding the Coordinator.

See [Implementation Plan](/docs/implementation-plan) for the staged rollout.