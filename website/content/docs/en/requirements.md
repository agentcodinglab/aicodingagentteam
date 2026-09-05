# Requirements Analysis

## Background

Modern software development increasingly delegates routine coding work to AI assistants. Today, two operating models are common:

1. **Single-assistant mode** — one coding CLI works on a task end-to-end. Simple, but limited: it cannot parallelize review and writing, it forgets past mistakes, and it has no governance boundary.
2. **Manual team mode** — a human team of PM / Architect / Backend / QA / Security / DevOps collaborates. Powerful but slow and expensive for smaller tasks.

AiCodingAgentTeam proposes a third model: a **virtual 9-role software team** in which the Coordinator orchestrates, real AI coding CLIs execute, and a deterministic quality gate enforces correctness — without the project holding any API key.

## Functional Scope

### In scope (MVP and beyond)

- Dispatch user requirements to a 9-role team (PM, Architect, Frontend, Backend, QA, Security, DevOps, plus Reviewer / Writer)
- Communicate between roles over the A2A protocol (InProc for development, Redis Pub/Sub for containerized)
- Ingest project memory and RAG knowledge into each role's context
- Enforce a deterministic quality gate (lint, vet, test)
- Expose four external protocols: gRPC, MCP, ACP, A2A
- TypeScript TUI client for interactive control

### Out of scope (deliberately deferred)

- Holding API keys for any AI provider (delegated to host CLI)
- Uploading project code to a remote service (local-first)
- Building a proprietary LLM (the platform orchestrates external CLIs)

## Non-functional Requirements

| Dimension | Target |
|---|---|
| Determinism | Quality gate is reproducible across runs |
| Latency | A2A message handoff p95 ≤ 500ms |
| Coverage | Go core packages ≥ 90% unit test coverage |
| Locality | All execution happens inside the container; no code leaves |
| Composability | Each role is an independent container, replaceable |

## Acceptance Criteria

The MVP is considered shipped when the user can run `aicodingagentteam run "<natural-language requirement>"` on a small project and receive:

- A passing quality gate (lint + vet + test green)
- A proof-pack artifact zip containing plan + verification log + scorecard
- A delivery summary suitable for human review

See [Architecture](/docs/architecture) for the design that realizes these requirements.