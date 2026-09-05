# AiCodingAgentTeam

## 🌐 Languages / 语言

[English](README.md) · [简体中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Русский](README.ru.md) · [Español](README.es.md) · [Italiano](README.it.md)
> Una piattaforma di orchestrazione per il coding con IA scritta in Go. Non ospita un LLM proprio; delega il lavoro a CLI di coding IA esterne (Codex e OpenCode come driver reali; Claude-Code e DeepSeek-DSH come stub), simula un team di sviluppo software a 9 ruoli che collabora tramite il protocollo A2A, applica una porta di qualità deterministica con audit di governance, è distribuita in container ed espone quattro protocolli — gRPC, MCP, ACP e A2A — oltre a un client TUI in TypeScript.

[![CI](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml/badge.svg)](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

## Caratteristiche principali

- **Separazione tra orchestrazione ed esecuzione** — Il Coordinator si limita a orchestrare e non scrive codice; le CLI host eseguono soltanto e non decidono.
- **Container = ruolo** — 9 ruoli di team (PM, Architect, Frontend, Backend, QA, Security, DevOps, …), ciascuno in un proprio container, scalabile e sostituibile in modo indipendente.
- **Collaborazione tramite A2A** — Comunicazione strutturata tra agent; supporto InProc (sviluppo) e Redis Pub/Sub (modalità containerizzata).
- **Porta di qualità deterministica** — Eseguita rigidamente da `golangci-lint` + `go vet` + `go test`, senza dipendere dall’autovalutazione del modello.
- **Base di conoscenza RAG + memoria di progetto** — Ricerca BM25 iniettata nel contesto; la memoria accumula lezioni sugli errori e soluzioni storiche.
- **Quattro protocolli esterni** — gRPC (TUI), MCP (integrazione strumenti esterni), ACP (client Agent standard), A2A (mesh tra istanze).
- **Local-first** — Il codice non esce dal container; non vengono conservate chiavi API; l’autenticazione è interamente delegata alla CLI host.

## Panoramica dell’architettura

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

I diagrammi completi sono in [docs/04-系统架构图.md](docs/04-系统架构图.md).

## Avvio rapido

### Modalità binario singolo

```bash
# Compilazione
go build -o bin/aicodingagentteam ./cmd/aicodingagentteam

# Inizializzazione del progetto
./bin/aicodingagentteam init

# Esecuzione dell’intera pipeline
./bin/aicodingagentteam run "Costruisci una REST API" --backend codex

# Modifica rapida
./bin/aicodingagentteam quick "Correggi lo stile del pulsante di login"

# Verifica di qualità
./bin/aicodingagentteam verify

# Demo della base di conoscenza RAG
./bin/aicodingagentteam knowledge demo

# Avvio del servizio Coordinator (gRPC + A2A HTTP)
./bin/aicodingagentteam serve
```

### Client TUI

```bash
cd tui && npm install && npm run build

# Avvia il Coordinator
cd .. && go run ./cmd/aicodingagentteam serve

# Avvia la TUI (in un altro terminale)
cd tui && node dist/cli.js
# Modalità demo (senza Coordinator)
node dist/cli.js --demo
```

### Deployment containerizzato

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

## Riferimento comandi

| Comando | Descrizione |
|---|---|
| `init` | Inizializza la configurazione del progetto |
| `run "requirement"` | Esegue l’intera pipeline |
| `quick "description"` | Modifica rapida leggera |
| `verify` | Esegue la verifica di qualità |
| `govern [--ci] [path]` | Scansione di governance |
| `report` | Genera il rapporto di conformità |
| `serve` | Avvia il Coordinator (gRPC + A2A) |
| `knowledge index [dir]` | Indicizza una directory nella base di conoscenza |
| `knowledge search "query"` | Recupera i top-5 chunk di conoscenza |
| `knowledge demo` | Demo end-to-end RAG + memoria |
| `version` | Stampa la versione |

## Struttura del progetto

```
agent_team/
├── cmd/aicodingagentteam/     # Ingresso del binario Go
├── internal/
│   ├── a2a/                   # Protocollo A2A (InProc + Redis)
│   ├── acp/                   # Agent Client Protocol
│   ├── agent/                 # Registro dei 9 ruoli
│   ├── audit/                 # Log di audit
│   ├── config/                # Caricatore di configurazione
│   ├── coordinator/           # Cuore di orchestrazione (flusso a 5 strati)
│   ├── governance/            # Regole di governance (113+ regole)
│   ├── host/                  # Driver host (codex/opencode/claude/dsh)
│   ├── knowledge/             # Motore di retrieval RAG (BM25)
│   ├── mcp/                   # Model Context Protocol
│   ├── memory/                # Memoria di progetto (facts/pitfalls/lessons)
│   ├── planner/               # Costruttore del piano DAG
│   ├── qualitygate/           # Motore della porta di qualità
│   ├── router/                # Router delle intenzioni
│   ├── scheduler/             # Scheduler dei ruoli + lock single-writer
│   └── types/                 # Tipi condivisi
├── pkg/
│   ├── api/                   # gRPC proto + server
│   ├── contracts/             # Contratti front/back
│   └── runtime/               # Trait del driver host
├── tui/                       # Client TUI TypeScript
├── deploy/                    # Deployment containerizzato
├── proto/                     # Definizioni gRPC proto
└── docs/                      # Documentazione (requisiti/architettura/piani/ADR/spec/plan)
```

## Baseline di qualità

| Dimensione | Soglia | Effettivo |
|---|---|---|
| Copertura test Go | ≥ 80 % (core ≥ 90 %) | router 100 % · scheduler 96 % · governance 96 % · codex 94,9 % |
| Lint | 0 nuovi avvisi | golangci-lint + eslint |
| Build | ≤ 5 minuti | CI completo ~3 minuti |
| Sicurezza | 0 CVE ad alta gravità | govulncheck + npm audit |

## Sviluppo

```bash
make all      # lint + vet + test + build
make test     # Suite di test completa
make lint     # golangci-lint
make run      # build + serve
```

## Documentazione

- [AGENTS.md](AGENTS.md) — Standard di progetto (documento radice)
- [docs/01-需求分析文档.md](docs/01-需求分析文档.md) — Requisiti di prodotto
- [docs/02-技术架构设计.md](docs/02-技术架构设计.md) — Progettazione architetturale
- [docs/03-系统设计与实施规划.md](docs/03-系统设计与实施规划.md) — Piano di implementazione
- [docs/adr/](docs/adr/) — Architecture Decision Records (ADR-0001 ~ 0013)
- [docs/CONSTRAINTS.md](docs/CONSTRAINTS.md) — Red line di qualità

## Licenza

[MIT](LICENSE)