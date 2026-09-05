# AiCodingAgentTeam

## 🌐 Languages / 语言

[English](README.md) · [简体中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Русский](README.ru.md) · [Español](README.es.md) · [Italiano](README.it.md)
> A Golang-based AI coding orchestration platform. It owns no LLM itself; instead it dispatches external AI coding CLIs (Codex & OpenCode as real drivers; Claude-Code & DeepSeek-DSH as stubs), simulates a 9-role software-development team that collaborates over the A2A protocol, enforces a deterministic quality gate with governance auditing, ships in containers, and exposes four protocols externally — gRPC, MCP, ACP, and A2A — alongside a TypeScript TUI client.

[![CI](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml/badge.svg)](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Docs](https://img.shields.io/badge/docs-aicodingagentteam.dev-blue)](https://agentcodinglab.github.io/aicodingagentteam/)

## Core Features

- **Orchestration vs execution separation** — The Coordinator only orchestrates and never writes code; host CLIs only execute and never decide.
- **Container-as-role** — 9 team roles (PM, Architect, Frontend, Backend, QA, Security, DevOps, …), each in its own container, independently scalable and replaceable.
- **A2A protocol collaboration** — Structured inter-agent communication, supporting InProc (development) and Redis Pub/Sub (containerized).
- **Deterministic quality gate** — Hard-executed by `golangci-lint` + `go vet` + `go test`; never relies on model self-evaluation.
- **RAG knowledge base + project memory** — BM25 retrieval injected into context; the memory store accumulates failure lessons and historical solutions.
- **Four external protocols** — gRPC (TUI communication), MCP (external tool integration), ACP (standard Agent client), A2A (cross-instance mesh).
- **Local-first** — Code never leaves the container; no API keys are held; all auth is delegated to the underlying host CLI.

## Architecture Overview

```
User → TypeScript TUI (Ink) → gRPC → Coordinator
                                       ├── Router (intent routing)
                                       ├── Planner (DAG plan)
                                       ├── Scheduler (role dispatch)
                                       │    ├── A2A Bus → Reviewer Agents (parallel)
                                       │    └── Single-Writer Lock → Writer Agents (serial)
                                       ├── Knowledge (BM25 RAG retrieval)
                                       ├── Memory (facts/pitfalls/lessons)
                                       └── Quality Gate (deterministic checks)
                                            ↓
Host CLIs: Codex (real) | OpenCode (real) | Claude (stub) | DSH (stub)
```

See [docs/04-系统架构图.md](docs/04-系统架构图.md) for the full architecture diagrams.

## Quick Start

### Single-binary mode

```bash
# Build
go build -o bin/aicodingagentteam ./cmd/aicodingagentteam

# Initialize a project
./bin/aicodingagentteam init

# Run a full pipeline
./bin/aicodingagentteam run "Build a REST API" --backend codex

# Quick edit
./bin/aicodingagentteam quick "Fix the login button style"

# Quality verification
./bin/aicodingagentteam verify

# RAG knowledge base demo
./bin/aicodingagentteam knowledge demo

# Start the coordinator service (gRPC + A2A HTTP)
./bin/aicodingagentteam serve
```

### TUI client

```bash
cd tui && npm install && npm run build

# Start the coordinator
cd .. && go run ./cmd/aicodingagentteam serve

# Launch the TUI (in another terminal)
cd tui && node dist/cli.js
# Demo mode (no coordinator required)
node dist/cli.js --demo
```

### Containerized deployment

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

## Command Reference

| Command | Description |
|---|---|
| `init` | Initialize project configuration |
| `run "requirement"` | Run the full pipeline |
| `quick "description"` | Lightweight quick edit |
| `verify` | Execute quality verification |
| `govern [--ci] [path]` | Governance scan |
| `report` | Emit compliance report |
| `serve` | Start coordinator (gRPC + A2A) |
| `knowledge index [dir]` | Index a directory into the knowledge base |
| `knowledge search "query"` | Retrieve top-5 knowledge chunks |
| `knowledge demo` | End-to-end RAG + memory demo |
| `version` | Print version |

## Project Structure

```
agent_team/
├── cmd/aicodingagentteam/     # Go binary entry
├── internal/
│   ├── a2a/                   # A2A protocol (InProc + Redis)
│   ├── acp/                   # Agent Client Protocol
│   ├── agent/                 # 9-role registry
│   ├── audit/                 # Audit log
│   ├── config/                # Configuration loader
│   ├── coordinator/           # Orchestration core (5-layer flow)
│   ├── governance/            # Governance rules (113+ rules)
│   ├── host/                  # Host drivers (codex/opencode/claude/dsh)
│   ├── knowledge/             # RAG retrieval engine (BM25)
│   ├── mcp/                   # Model Context Protocol
│   ├── memory/                # Project memory (facts/pitfalls/lessons)
│   ├── planner/               # DAG plan builder
│   ├── qualitygate/           # Quality-gate engine
│   ├── router/                # Intent router
│   ├── scheduler/             # Role scheduler + single-writer lock
│   └── types/                 # Shared types
├── pkg/
│   ├── api/                   # gRPC proto + server
│   ├── contracts/             # Front/back-end contracts
│   └── runtime/               # Host driver trait
├── tui/                       # TypeScript TUI client
├── deploy/                    # Containerized deployment
├── proto/                     # gRPC proto definitions
└── docs/                      # Documentation (requirements/architecture/plans/ADR/spec/plan)
```

## Quality Baseline

| Dimension | Threshold | Actual |
|---|---|---|
| Go test coverage | ≥ 80% (core ≥ 90%) | router 100% · scheduler 96% · governance 96% · codex 94.9% |
| Lint | 0 new warnings | golangci-lint + eslint |
| Build | ≤ 5 minutes | CI full flow ~3 minutes |
| Security | 0 high-severity CVEs | govulncheck + npm audit |

## Development

```bash
make all      # lint + vet + test + build
make test     # Full test suite
make lint     # golangci-lint
make run      # build + serve
```

## Documentation

- [AGENTS.md](AGENTS.md) — Project standards (root doc)
- [docs/01-需求分析文档.md](docs/01-需求分析文档.md) — Product requirements
- [docs/02-技术架构设计.md](docs/02-技术架构设计.md) — Architecture design
- [docs/03-系统设计与实施规划.md](docs/03-系统设计与实施规划.md) — Implementation plan
- [docs/adr/](docs/adr/) — Architecture Decision Records (ADR-0001 ~ 0013)
- [docs/CONSTRAINTS.md](docs/CONSTRAINTS.md) — Quality red lines

## License

[MIT](LICENSE)