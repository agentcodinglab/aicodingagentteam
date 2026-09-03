# 系统设计与实施规划 — AiCodingAgentTeam

> 项目名称：AiCodingAgentTeam
> 文档版本：v1.0
> 编写日期：2026-09-02
> 关联文档：01-需求分析文档.md、02-技术架构设计.md

---

## 目录
1. [配置体系设计](#1-配置体系设计)
2. [命令系统设计](#2-命令系统设计)
3. [编译构建与发布](#3-编译构建与发布)
4. [技术选型清单](#4-技术选型清单)
5. [实施阶段规划](#5-实施阶段规划)
6. [开发里程碑](#6-开发里程碑)
7. [风险与对策](#7-风险与对策)
8. [源码阅读与开发路径](#8-源码阅读与开发路径)

---

## 1. 配置体系设计

### 1.1 全局配置 `~/.aicodingaicodingaicodingagentteam/config.toml`

```toml
[default]
backend = "codex"              # 默认宿主 CLI
workspace = "."                # 默认工作目录
auto_approve_gates = false     # 自动通过门禁（CI 模式）

[coordinator]
port_grpc = 8080               # TUI 通信端口
port_mcp  = 8081               # MCP 服务端口
port_acp  = 8082               # ACP 服务端口
port_a2a  = 8083               # A2A 服务端口

[a2a]
bus = "redis://localhost:6379" # A2A 消息总线
agent_discovery = "static"      # static / dns / consul

[host]
claude_bin = "claude"           # 可被环境变量覆盖
codex_bin = "codex"
opencode_bin = "opencode"
dsh_bin = "dsh"
worker_timeout = 300            # 单任务超时秒数

[quality]
threshold = 90                 # 质量得分阈值
skip_checks = []

[governance]
fail_open = true               # 治理故障放行
audit_dir = ".aicodingaicodingaicodingagentteam/audit"

[knowledge]
enabled = true
engine = "hybrid"              # bm25 / hybrid
top_k = 6
cloud_embed = false            # 云端向量默认关闭

[memory]
capture = true
recall = true
scope = "project"              # 不跨项目共享
```

### 1.2 项目配置 `.aicodingaicodingagentteamrc`

```toml
[project]
name = "my-app"
stack = ["go", "react", "postgres"]

[governance]
[disabled]
clauses = []
[exclusions]
paths = ["src/legacy/**", "**/*.test.ts"]
```

### 1.3 环境变量全集

| 环境变量 | 作用 | 默认值 |
|---|---|---|
| `AICODINGAGENTTEAM_*_BIN` | 覆盖各宿主二进制路径 | — |
| `AICODINGAGENTTEAM_WORKER_TIMEOUT` | 单任务超时秒数 | 300 |
| `AICODINGAGENTTEAM_VERIFY_TIMEOUT_SECS` | 校验超时秒数 | 120 |
| `AICODINGAGENTTEAM_NO_GOAL_MODE` | `=1` 关闭 /goal 模式 | — |
| `AICODINGAGENTTEAM_EMBED_MODEL_DIR` | 本地向量模型目录 | — |
| `OPENAI_EMBED_KEY` | 远程 embedding 密钥 | — |
| `AICODINGAGENTTEAM_ALLOW_CLOUD_EMBED` | `=1` 允许云端 embedding | 0 |
| `AICODINGAGENTTEAM_A2A_BUS` | A2A 消息总线地址 | redis://localhost:6379 |
| `AICODINGAGENTTEAM_WORKSPACE` | 工作目录路径 | . |
| `AGENT_ROLE` | Agent 容器角色标识 | — |

---

## 2. 命令系统设计

### 2.1 CLI 子命令

```bash
# 项目管理
aicodingagentteam init                        # 初始化项目配置
aicodingagentteam adopt                       # 导入已有存量项目

# 流程执行
aicodingagentteam run "需求描述" --backend codex   # 非交互完整流水线
aicodingagentteam quick "修复登录按钮样式"          # 轻量快速编辑
aicodingagentteam continue                    # 继续暂停的任务
aicodingagentteam revise                      # 修订文档

# 质量与治理
aicodingagentteam verify --runtime            # 执行质量校验 + 运行时探针
aicodingagentteam report                      # 输出合规报告
aicodingagentteam ci                          # CI 治理扫描

# 服务模式
aicodingagentteam mcp serve                   # 启动 MCP 服务
aicodingagentteam a2a serve                   # 启动 A2A 服务
aicodingagentteam acp serve                   # 启动 ACP 服务

# 知识与记忆
aicodingagentteam knowledge-manage add ./docs # 添加自定义知识库
aicodingagentteam memory capture off --scope project --store facts
aicodingagentteam memory recall off --scope project --store recipes
```

### 2.2 TUI Slash 命令

| 分类 | 命令 |
|---|---|
| 流程控制 | `/run` `/goal` `/quick` `/plan` `/continue` `/revise` |
| 宿主切换 | `/backend claude` `/backend codex` ... |
| 质量检查 | `/verify` `/report` |
| 知识管理 | `/knowledge add` `/knowledge list` |
| 记忆管理 | `/memory on` `/memory off` `/memory show` |
| 预览 | `/preview` `/deploy-check` |

---

## 3. 编译构建与发布

### 3.1 源码编译环境

- Go ≥ 1.22；
- Docker ≥ 24.0（容器化构建）；
- Node.js ≥ 18（仅 TUI 客户端构建）。

### 3.2 构建命令

```bash
# 克隆
git clone https://github.com/agentcodinglab/aicodingagentteam.git
cd aicodingagentteam

# 构建 Go 主二进制
go build -o bin/aicodingagentteam ./cmd/aicodingagentteam

# 构建 TUI 客户端
cd tui && npm install && npm run build

# 构建 Docker 镜像
docker compose -f deploy/compose/docker-compose.yml build

# 跨平台构建
GOOS=linux   GOARCH=amd64 go build -o bin/aicodingaicodingagentteam-linux-amd64 ./cmd/aicodingagentteam
GOOS=darwin  GOARCH=arm64 go build -o bin/aicodingaicodingagentteam-darwin-arm64 ./cmd/aicodingagentteam
GOOS=windows GOARCH=amd64 go build -o bin/aicodingagentteam.exe ./cmd/aicodingagentteam
```

### 3.3 发布分发

| 产物 | 渠道 | 说明 |
|---|---|---|
| Docker 镜像 | Docker Hub / GHCR | `aicodingaicodingagentteam/coordinator:latest` |
| TUI npm 包 | npm registry | `aicodingaicodingagentteam-tui` |
| Release 二进制 | GitHub Releases | 跨平台 Go 二进制 |

### 3.4 构建优化

```dockerfile
# deploy/docker/coordinator.Dockerfile — 多阶段构建
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /aicodingagentteam ./cmd/aicodingagentteam

FROM alpine:3.19
RUN apk add --no-cache ca-certificates git
COPY --from=builder /aicodingagentteam /usr/local/bin/aicodingagentteam
ENTRYPOINT ["aicodingagentteam"]
```

---

## 4. 技术选型清单

| 层 | 组件 | 选型 | 理由 |
|---|---|---|---|
| 语言 | 后端 | Go 1.22+ | 并发模型、单二进制、跨平台 |
| 语言 | TUI | TypeScript | 生态丰富、跨平台 npm 分发 |
| TUI 框架 | 渲染 | Ink | React for CLI，TS 原生 |
| 通信 | 内部 RPC | gRPC | 强类型、流式、高性能 |
| 通信 | A2A 总线 | Redis | 轻量 Pub/Sub，容器友好 |
| 通信 | 对外 | MCP / ACP / A2A | 三协议覆盖 |
| 序列化 | 结构化 | Protocol Buffers / JSON | gRPC 用 PB，A2A 用 JSON |
| 检索 | BM25 | 纯 Go 实现 | 零依赖保底 |
| 检索 | 向量 | ONNX Runtime (Go) | 本地 e5-small 推理 |
| 日志 | 结构化 | zerolog | 高性能 JSON 日志 |
| 指标 | 监控 | Prometheus client | 标准生态 |
| 追踪 | 分布式 | OpenTelemetry | 跨容器链路追踪 |
| 配置 | 格式 | TOML | 人类可读 |
| 容器 | 编排 | Docker Compose | 开发友好；K8s 可选 |
| CI | 构建 | GitHub Actions | 多平台矩阵 |


---

## 5. 实施阶段规划

### Phase 0：项目初始化（1-2 周）

| 任务 | 产出 |
|---|---|
| Go Module 脚手架搭建 | `go.mod` + 包结构 |
| 核心类型定义（`internal/types/`） | 共享类型、接口 |
| Runtime trait 接口定义 | `pkg/runtime/runtime.go` |
| 配置加载框架 | `internal/config/` |
| Docker Compose 骨架 | `deploy/compose/` |
| TUI 项目脚手架 | `tui/` npm 项目 |

### Phase 1：宿主驱动层（3-4 周）

| 任务 | 产出 |
|---|---|
| Codex JSON-RPC 驱动 | `internal/host/codex/` |
| Claude-Code 流协议驱动 | `internal/host/claude/` |
| OpenCode HTTP-SSE 驱动 | `internal/host/opencode/` |
| DeepSeek-DSH 驱动 | `internal/host/dsh/` |
| 能力差异处理 | Capabilities 查询 + 降级逻辑 |
| 宿主容器 Dockerfile | `deploy/docker/host-*` |
| 集成测试 | 各驱动冒烟测试 |

**关键里程碑**：4 款 CLI 均可被 Coordinator 调度执行简单编码任务。

### Phase 2：编排核心引擎（4-5 周）

| 任务 | 产出 |
|---|---|
| 意图路由器 | `internal/router/` |
| DAG 计划构建器 | `internal/planner/` + `plan.json` |
| 角色调度器 | `internal/scheduler/` |
| Coordinator 调度循环 | `internal/coordinator/director.go` |
| 工件文件管理 | 工作目录结构 + 读写隔离 |
| 规模自适应判定 | 任务复杂度评分 |
| 完整 Build 流水线串联 | 9 阶段端到端跑通 |

**关键里程碑**：绿场项目可从需求 → 完整源码交付，无人工门禁自动模式。

### Phase 3：A2A 协议与容器化（3-4 周）

| 任务 | 产出 |
|---|---|
| A2A 协议实现 | `internal/a2a/` |
| Agent Card 声明与发现 | Card 注册 + DNS 发现 |
| A2A Bus（Redis）集成 | 消息总线配置 |
| 评审角色 Agent 容器 | `deploy/docker/agent/` |
| 容器编排 Compose 完善 | 全角色容器联调 |
| A2A 消息审计 | `audit/a2a-messages.jsonl` |

**关键里程碑**：PM/架构/QA/安全/DevOps Agent 容器并行运行，A2A 通信闭环。

### Phase 4：质量门禁与治理（3 周）

| 任务 | 产出 |
|---|---|
| Quality-Gate 引擎 | `internal/qualitygate/` |
| 113+ 治理规则 | `internal/governance/rules.go` |
| 前后端契约交叉校验 | OpenAPI 解析 + 调用比对 |
| Runtime 探针 | 启动应用 + 路由访问 |
| fail-open 容错 | 治理引擎异常处理 |
| 合规映射 | SOC-2/ISO27001/EU-AI-Act |

**关键里程碑**：质量门禁可阻塞不合格交付，生成 quality-gate.json + scorecard。

### Phase 5：知识库与记忆（2-3 周）

| 任务 | 产出 |
|---|---|
| BM25 检索引擎（纯 Go） | `internal/knowledge/bm25/` |
| HyDE 查询扩展 | 假想文档生成 |
| RRF 融合 | 双路召回排序 |
| 本地向量模型集成 | ONNX Runtime + e5-small |
| repo-map 符号索引 | 源码符号扫描 + 排序 |
| 项目记忆库 | facts/recipes/pitfalls/lessons |
| 云端 embedding 安全限制 | 双环境变量开关 |

### Phase 6：对外协议与 TUI（3-4 周）

| 任务 | 产出 |
|---|---|
| MCP Server 实现 | `internal/mcp/` |
| ACP Server 实现 | `internal/acp/` |
| A2A Server 对外暴露 | Agent Card 发现 + 跨实例互联 |
| gRPC API 定义 | `pkg/api/` proto |
| TypeScript TUI 客户端 | Ink + gRPC-Web + Slash 命令 |
| TUI 实时状态展示 | 任务进度/Agent 状态/审计流 |
| npm 包发布 | `aicodingaicodingagentteam-tui` |

**关键里程碑**：TUI 可交互式驱动完整流水线，外部 MCP 客户端可调用。

### Phase 7：集成测试与发布（2 周）

| 任务 | 产出 |
|---|---|
| 端到端集成测试 | 绿场项目全链路验证 |
| Quick Edit 流程测试 | 轻量路径验证 |
| 治理 fail-open 测试 | 故障注入 |
| 多宿主切换测试 | 4 款 CLI 切换 |
| Docker 镜像发布 | GHCR / Docker Hub |
| npm TUI 发布 | npm registry |
| 文档完善 | README + 快速上手 |

---

## 6. 开发里程碑

```
Phase 0  ████░░░░░░░░░░░░  项目初始化        (W1-2)
Phase 1  ████████░░░░░░░░  宿主驱动层        (W3-6)
Phase 2  ██████████░░░░░░  编排核心引擎      (W5-10)
Phase 3  ████████░░░░░░░░  A2A+容器化       (W9-13)
Phase 4  ██████░░░░░░░░░░  质量门禁治理      (W12-15)
Phase 5  ██████░░░░░░░░░░  知识库记忆       (W15-18)
Phase 6  ████████░░░░░░░░  对外协议+TUI     (W17-21)
Phase 7  ████░░░░░░░░░░░░  集成测试发布     (W21-23)
                                    总计 ~23 周
```

> Phase 1-3 可部分并行（宿主驱动与编排引擎有依赖但知识库/TUI 可提前启动）。

---

## 7. 风险与对策

| 风险 | 等级 | 对策 |
|---|---|---|
| 外部 CLI 协议变更 | 高 | 抽象 Runtime trait，驱动层隔离变更 |
| A2A 跨容器网络延迟 | 中 | Redis 本地总线 + 超时降级 + 工件文件兜底 |
| 容器资源占用高 | 中 | 评审 Agent 按需启动，闲置自动缩容 |
| 本地向量模型 224MB | 低 | 可选 feature，默认纯 BM25 |
| NFS/SMB 锁不可靠 | 中 | 单写者模型 + 文件锁 + 单容器写 |
| 治理引擎误判阻断 | 中 | fail-open 设计 + 可配置跳过 |
| CLI 鉴权失败 | 中 | AuthStatus 预检，不持有密钥 |
| 模型输出伪造完成 | 高 | 确定性校验 + 指纹快照 + Runtime 探针 |

---

## 8. 源码阅读与开发路径

### 8.1 推荐阅读顺序

1. `internal/types/`：共享类型定义，理解数据模型；
2. `pkg/runtime/runtime.go`：Runtime trait 接口，理解宿主抽象；
3. `internal/router/`：意图路由逻辑；
4. `internal/coordinator/director.go`：核心调度循环；
5. `internal/scheduler/`：角色调度与 A2A 分发；
6. `internal/host/`：4 款宿主驱动实现；
7. `internal/qualitygate/`：质量门禁引擎；
8. `internal/governance/rules.go`：治理规则；
9. `internal/a2a/`：A2A 协议实现；
10. `cmd/aicodingaicodingagentteam/main.go`：二进制入口。

### 8.2 扩展开发点

#### 新增宿主 Backend
1. 在 `internal/host/` 新增驱动包，实现 `Runtime` 接口；
2. 处理子进程生命周期、消息协议转换；
3. 适配会话、工具调用、事件映射；
4. 更新配置常量，增加 Backend ID；
5. 新增 `deploy/docker/host-{name}/` Dockerfile；
6. CLI/TUI 命令增加新后端选项。

> 必须处理宿主能力差异，不具备的能力返回不可用，禁止模拟伪造。

#### 新增治理规则
1. 在 `internal/governance/rules.go` 新增规则结构；
2. 实现检查逻辑；
3. 配置 `rules.toml` 默认启用状态；
4. 添加路径排除支持。

#### 新增角色
1. 在 `internal/types/` 定义角色枚举；
2. 创建 Agent 容器配置 `deploy/docker/agent-{role}/`；
3. 实现 A2A Agent Card 声明；
4. Scheduler 中注册调度策略。
