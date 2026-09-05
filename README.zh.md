# AiCodingAgentTeam

## 🌐 Languages / 语言

[English](README.md) · [简体中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Русский](README.ru.md) · [Español](README.es.md) · [Italiano](README.it.md)
> 基于 Golang 的 AI 编码编排平台——本身不拥有大模型，调度外部 AI 编码 CLI（Codex、OpenCode 真实驱动；Claude-Code、DeepSeek-DSH stub），模拟软件开发团队 9 角色协作，通过 A2A 协议通信，带质量门禁与治理审计，容器化部署，对外暴露 gRPC/MCP/ACP/A2A 四协议，提供 TypeScript TUI 客户端。

[![CI](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml/badge.svg)](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

## 核心特性

- **编排与执行分离**：Coordinator 只编排不写代码，宿主 CLI 只执行不决策
- **容器即角色**：9 个团队角色（PM/Architect/Frontend/Backend/QA/Security/DevOps 等），独立容器可伸缩替换
- **A2A 协议协作**：Agent 间结构化通信，支持 InProc（开发）+ Redis Pub/Sub（容器化）
- **确定性质量门禁**：golangci-lint + go vet + go test 机器硬执行，不依赖模型自评
- **RAG 知识库 + 项目记忆**：BM25 检索注入上下文，记忆库沉淀失败教训与历史方案
- **四协议对外暴露**：gRPC（TUI 通信）、MCP（外部工具集成）、ACP（标准 Agent 客户端）、A2A（跨实例互联）
- **本地优先**：代码不出容器，不持有任何 API Key，鉴权全部交底层宿主 CLI

## 架构总览

```
用户 → TypeScript TUI (Ink) → gRPC → Coordinator
                                         ├── Router (意图路由)
                                         ├── Planner (DAG 计划)
                                         ├── Scheduler (角色调度)
                                         │    ├── A2A Bus → Reviewer Agents (并行)
                                         │    └── Single-Writer Lock → Writer Agents (串行)
                                         ├── Knowledge (BM25 RAG 检索)
                                         ├── Memory (facts/pitfalls/lessons)
                                         └── Quality Gate (确定性校验)
                                              ↓
宿主 CLI: Codex (real) | OpenCode (real) | Claude (stub) | DSH (stub)
```

详细架构图见 [docs/04-系统架构图.md](docs/04-系统架构图.md)。

## 快速上手

### 单二进制模式

```bash
# 构建
go build -o bin/aicodingagentteam ./cmd/aicodingagentteam

# 初始化项目
./bin/aicodingagentteam init

# 运行完整流水线
./bin/aicodingagentteam run "搭建一个 REST API" --backend codex

# 快速编辑
./bin/aicodingagentteam quick "修复登录按钮样式"

# 质量校验
./bin/aicodingagentteam verify

# RAG 知识库 demo
./bin/aicodingagentteam knowledge demo

# 启动协调器服务（gRPC + A2A HTTP）
./bin/aicodingagentteam serve
```

### TUI 客户端

```bash
cd tui && npm install && npm run build

# 启动协调器
cd .. && go run ./cmd/aicodingagentteam serve

# 启动 TUI（另一终端）
cd tui && node dist/cli.js
# Demo 模式（无需协调器）
node dist/cli.js --demo
```

### 容器化部署

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

## 命令速查

| 命令 | 说明 |
|---|---|
| `init` | 初始化项目配置 |
| `run "需求"` | 运行完整流水线 |
| `quick "描述"` | 轻量快速编辑 |
| `verify` | 执行质量校验 |
| `govern [--ci] [path]` | 治理扫描 |
| `report` | 输出合规报告 |
| `serve` | 启动协调器（gRPC + A2A） |
| `knowledge index [dir]` | 索引目录到知识库 |
| `knowledge search "query"` | 检索 top-5 知识块 |
| `knowledge demo` | 端到端 RAG + 记忆演示 |
| `version` | 打印版本 |

## 项目结构

```
agent_team/
├── cmd/aicodingagentteam/     # Go 二进制入口
├── internal/
│   ├── a2a/                   # A2A 协议（InProc + Redis）
│   ├── acp/                  # Agent Client Protocol
│   ├── agent/                # 9 角色注册
│   ├── audit/                # 审计日志
│   ├── config/               # 配置加载
│   ├── coordinator/          # 编排核心（5 层流）
│   ├── governance/           # 治理规则（113+ 规则）
│   ├── host/                 # 宿主驱动（codex/opencode/claude/dsh）
│   ├── knowledge/            # RAG 检索引擎（BM25）
│   ├── mcp/                  # Model Context Protocol
│   ├── memory/               # 项目记忆（facts/pitfalls/lessons）
│   ├── planner/              # DAG 计划构建
│   ├── qualitygate/          # 质量门禁引擎
│   ├── router/               # 意图路由
│   ├── scheduler/            # 角色调度 + 单写者
│   └── types/                # 共享类型
├── pkg/
│   ├── api/                  # gRPC proto + server
│   ├── contracts/            # 前后端契约
│   └── runtime/              # 宿主驱动 trait
├── tui/                       # TypeScript TUI 客户端
├── deploy/                    # 容器化部署
├── proto/                     # gRPC proto 定义
└── docs/                      # 文档（需求/架构/规划/ADR/spec/plan）
```

## 质量基线

| 维度 | 阈值 | 实际 |
|---|---|---|
| Go 测试覆盖率 | ≥ 80%（核心 ≥ 90%） | router 100% · scheduler 96% · governance 96% · codex 94.9% |
| Lint | 0 新增警告 | golangci-lint + eslint |
| 构建 | ≤ 5 分钟 | CI 全流程 ~3 分钟 |
| 安全 | 0 高危漏洞 | govulncheck + npm audit |

## 开发

```bash
make all      # lint + vet + test + build
make test     # 全量测试
make lint     # golangci-lint
make run      # build + serve
```

## 文档

- [AGENTS.md](AGENTS.md) — 项目规范总纲
- [docs/01-需求分析文档.md](docs/01-需求分析文档.md) — 产品需求
- [docs/02-技术架构设计.md](docs/02-技术架构设计.md) — 架构设计
- [docs/03-系统设计与实施规划.md](docs/03-系统设计与实施规划.md) — 实施规划
- [docs/adr/](docs/adr/) — 架构决策记录（ADR-0001 ~ 0013）
- [docs/CONSTRAINTS.md](docs/CONSTRAINTS.md) — 质量红线

## License

[MIT](LICENSE)