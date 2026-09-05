# AiCodingAgentTeam

## 🌐 Languages / 语言

[English](README.md) · [简体中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Русский](README.ru.md) · [Español](README.es.md) · [Italiano](README.it.md)
> Платформа оркестровки AI-кодинга на Go. Сама по себе она не содержит LLM, а делегирует работу внешним CLI-кодинга (Codex и OpenCode — реальные драйверы; Claude-Code и DeepSeek-DSH — заглушки), моделирует команду разработки ПО из 9 ролей, взаимодействующих через протокол A2A, обеспечивает детерминированный quality gate с аудитом governance, поставляется в контейнерах и предоставляет наружу четыре протокола — gRPC, MCP, ACP и A2A, а также TUI-клиент на TypeScript.

[![CI](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml/badge.svg)](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

## Ключевые возможности

- **Разделение оркестровки и исполнения** — Coordinator только оркеструет и не пишет код; CLI-хосты только исполняют и не принимают решений.
- **Контейнер = роль** — 9 ролей команды (PM, Architect, Frontend, Backend, QA, Security, DevOps, …), каждая в своём контейнере, независимо масштабируется и заменяется.
- **Совместная работа через A2A** — Структурированное межагентное взаимодействие; поддерживаются InProc (для разработки) и Redis Pub/Sub (контейнерный режим).
- **Детерминированный quality gate** — Жёсткое исполнение через `golangci-lint` + `go vet` + `go test`, без самооценки модели.
- **RAG-база знаний + память проекта** — BM25-поиск внедряется в контекст; память накапливает уроки неудач и прошлые решения.
- **Четыре внешних протокола** — gRPC (TUI), MCP (интеграция внешних инструментов), ACP (стандартный клиент Agent), A2A (mesh между инстансами).
- **Локальный приоритет** — Код не покидает контейнер; API-ключи не хранятся; вся аутентификация делегируется хостовой CLI.

## Обзор архитектуры

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

Полные диаграммы архитектуры см. в [docs/04-系统架构图.md](docs/04-系统架构图.md).

## Быстрый старт

### Режим единого бинарника

```bash
# Сборка
go build -o bin/aicodingagentteam ./cmd/aicodingagentteam

# Инициализация проекта
./bin/aicodingagentteam init

# Запуск полного пайплайна
./bin/aicodingagentteam run "Создать REST API" --backend codex

# Быстрая правка
./bin/aicodingagentteam quick "Исправить стиль кнопки входа"

# Проверка качества
./bin/aicodingagentteam verify

# Демо RAG-базы знаний
./bin/aicodingagentteam knowledge demo

# Запуск сервиса Coordinator (gRPC + A2A HTTP)
./bin/aicodingagentteam serve
```

### TUI-клиент

```bash
cd tui && npm install && npm run build

# Запустить Coordinator
cd .. && go run ./cmd/aicodingagentteam serve

# Запустить TUI (в другом терминале)
cd tui && node dist/cli.js
# Демо-режим (без Coordinator)
node dist/cli.js --demo
```

### Контейнерное развёртывание

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

## Справочник команд

| Команда | Описание |
|---|---|
| `init` | Инициализация конфигурации проекта |
| `run "requirement"` | Запуск полного пайплайна |
| `quick "description"` | Лёгкая быстрая правка |
| `verify` | Выполнить проверку качества |
| `govern [--ci] [path]` | Сканирование governance |
| `report` | Сформировать отчёт о соответствии |
| `serve` | Запустить Coordinator (gRPC + A2A) |
| `knowledge index [dir]` | Индексировать каталог в базу знаний |
| `knowledge search "query"` | Получить топ-5 фрагментов знаний |
| `knowledge demo` | Сквозное демо RAG + памяти |
| `version` | Показать версию |

## Структура проекта

```
agent_team/
├── cmd/aicodingagentteam/     # Точка входа Go-бинарника
├── internal/
│   ├── a2a/                   # Протокол A2A (InProc + Redis)
│   ├── acp/                   # Agent Client Protocol
│   ├── agent/                 # Реестр 9 ролей
│   ├── audit/                 # Журнал аудита
│   ├── config/                # Загрузчик конфигурации
│   ├── coordinator/           # Ядро оркестровки (5-слойный поток)
│   ├── governance/            # Правила governance (113+ правил)
│   ├── host/                  # Драйверы хостов (codex/opencode/claude/dsh)
│   ├── knowledge/             # Поисковый движок RAG (BM25)
│   ├── mcp/                   # Model Context Protocol
│   ├── memory/                # Память проекта (facts/pitfalls/lessons)
│   ├── planner/               # Конструктор DAG-плана
│   ├── qualitygate/           # Движок quality gate
│   ├── router/                # Маршрутизатор намерений
│   ├── scheduler/             # Планировщик ролей + single-writer lock
│   └── types/                 # Общие типы
├── pkg/
│   ├── api/                   # gRPC proto + сервер
│   ├── contracts/             # Контракты фронт/бэк
│   └── runtime/               # Trait драйвера хоста
├── tui/                       # TypeScript TUI-клиент
├── deploy/                    # Контейнерное развёртывание
├── proto/                     # Определения gRPC proto
└── docs/                      # Документация (требования/архитектура/планы/ADR/spec/plan)
```

## Базовый уровень качества

| Показатель | Порог | Факт |
|---|---|---|
| Покрытие тестами Go | ≥ 80 % (ядро ≥ 90 %) | router 100 % · scheduler 96 % · governance 96 % · codex 94,9 % |
| Lint | 0 новых предупреждений | golangci-lint + eslint |
| Сборка | ≤ 5 минут | Полный CI ~3 минуты |
| Безопасность | 0 CVE высокой критичности | govulncheck + npm audit |

## Разработка

```bash
make all      # lint + vet + test + build
make test     # Полный набор тестов
make lint     # golangci-lint
make run      # build + serve
```

## Документация

- [AGENTS.md](AGENTS.md) — Стандарты проекта (корневой документ)
- [docs/01-需求分析文档.md](docs/01-需求分析文档.md) — Продуктовые требования
- [docs/02-技术架构设计.md](docs/02-技术架构设计.md) — Проект архитектуры
- [docs/03-系统设计与实施规划.md](docs/03-系统设计与实施规划.md) — План реализации
- [docs/adr/](docs/adr/) — Architecture Decision Records (ADR-0001 ~ 0013)
- [docs/CONSTRAINTS.md](docs/CONSTRAINTS.md) — Красные линии качества

## Лицензия

[MIT](LICENSE)