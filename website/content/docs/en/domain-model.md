# Domain Model

The project's ubiquitous language. Use these terms consistently in code, docs, and discussion.

## Core Entities

### Coordinator
The orchestrator. Routes the user's request, builds a plan, dispatches roles, and enforces the quality gate. **Never writes code.**

### Role
A specialized function in the team. Current roles: PM, Architect, Frontend, Backend, QA, Security, DevOps. Each role is a separate container.

### Task
A unit of work passed from one role to another. Has a typed input, expected output, and verdict.

### Verdict
The outcome of a task. One of: `pass`, `fail`, `park`. The quality gate inspects every verdict; a `pass` is required for the overall run to succeed.

### Reviewer
A role whose task is to critique another role's output. Runs in parallel with other reviewers.

### Writer
A role whose task is to mutate project state. Runs serially under a single-writer lock to prevent concurrent edits.

## Communication

### A2A (Agent-to-Agent)
The only channel between roles. Schema defined in `docs/spec/a2a-protocol.md`. Validated by CI.

### Bus
The transport. `InProc` for development, `Redis` (Pub/Sub) for containerized. The schema is identical; only the transport differs.

## Knowledge Layer

### Knowledge
BM25-retrieved context. Built by `aicodingagentteam knowledge index <dir>`. Retrieved at every role invocation.

### Memory
Three kinds, persisted across runs:
- **facts** — verified truths about the project
- **pitfalls** — known failure modes to avoid
- **lessons** — learned from past runs, surfaced as hints

## Quality Layer

### Quality Gate
A fixed set of machine checks. The gate is not a model; it is `golangci-lint`, `go vet`, `go test`. The gate is allowed to fail; the run then fails.

### Proof Pack
A zip artifact containing `plan.json`, `verify.jsonl`, `scorecard.md`, `delivery-summary.md`. Produced by every successful run.

## External Surface

### gRPC
The TUI's transport.

### MCP (Model Context Protocol)
External tool integration.

### ACP (Agent Client Protocol)
Standard Agent client endpoint.

### A2A HTTP
Cross-instance mesh endpoint.

## Principles

1. **Orchestration and execution are separate.** A role cannot decide and execute at the same time.
2. **Containers are roles.** No role shares memory with another.
3. **Protocols are contracts.** Changing the wire format requires an ADR.
4. **Artifacts are communication.** Roles exchange files, not free-form chat.
5. **Determinism wins.** Quality gates are machine-run, not model-graded.
6. **Local first.** Code does not leave the container by default.
7. **No keys held.** Authentication is delegated to the underlying host CLI.
8. **Scale adaptive.** Small tasks get a lightweight flow; big projects get the full team.