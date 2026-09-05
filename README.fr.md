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

> Une plateforme d’orchestration de codage IA écrite en Go. Elle ne détient aucun LLM en propre ; elle délègue l’exécution à des CLI de codage IA externes (Codex et OpenCode en drivers réels ; Claude-Code et DeepSeek-DSH en stubs), simule une équipe de développement logiciel à 9 rôles qui collabore via le protocole A2A, applique une porte de qualité déterministe avec audit de gouvernance, est distribuée en conteneurs, expose quatre protocoles — gRPC, MCP, ACP et A2A — et fournit un client TUI en TypeScript.

[![CI](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml/badge.svg)](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

## Fonctionnalités clés

- **Séparation orchestration / exécution** — Le Coordinator se contente d’orchestrer et n’écrit jamais de code ; les CLI hôtes exécutent mais ne décident pas.
- **Conteneur = rôle** — 9 rôles d’équipe (PM, Architect, Frontend, Backend, QA, Security, DevOps, …), chacun dans son propre conteneur, indépendamment scalable et remplaçable.
- **Collaboration via A2A** — Communication structurée entre agents, support d’InProc (développement) et de Redis Pub/Sub (mode conteneurisé).
- **Porte de qualité déterministe** — Exécution mécanique par `golangci-lint` + `go vet` + `go test`, sans dépendre de l’auto-évaluation du modèle.
- **Base de connaissances RAG + mémoire projet** — Recherche BM25 injectée dans le contexte ; la mémoire accumule les leçons d’échec et les solutions passées.
- **Quatre protocoles externes** — gRPC (TUI), MCP (intégration d’outils externes), ACP (client Agent standard), A2A (maillage inter-instances).
- **Local d’abord** — Le code ne quitte pas le conteneur ; aucune clé API n’est conservée ; toute l’authentification est déléguée à la CLI hôte.

## Vue d’ensemble de l’architecture

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

Voir [docs/04-系统架构图.md](docs/04-系统架构图.md) pour les diagrammes d’architecture complets.

## Démarrage rapide

### Mode binaire unique

```bash
# Build
go build -o bin/aicodingagentteam ./cmd/aicodingagentteam

# Initialiser un projet
./bin/aicodingagentteam init

# Exécuter un pipeline complet
./bin/aicodingagentteam run "Construire une API REST" --backend codex

# Édition rapide
./bin/aicodingagentteam quick "Corriger le style du bouton de connexion"

# Vérification qualité
./bin/aicodingagentteam verify

# Démo de la base de connaissances RAG
./bin/aicodingagentteam knowledge demo

# Démarrer le service Coordinator (gRPC + A2A HTTP)
./bin/aicodingagentteam serve
```

### Client TUI

```bash
cd tui && npm install && npm run build

# Démarrer le Coordinator
cd .. && go run ./cmd/aicodingagentteam serve

# Lancer le TUI (dans un autre terminal)
cd tui && node dist/cli.js
# Mode démo (sans Coordinator)
node dist/cli.js --demo
```

### Déploiement conteneurisé

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

## Référence des commandes

| Commande | Description |
|---|---|
| `init` | Initialiser la configuration du projet |
| `run "requirement"` | Exécuter le pipeline complet |
| `quick "description"` | Édition rapide légère |
| `verify` | Lancer la vérification qualité |
| `govern [--ci] [path]` | Scan de gouvernance |
| `report` | Émettre le rapport de conformité |
| `serve` | Démarrer le Coordinator (gRPC + A2A) |
| `knowledge index [dir]` | Indexer un répertoire dans la base de connaissances |
| `knowledge search "query"` | Récupérer les 5 meilleurs chunks |
| `knowledge demo` | Démo end-to-end RAG + mémoire |
| `version` | Afficher la version |

## Structure du projet

```
agent_team/
├── cmd/aicodingagentteam/     # Point d’entrée du binaire Go
├── internal/
│   ├── a2a/                   # Protocole A2A (InProc + Redis)
│   ├── acp/                   # Agent Client Protocol
│   ├── agent/                 # Registre des 9 rôles
│   ├── audit/                 # Journal d’audit
│   ├── config/                # Chargeur de configuration
│   ├── coordinator/           # Cœur d’orchestration (flux 5 couches)
│   ├── governance/            # Règles de gouvernance (113+ règles)
│   ├── host/                  # Drivers hôtes (codex/opencode/claude/dsh)
│   ├── knowledge/             # Moteur de recherche RAG (BM25)
│   ├── mcp/                   # Model Context Protocol
│   ├── memory/                # Mémoire projet (facts/pitfalls/lessons)
│   ├── planner/               # Constructeur de plan DAG
│   ├── qualitygate/           # Moteur de porte de qualité
│   ├── router/                # Routeur d’intention
│   ├── scheduler/             # Planificateur de rôles + verrou single-writer
│   └── types/                 # Types partagés
├── pkg/
│   ├── api/                   # gRPC proto + serveur
│   ├── contracts/             # Contrats front/back
│   └── runtime/               # Trait de driver hôte
├── tui/                       # Client TUI TypeScript
├── deploy/                    # Déploiement conteneurisé
├── proto/                     # Définitions gRPC proto
└── docs/                      # Documentation (requirements/architecture/plans/ADR/spec/plan)
```

## Base qualité

| Dimension | Seuil | Réel |
|---|---|---|
| Couverture tests Go | ≥ 80 % (cœur ≥ 90 %) | router 100 % · scheduler 96 % · governance 96 % · codex 94,9 % |
| Lint | 0 nouvel avertissement | golangci-lint + eslint |
| Build | ≤ 5 minutes | CI complète ~3 minutes |
| Sécurité | 0 CVE haute sévérité | govulncheck + npm audit |

## Développement

```bash
make all      # lint + vet + test + build
make test     # Suite de tests complète
make lint     # golangci-lint
make run      # build + serve
```

## Documentation

- [AGENTS.md](AGENTS.md) — Standards du projet (document racine)
- [docs/01-需求分析文档.md](docs/01-需求分析文档.md) — Exigences produit
- [docs/02-技术架构设计.md](docs/02-技术架构设计.md) — Conception d’architecture
- [docs/03-系统设计与实施规划.md](docs/03-系统设计与实施规划.md) — Plan d’implémentation
- [docs/adr/](docs/adr/) — Architecture Decision Records (ADR-0001 ~ 0013)
- [docs/CONSTRAINTS.md](docs/CONSTRAINTS.md) — Lignes rouges qualité

## Licence

[MIT](LICENSE)