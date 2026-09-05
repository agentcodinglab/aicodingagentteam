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

> Go 기반 AI 코딩 오케스트레이션 플랫폼입니다. 자체 LLM 을 보유하지 않고, 외부 AI 코딩 CLI(Codex 및 OpenCode 는 실제 드라이버, Claude-Code 및 DeepSeek-DSH 는 스텁)에 작업을 위임하여 A2A 프로토콜로 협업하는 9 개 역할의 소프트웨어 개발 팀을 시뮬레이션합니다. 결정론적 품질 게이트와 거버넌스 감사를 적용하고, 컨테이너로 배포되며, gRPC / MCP / ACP / A2A 의 4 종 프로토콜을 외부에 노출하고 TypeScript TUI 클라이언트를 함께 제공합니다.

[![CI](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml/badge.svg)](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

## 핵심 기능

- **오케스트레이션과 실행의 분리** — Coordinator 는 오케스트레이션만 수행하고 코드를 작성하지 않습니다. 호스트 CLI 는 실행만 담당하며 결정을 내리지 않습니다.
- **컨테이너 = 역할** — 9 개 팀 역할(PM / Architect / Frontend / Backend / QA / Security / DevOps 등)이 각각 독립된 컨테이너에서 실행되며, 독립적으로 확장 및 교체할 수 있습니다.
- **A2A 프로토콜 협업** — 에이전트 간 구조화된 통신을 지원하며 InProc(개발) 와 Redis Pub/Sub(컨테이너) 모드를 제공합니다.
- **결정론적 품질 게이트** — `golangci-lint` + `go vet` + `go test` 를 기계적으로 실행하며 모델 자체 평가에 의존하지 않습니다.
- **RAG 지식 베이스 + 프로젝트 메모리** — BM25 검색을 컨텍스트에 주입하고, 실패 교훈과 과거 해결책을 메모리에 축적합니다.
- **4 종 외부 프로토콜** — gRPC(TUI 통신), MCP(외부 도구 통합), ACP(표준 Agent 클라이언트), A2A(인스턴스 간 메시).
- **로컬 우선** — 코드는 컨테이너 밖으로 나가지 않으며 API 키를 보유하지 않습니다. 모든 인증은 하위 호스트 CLI 에 위임됩니다.

## 아키텍처 개요

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

전체 아키텍처 다이어그램은 [docs/04-系统架构图.md](docs/04-系统架构图.md) 를 참조하세요.

## 빠른 시작

### 단일 바이너리 모드

```bash
# 빌드
go build -o bin/aicodingagentteam ./cmd/aicodingagentteam

# 프로젝트 초기화
./bin/aicodingagentteam init

# 전체 파이프라인 실행
./bin/aicodingagentteam run "REST API 를 만들어줘" --backend codex

# 빠른 편집
./bin/aicodingagentteam quick "로그인 버튼 스타일 수정"

# 품질 검증
./bin/aicodingagentteam verify

# RAG 지식 베이스 데모
./bin/aicodingagentteam knowledge demo

# Coordinator 서비스 시작(gRPC + A2A HTTP)
./bin/aicodingagentteam serve
```

### TUI 클라이언트

```bash
cd tui && npm install && npm run build

# Coordinator 시작
cd .. && go run ./cmd/aicodingagentteam serve

# TUI 시작(다른 터미널에서)
cd tui && node dist/cli.js
# 데모 모드(Coordinator 불필요)
node dist/cli.js --demo
```

### 컨테이너 배포

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

## 명령어 참조

| 명령어 | 설명 |
|---|---|
| `init` | 프로젝트 설정 초기화 |
| `run "requirement"` | 전체 파이프라인 실행 |
| `quick "description"` | 경량 빠른 편집 |
| `verify` | 품질 검증 실행 |
| `govern [--ci] [path]` | 거버넌스 스캔 |
| `report` | 컴플라이언스 보고서 출력 |
| `serve` | Coordinator 시작(gRPC + A2A) |
| `knowledge index [dir]` | 디렉터리를 지식 베이스에 인덱싱 |
| `knowledge search "query"` | 상위 5 개 지식 청크 조회 |
| `knowledge demo` | RAG + 메모리 end-to-end 데모 |
| `version` | 버전 출력 |

## 프로젝트 구조

```
agent_team/
├── cmd/aicodingagentteam/     # Go 바이너리 진입점
├── internal/
│   ├── a2a/                   # A2A 프로토콜(InProc + Redis)
│   ├── acp/                   # Agent Client Protocol
│   ├── agent/                 # 9 개 역할 레지스트리
│   ├── audit/                 # 감사 로그
│   ├── config/                # 설정 로더
│   ├── coordinator/           # 오케스트레이션 코어(5 계층 흐름)
│   ├── governance/            # 거버넌스 규칙(113+ 규칙)
│   ├── host/                  # 호스트 드라이버(codex/opencode/claude/dsh)
│   ├── knowledge/             # RAG 검색 엔진(BM25)
│   ├── mcp/                   # Model Context Protocol
│   ├── memory/                # 프로젝트 메모리(facts/pitfalls/lessons)
│   ├── planner/               # DAG 플래너
│   ├── qualitygate/           # 품질 게이트 엔진
│   ├── router/                # 인텐트 라우터
│   ├── scheduler/             # 역할 스케줄러 + 단일 작성자 락
│   └── types/                 # 공유 타입
├── pkg/
│   ├── api/                   # gRPC proto + server
│   ├── contracts/             # 프론트/백엔드 계약
│   └── runtime/               # 호스트 드라이버 trait
├── tui/                       # TypeScript TUI 클라이언트
├── deploy/                    # 컨테이너 배포
├── proto/                     # gRPC proto 정의
└── docs/                      # 문서(요구사항/아키/계획/ADR/spec/plan)
```

## 품질 기준선

| 항목 | 임계값 | 실제 |
|---|---|---|
| Go 테스트 커버리지 | ≥ 80%(코어 ≥ 90%) | router 100% · scheduler 96% · governance 96% · codex 94.9% |
| Lint | 신규 경고 0 | golangci-lint + eslint |
| 빌드 | ≤ 5 분 | CI 전체 ~3 분 |
| 보안 | 고위험 CVE 0 | govulncheck + npm audit |

## 개발

```bash
make all      # lint + vet + test + build
make test     # 전체 테스트
make lint     # golangci-lint
make run      # build + serve
```

## 문서

- [AGENTS.md](AGENTS.md) — 프로젝트 규범(루트 문서)
- [docs/01-需求分析文档.md](docs/01-需求分析文档.md) — 제품 요구사항
- [docs/02-技术架构设计.md](docs/02-技术架构设计.md) — 아키텍처 설계
- [docs/03-系统设计与实施规划.md](docs/03-系统设计与实施规划.md) — 구현 계획
- [docs/adr/](docs/adr/) — 아키텍처 결정 기록(ADR-0001 ~ 0013)
- [docs/CONSTRAINTS.md](docs/CONSTRAINTS.md) — 품질 레드라인

## 라이선스

[MIT](LICENSE)