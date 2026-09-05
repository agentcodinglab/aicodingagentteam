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

> Una plataforma de orquestación de codificación con IA escrita en Go. No posee un LLM propio; delega el trabajo a CLI externas de codificación con IA (Codex y OpenCode como drivers reales; Claude-Code y DeepSeek-DSH como stubs), simula un equipo de desarrollo de software de 9 roles que colabora mediante el protocolo A2A, aplica una puerta de calidad determinista con auditoría de gobernanza, se distribuye en contenedores y expone cuatro protocolos — gRPC, MCP, ACP y A2A — además de un cliente TUI en TypeScript.

[![CI](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml/badge.svg)](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

## Características principales

- **Separación de orquestación y ejecución** — El Coordinator solo orquesta y nunca escribe código; las CLI anfitrionas solo ejecutan y nunca deciden.
- **Contenedor = rol** — 9 roles de equipo (PM, Architect, Frontend, Backend, QA, Security, DevOps, …), cada uno en su propio contenedor, escalable y reemplazable de forma independiente.
- **Colaboración vía A2A** — Comunicación estructurada entre agentes; soporta InProc (desarrollo) y Redis Pub/Sub (modo contenedorizado).
- **Puerta de calidad determinista** — Ejecutada de forma rígida por `golangci-lint` + `go vet` + `go test`, sin depender de la autoevaluación del modelo.
- **Base de conocimientos RAG + memoria del proyecto** — Búsqueda BM25 inyectada en el contexto; la memoria acumula lecciones de fallos y soluciones históricas.
- **Cuatro protocolos externos** — gRPC (TUI), MCP (integración de herramientas externas), ACP (cliente Agent estándar), A2A (mesh entre instancias).
- **Local primero** — El código no sale del contenedor; no se conservan claves de API; toda la autenticación se delega a la CLI anfitriona.

## Vista general de la arquitectura

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

Los diagramas completos están en [docs/04-系统架构图.md](docs/04-系统架构图.md).

## Inicio rápido

### Modo binario único

```bash
# Compilar
go build -o bin/aicodingagentteam ./cmd/aicodingagentteam

# Inicializar el proyecto
./bin/aicodingagentteam init

# Ejecutar el pipeline completo
./bin/aicodingagentteam run "Construir una API REST" --backend codex

# Edición rápida
./bin/aicodingagentteam quick "Arreglar el estilo del botón de inicio de sesión"

# Verificación de calidad
./bin/aicodingagentteam verify

# Demo de la base de conocimientos RAG
./bin/aicodingagentteam knowledge demo

# Iniciar el servicio Coordinator (gRPC + A2A HTTP)
./bin/aicodingagentteam serve
```

### Cliente TUI

```bash
cd tui && npm install && npm run build

# Iniciar el Coordinator
cd .. && go run ./cmd/aicodingagentteam serve

# Lanzar la TUI (en otra terminal)
cd tui && node dist/cli.js
# Modo demo (sin Coordinator)
node dist/cli.js --demo
```

### Despliegue en contenedores

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

## Referencia de comandos

| Comando | Descripción |
|---|---|
| `init` | Inicializar configuración del proyecto |
| `run "requirement"` | Ejecutar el pipeline completo |
| `quick "description"` | Edición rápida ligera |
| `verify` | Ejecutar la verificación de calidad |
| `govern [--ci] [path]` | Escaneo de gobernanza |
| `report` | Emitir informe de cumplimiento |
| `serve` | Iniciar Coordinator (gRPC + A2A) |
| `knowledge index [dir]` | Indexar un directorio en la base de conocimientos |
| `knowledge search "query"` | Recuperar los 5 mejores chunks |
| `knowledge demo` | Demo end-to-end RAG + memoria |
| `version` | Mostrar versión |

## Estructura del proyecto

```
agent_team/
├── cmd/aicodingagentteam/     # Entrada del binario Go
├── internal/
│   ├── a2a/                   # Protocolo A2A (InProc + Redis)
│   ├── acp/                   # Agent Client Protocol
│   ├── agent/                 # Registro de los 9 roles
│   ├── audit/                 # Registro de auditoría
│   ├── config/                # Cargador de configuración
│   ├── coordinator/           # Núcleo de orquestación (flujo de 5 capas)
│   ├── governance/            # Reglas de gobernanza (113+ reglas)
│   ├── host/                  # Drivers anfitriones (codex/opencode/claude/dsh)
│   ├── knowledge/             # Motor de recuperación RAG (BM25)
│   ├── mcp/                   # Model Context Protocol
│   ├── memory/                # Memoria del proyecto (facts/pitfalls/lessons)
│   ├── planner/               # Constructor del plan DAG
│   ├── qualitygate/           # Motor de quality gate
│   ├── router/                # Router de intención
│   ├── scheduler/             # Scheduler de roles + cerrojo single-writer
│   └── types/                 # Tipos compartidos
├── pkg/
│   ├── api/                   # gRPC proto + servidor
│   ├── contracts/             # Contratos front/back
│   └── runtime/               # Trait del driver anfitrión
├── tui/                       # Cliente TUI TypeScript
├── deploy/                    # Despliegue en contenedores
├── proto/                     # Definiciones gRPC proto
└── docs/                      # Documentación (requisitos/arquitectura/planes/ADR/spec/plan)
```

## Línea base de calidad

| Dimensión | Umbral | Real |
|---|---|---|
| Cobertura de tests Go | ≥ 80 % (núcleo ≥ 90 %) | router 100 % · scheduler 96 % · governance 96 % · codex 94,9 % |
| Lint | 0 nuevas advertencias | golangci-lint + eslint |
| Build | ≤ 5 minutos | CI completo ~3 minutos |
| Seguridad | 0 CVE de alta severidad | govulncheck + npm audit |

## Desarrollo

```bash
make all      # lint + vet + test + build
make test     # Suite completa de tests
make lint     # golangci-lint
make run      # build + serve
```

## Documentación

- [AGENTS.md](AGENTS.md) — Estándares del proyecto (documento raíz)
- [docs/01-需求分析文档.md](docs/01-需求分析文档.md) — Requisitos del producto
- [docs/02-技术架构设计.md](docs/02-技术架构设计.md) — Diseño de arquitectura
- [docs/03-系统设计与实施规划.md](docs/03-系统设计与实施规划.md) — Plan de implementación
- [docs/adr/](docs/adr/) — Architecture Decision Records (ADR-0001 ~ 0013)
- [docs/CONSTRAINTS.md](docs/CONSTRAINTS.md) — Líneas rojas de calidad

## Licencia

[MIT](LICENSE)