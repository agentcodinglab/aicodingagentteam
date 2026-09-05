# AiCodingAgentTeam

## 🌐 Languages / 语言

- [English](README.md)
- [简体中文](README.zh.md)
- [日本語](README.ja.md)
- [한국어](README.ko.md)
- [Français](README.fr.md)
- [Deutsch](README.de.md)
- [Русский](README.ru.md)
- [Español](README.es.md)
- [Italiano](README.it.md)

> Eine in Go geschriebene KI-Coding-Orchestrierungsplattform. Sie besitzt selbst kein LLM, sondern delegiert an externe KI-Coding-CLIs (Codex und OpenCode als reale Treiber; Claude-Code und DeepSeek-DSH als Stubs), simuliert ein 9-Rollen-Softwareentwicklungsteam, das über das A2A-Protokoll zusammenarbeitet, erzwingt ein deterministisches Quality Gate mit Governance-Audit, wird in Containern ausgeliefert, exponiert vier Protokolle — gRPC, MCP, ACP und A2A — und stellt einen TypeScript-TUI-Client bereit.

[![CI](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml/badge.svg)](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

## Kernfunktionen

- **Trennung von Orchestrierung und Ausführung** — Der Coordinator orchestriert nur und schreibt keinen Code; die Host-CLIs führen nur aus und entscheiden nicht.
- **Container = Rolle** — 9 Teamrollen (PM, Architect, Frontend, Backend, QA, Security, DevOps, …) laufen jeweils in eigenen Containern, unabhängig skalierbar und austauschbar.
- **A2A-Protokoll-Kollaboration** — Strukturierte Agent-Kommunikation mit Unterstützung für InProc (Entwicklung) und Redis Pub/Sub (Container-Modus).
- **Deterministisches Quality Gate** — Hart ausgeführt durch `golangci-lint` + `go vet` + `go test`, ohne Selbstbewertung des Modells.
- **RAG-Wissensbasis + Projektspeicher** — BM25-Suche wird in den Kontext injiziert; der Speicher akkumuliert Fehlerlektionen und historische Lösungen.
- **Vier externe Protokolle** — gRPC (TUI), MCP (externe Werkzeugintegration), ACP (Standard-Agent-Client), A2A (Instanz-Mesh).
- **Lokal zuerst** — Code verlässt den Container nicht; keine API-Schlüssel im Bestand; jegliche Authentifizierung wird an die Host-CLI delegiert.

## Architekturüberblick

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

Vollständige Architekturdiagramme siehe [docs/04-系统架构图.md](docs/04-系统架构图.md).

## Schnellstart

### Einzelner Binärmodus

```bash
# Bauen
go build -o bin/aicodingagentteam ./cmd/aicodingagentteam

# Projekt initialisieren
./bin/aicodingagentteam init

# Gesamte Pipeline ausführen
./bin/aicodingagentteam run "Erstelle eine REST-API" --backend codex

# Schnelle Bearbeitung
./bin/aicodingagentteam quick "Login-Button-Stil korrigieren"

# Qualitätsprüfung
./bin/aicodingagentteam verify

# RAG-Wissensbasis-Demo
./bin/aicodingagentteam knowledge demo

# Coordinator-Dienst starten (gRPC + A2A HTTP)
./bin/aicodingagentteam serve
```

### TUI-Client

```bash
cd tui && npm install && npm run build

# Coordinator starten
cd .. && go run ./cmd/aicodingagentteam serve

# TUI starten (in einem anderen Terminal)
cd tui && node dist/cli.js
# Demo-Modus (kein Coordinator erforderlich)
node dist/cli.js --demo
```

### Container-Deployment

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

## Befehlsreferenz

| Befehl | Beschreibung |
|---|---|
| `init` | Projektkonfiguration initialisieren |
| `run "requirement"` | Gesamte Pipeline ausführen |
| `quick "description"` | Leichte Schnellbearbeitung |
| `verify` | Qualitätsprüfung ausführen |
| `govern [--ci] [path]` | Governance-Scan |
| `report` | Compliance-Bericht ausgeben |
| `serve` | Coordinator starten (gRPC + A2A) |
| `knowledge index [dir]` | Verzeichnis in Wissensbasis indizieren |
| `knowledge search "query"` | Top-5 Wissens-Chunks abrufen |
| `knowledge demo` | End-to-End RAG + Speicher Demo |
| `version` | Version ausgeben |

## Projektstruktur

```
agent_team/
├── cmd/aicodingagentteam/     # Go-Binär-Einstieg
├── internal/
│   ├── a2a/                   # A2A-Protokoll (InProc + Redis)
│   ├── acp/                   # Agent Client Protocol
│   ├── agent/                 # 9-Rollen-Registry
│   ├── audit/                 # Audit-Log
│   ├── config/                # Konfigurationslader
│   ├── coordinator/           # Orchestrierungs-Kern (5-Schicht-Flow)
│   ├── governance/            # Governance-Regeln (113+ Regeln)
│   ├── host/                  # Host-Treiber (codex/opencode/claude/dsh)
│   ├── knowledge/             # RAG-Retrieval-Engine (BM25)
│   ├── mcp/                   # Model Context Protocol
│   ├── memory/                # Projektspeicher (facts/pitfalls/lessons)
│   ├── planner/               # DAG-Planer
│   ├── qualitygate/           # Quality-Gate-Engine
│   ├── router/                # Intent-Router
│   ├── scheduler/             # Rollen-Scheduler + Single-Writer-Lock
│   └── types/                 # Geteilte Typen
├── pkg/
│   ├── api/                   # gRPC-Proto + Server
│   ├── contracts/             # Front-/Back-End-Verträge
│   └── runtime/               # Host-Treiber-Trait
├── tui/                       # TypeScript-TUI-Client
├── deploy/                    # Container-Deployment
├── proto/                     # gRPC-Proto-Definitionen
└── docs/                      # Dokumentation (Anforderungen/Architektur/Pläne/ADR/spec/plan)
```

## Qualitätsbaseline

| Dimension | Schwellenwert | Tatsächlich |
|---|---|---|
| Go-Testabdeckung | ≥ 80 % (Kern ≥ 90 %) | router 100 % · scheduler 96 % · governance 96 % · codex 94,9 % |
| Lint | 0 neue Warnungen | golangci-lint + eslint |
| Build | ≤ 5 Minuten | CI gesamt ~3 Minuten |
| Sicherheit | 0 hochkritische CVEs | govulncheck + npm audit |

## Entwicklung

```bash
make all      # lint + vet + test + build
make test     # Komplette Testsuite
make lint     # golangci-lint
make run      # build + serve
```

## Dokumentation

- [AGENTS.md](AGENTS.md) — Projektstandards (Wurzeldokument)
- [docs/01-需求分析文档.md](docs/01-需求分析文档.md) — Produktanforderungen
- [docs/02-技术架构设计.md](docs/02-技术架构设计.md) — Architekturentwurf
- [docs/03-系统设计与实施规划.md](docs/03-系统设计与实施规划.md) — Implementierungsplan
- [docs/adr/](docs/adr/) — Architecture Decision Records (ADR-0001 ~ 0013)
- [docs/CONSTRAINTS.md](docs/CONSTRAINTS.md) — Qualitäts-Red-Lines

## Lizenz

[MIT](LICENSE)