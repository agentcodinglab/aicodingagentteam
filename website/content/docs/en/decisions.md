# Architecture Decision Records

This page provides a chronological overview of all architectural decisions in the AiCodingAgentTeam project.

## Decision Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| ADR-0001 | Go over Rust for orchestration engine | Accepted | 2026-09-02 |
| ADR-0002 | Redis as A2A message bus | Accepted | 2026-09-02 |
| ADR-0003 | Container-per-role isolation | Accepted | 2026-09-02 |
| ADR-0004 | Single-writer model for artifacts | Accepted | 2026-09-02 |
| ADR-0005 | No API key custody in Coordinator | Accepted | 2026-09-02 |
| ADR-0006 | Fail-open governance policy | Accepted | 2026-09-02 |
| ADR-0007 | Host CLI feasibility (superseded by ADR-0021) | Superseded | 2026-09-04 |
| ADR-0008 | A2A bus infrastructure | Accepted | 2026-09-02 |
| ADR-0009 | Quality gate execution model | Accepted | 2026-09-02 |
| ADR-0010 | Volume concurrency and single-writer | Accepted | 2026-09-02 |
| ADR-0011 | Config merge defaults and env override | Accepted | 2026-09-04 |
| ADR-0012 | Quality gate details propagation | Accepted | 2026-09-04 |
| ADR-0013 | RAG memory director wiring | Accepted | 2026-09-05 |
| ADR-0014 | ACP/MCP real implementation | Accepted | 2026-09-05 |
| ADR-0015 | E2E proof and security | Accepted | 2026-09-05 |
| ADR-0016 | Direction A: governance | Accepted | 2026-09-05 |
| ADR-0017 | Direction D: RAG demo | Accepted | 2026-09-05 |
| ADR-0018 | Direction C: real host e2e | Accepted | 2026-09-06 |
| ADR-0019 | P4 streaming stdout | Accepted | 2026-09-06 |
| ADR-0020 | P6 ACP session/newTask | Accepted | 2026-09-06 |
| ADR-0021 | P5 OpenCode serve HTTP | Accepted | 2026-09-06 |
| ADR-0022 | Coverage gap and roadmap | Accepted | 2026-09-06 |
| ADR-0023 | Plugin discovery and init wizard | Accepted | 2026-09-08 |
| ADR-0024 | Quality gate 10 checks | Accepted | 2026-09-09 |

## Key Decisions

### No Key Custody (ADR-0005)
The Coordinator never holds API keys. Authentication is delegated entirely to the underlying CLI. This ensures the orchestration layer remains secure by design.

### Fail-Open Governance (ADR-0006)
When the governance engine panics, the system fails open (passes) rather than blocking delivery. This prevents quality gate infrastructure from becoming a single point of failure.

### Container-per-Role (ADR-0003)
Each role Agent runs in an independent container, enabling independent scaling and replacement. Role logic never cross-calls between containers directly.

### CheckFunc Extension (ADR-0024)
Quality gate checks support custom validation functions via `CheckFunc`, enabling file-based checks, governance scans, and contract cross-validation alongside CLI commands.