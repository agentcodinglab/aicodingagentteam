# 技术架构设计文档 — AiCodingAgentTeam

> 项目名称：AiCodingAgentTeam
> 语言/平台：Golang + Docker
> 文档版本：v1.0
> 编写日期：2026-09-02

---

## 目录
1. [架构总览](#1-架构总览)
2. [分层架构详解](#2-分层架构详解)
3. [Go 模块划分](#3-go-模块划分)
4. [容器化部署架构](#4-容器化部署架构)
5. [A2A 协议设计](#5-a2a-协议设计)
6. [宿主驱动层](#6-宿主驱动层)
7. [角色编排引擎](#7-角色编排引擎)
8. [质量门禁与治理](#8-质量门禁与治理)
9. [知识库与检索](#9-知识库与检索)
10. [数据存储设计](#10-数据存储设计)
11. [对外协议层（MCP/ACP/A2A）](#11-对外协议层)
12. [TUI 客户端架构](#12-tui-客户端架构)
13. [安全设计](#13-安全设计)
14. [可观测性](#14-可观测性)

---

## 1. 架构总览

### 1.1 高层架构图

```
┌─────────────────────────────────────────────────────────┐
│                    外部调用方                             │
│  TS TUI 客户端 │ MCP 客户端 │ ACP 客户端 │ 外部 A2A Agent  │
└───────┬──────────┬──────────┬──────────┬────────────────┘
        │ gRPC/WS  │ MCP      │ ACP      │ A2A
┌───────▼──────────▼──────────▼──────────▼────────────────┐
│                  API Gateway / 协议适配层                   │
│          (MCP Server │ ACP Server │ A2A Server)           │
└───────────────────────┬─────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────┐
│                   Coordinator（协调器）                      │
│  意图路由 → 计划构建(DAG) → 调度执行 → 质量校验 → 交付打包    │
└───┬───────────┬───────────┬───────────┬───────────┬────┘
    │ A2A       │ A2A       │ A2A       │ A2A       │
┌───▼──┐  ┌────▼───┐  ┌────▼───┐  ┌────▼───┐  ┌────▼───┐
│ PM   │  │Architect│  │ UIUX  │  │  QA    │  │Security │  ← 评审角色（并行容器）
│Agent │  │ Agent  │  │ Agent │  │ Agent  │  │ Agent  │
└──────┘  └────────┘  └───────┘  └────────┘  └────────┘
    │ A2A                                      │ A2A
┌───▼───────────────────────────┐  ┌──────────▼──────┐
│   Frontend/Backend (写角色)     │  │   DevOps Agent  │
│   串行执行，单写者模型           │  │  部署证据收集    │
└───────────┬───────────────────┘  └─────────────────┘
            │ 调用
┌───────────▼───────────────────────────────────────────┐
│              Host Driver Layer（宿主驱动层）              │
│  Claude-Code │ Codex │ OpenCode │ DeepSeek-DSH        │
│  (各自容器/子进程，各协议适配)                            │
└───────────────────────────────────────────────────────┘
            │ 读写
┌───────────▼───────────────────────────────────────────┐
│         共享工件卷（Docker Volume / Bind Mount）           │
│  output/ .aicodingaicodingaicodingagentteam/ contracts/ audit/ memory/ src/    │
└───────────────────────────────────────────────────────┘
```

### 1.2 架构原则

1. **编排与执行分离**：Coordinator 只编排不写代码，宿主 CLI 只执行不决策；
2. **容器即角色**：每个角色 Agent 独立容器，可独立伸缩/替换；
3. **协议即契约**：Agent 间 A2A 通信，对外 MCP/ACP/A2A 三协议；
4. **工件即通信**：角色间通过文件 + 结构化 Verdict 通信，禁止自由对话；
5. **确定性优先**：质量校验机器硬执行，不依赖模型自评；
6. **本地优先**：代码/文档默认不出容器，云端能力需显式开启。

---

## 2. 分层架构详解

### 2.1 五层执行流转模型

```
用户请求
  │
  ▼
① Route 意图路由层
  │  判定: Chat / Explain / QuickEdit / Debug / Build
  │  评估: 深度、写权限、范围；必要时向用户澄清
  ▼
② Plan 计划构建层
  │  Build 类型 → 生成 DAG 任务图 → plan.json
  │  用户可 /plan 干预（增删/调整顺序）
  ▼
③ Schedule 调度执行层
  │  写角色(前端/后端): 串行执行，单写者
  │  评审角色(PM/架构/QA/安全): 并行 A2A 子任务
  │  通信: 工件文件 + RoleVerdict 结构化结论
  ▼
④ Verify 自校正验证层
  │  确定性校验: 构建/测试/lint/契约/安全扫描
  │  阻塞项生成修复方案；指纹快照防无限循环
  ▼
⑤ Finalize 交付产物层
  │  输出对应深度工件
  │  更新项目记忆库；生成审计证据包
```

### 2.2 各层职责矩阵

| 层 | 输入 | 输出 | 关键组件 |
|---|---|---|---|
| Route | 用户消息 | IntentType + 深度评估 | Router、Clarifier |
| Plan | IntentType + 上下文 | plan.json (DAG) | Planner、DAGBuilder |
| Schedule | plan.json | 各角色工件 + Verdict | Coordinator、A2A Bus |
| Verify | 工件 + 源码 | quality-gate.json | QualityGate、RuntimeProbe |
| Finalize | 校验结果 | 证据包 + 记忆更新 | Delivery、MemoryWriter |

---

## 3. Go 模块划分

采用 Go Module 多包结构，单一仓库管理：

```
aicodingaicodingagentteam/
├── cmd/
│   ├── aicodingaicodingagentteam/          # 主二进制入口
│   └── tui/                # TS TUI（独立 npm 包，此为 Go 端 gRPC stub 不需要）
├── internal/
│   ├── coordinator/       # 协调器：路由→计划→调度→验证→交付
│   ├── router/            # 意图路由
│   ├── planner/           # DAG 计划构建
│   ├── scheduler/         # 角色调度（A2A 任务分发）
│   ├── qualitygate/       # 质量门禁引擎
│   ├── governance/        # 治理规则引擎（113+ 规则）
│   ├── knowledge/         # BM25 + 向量混合检索
│   ├── memory/            # 项目记忆（facts/recipes/pitfalls/lessons）
│   ├── host/              # 宿主驱动（4 款 CLI 适配）
│   │   ├── claude/        # Claude-Code 驱动
│   │   ├── codex/         # Codex JSON-RPC 驱动
│   │   ├── opencode/      # OpenCode ACP stdio JSON-RPC 驱动 (ADR-0021)
│   │   └── dsh/           # DeepSeek-DSH 驱动
│   ├── a2a/               # A2A 协议（Agent Card / RPC / Bus）
│   ├── mcp/               # MCP Server 适配
│   ├── acp/               # ACP Server 适配
│   ├── audit/             # 审计日志
│   ├── config/            # 配置加载
│   └── types/             # 共享类型定义
├── pkg/
│   ├── api/               # 对外 gRPC/HTTP API
│   ├── runtime/           # Runtime trait 接口定义
│   └── contracts/         # OpenAPI 契约解析与校验
├── deploy/
│   ├── docker/            # Dockerfile 集合
│   ├── compose/          # docker-compose.yml
│   └── k8s/               # K8s 清单（可选）
├── tui/                   # TypeScript TUI 客户端源码
├── go.mod
└── go.sum
```

### 3.1 模块依赖关系

```
cmd/aicodingagentteam
  └→ internal/coordinator
       ├→ internal/router
       ├→ internal/planner
       ├→ internal/scheduler ──→ internal/a2a
       ├→ internal/qualitygate
       ├→ internal/governance
       ├→ internal/knowledge
       ├→ internal/memory
       └→ internal/host ──→ pkg/runtime
  └→ pkg/api
       ├→ internal/mcp
       ├→ internal/acp
       └→ internal/a2a
```

### 3.2 核心接口定义（Go）

```go
// pkg/runtime/runtime.go — 宿主驱动抽象接口
package runtime

type Runtime interface {
    // 生命周期
    StartSession(ctx context.Context, opts SessionOpts) (SessionID, error)
    DestroySession(ctx context.Context, id SessionID) error

    // 任务执行
    SendTask(ctx context.Context, id SessionID, task TaskPayload) (<-chan Event, error)

    // 能力查询
    Capabilities() HostCapabilities
    ModelInfo() ModelInfo

    // 控制流
    Pause(ctx context.Context, id SessionID) error
    Resume(ctx context.Context, id SessionID) error

    // 鉴权状态（不持有密钥，只查询状态）
    AuthStatus(ctx context.Context, id SessionID) AuthStatus
}
```

```go
// internal/a2a/agent.go — A2A Agent 接口
package a2a

type Agent interface {
    Card() AgentCard                       // 能力声明
    Execute(ctx context.Context, task Task) (Result, error)
    Status(ctx context.Context) AgentStatus
}

type AgentCard struct {
    ID          string
    Name        string
    Role        Role           // pm/architect/frontend/...
    Capabilities []string
    Endpoint    string         // A2A RPC 地址
}
```

```go
// internal/coordinator/director.go — 核心调度循环
type Director struct {
    router   *router.Router
    planner  *planner.Planner
    sched    *scheduler.Scheduler
    gate     *qualitygate.Engine
    memory   *memory.Store
}

func (d *Director) Handle(ctx context.Context, req UserRequest) (*Delivery, error) {
    intent := d.router.Route(ctx, req)
    plan := d.planner.Build(ctx, intent)
    artifacts := d.sched.Execute(ctx, plan)   // A2A 调度
    verdict := d.gate.Verify(ctx, artifacts)  // 确定性校验
    return d.finalize(ctx, artifacts, verdict)
}
```


---

## 4. 容器化部署架构

### 4.1 容器拓扑

```
┌──────────────────────────────────────────────┐
│              Docker Compose 网络                │
│                                               │
│  ┌──────────────┐    ┌──────────────┐        │
│  │ coordinator  │    │  host-claude │        │
│  │  (主控容器)   │    │  (CLI 容器)  │        │
│  │  :8080 gRPC  │    │              │        │
│  │  :8081 MCP   │    ├──────────────┤        │
│  │  :8082 ACP   │    │  host-codex  │        │
│  │  :8083 A2A   │    │  (CLI 容器)  │        │
│  └──────┬───────┘    ├──────────────┤        │
│         │            │ host-opencode│        │
│  ┌──────▼───────┐    │  (CLI 容器)   │        │
│  │  a2a-bus     │    ├──────────────┤        │
│  │ (消息总线)    │    │  host-dsh    │        │
│  └──┬──┬──┬──┬──┘    │  (CLI 容器)   │        │
│     │  │  │  │       └──────────────┘        │
│  ┌──▼┐│┌─▼┐┌─▼┐                            │
│  │PM │││Ar││QA│  ← 评审角色 Agent 容器池      │
│  └───┘│└──┘└──┘                            │
│       │                                     │
│  ┌────▼─────────────────────────────┐        │
│  │  Shared Volume (工件卷)          │        │
│  │  /workspace/output/             │        │
│  │  /workspace/.aicodingaicodingaicodingagentteam/          │        │
│  │  /workspace/src/  (用户源码)      │        │
│  └─────────────────────────────────┘        │
└──────────────────────────────────────────────┘
```

### 4.2 Docker Compose 设计

```yaml
# deploy/compose/docker-compose.yml
version: "3.9"
services:
  coordinator:
    build: ../../
    ports:
      - "8080:8080"  # gRPC (TUI)
      - "8081:8081"  # MCP
      - "8082:8082"  # ACP
      - "8083:8083"  # A2A
    environment:
      - AICODINGAGENTTEAM_A2A_BUS=redis://a2a-bus:6379
      - AICODINGAGENTTEAM_WORKSPACE=/workspace
    volumes:
      - workspace:/workspace
    depends_on: [a2a-bus]

  a2a-bus:
    image: redis:7-alpine
    ports: ["6379:6379"]

  host-claude:
    build: ../../deploy/docker/host-claude
    volumes: [workspace:/workspace:rw]

  host-codex:
    build: ../../deploy/docker/host-codex
    volumes: [workspace:/workspace:rw]

  host-opencode:
    build: ../../deploy/docker/host-opencode
    volumes: [workspace:/workspace:rw]

  host-dsh:
    build: ../../deploy/docker/host-dsh
    volumes: [workspace:/workspace:rw]

  # 评审角色 Agent（可 replicas 扩展）
  agent-pm:
    build: ../../deploy/docker/agent
    environment:
      - AGENT_ROLE=pm
      - AICODINGAGENTTEAM_A2A_BUS=redis://a2a-bus:6379
    volumes: [workspace:/workspace:rw]
    deploy:
      replicas: 1  # PM 单实例

  agent-architect:
    build: ../../deploy/docker/agent
    environment:
      - AGENT_ROLE=architect
    volumes: [workspace:/workspace:rw]

  agent-qa:
    build: ../../deploy/docker/agent
    environment:
      - AGENT_ROLE=qa
    volumes: [workspace:/workspace:rw]

  agent-security:
    build: ../../deploy/docker/agent
    environment:
      - AGENT_ROLE=security
    volumes: [workspace:/workspace:rw]

  agent-devops:
    build: ../../deploy/docker/agent
    environment:
      - AGENT_ROLE=devops
    volumes: [workspace:/workspace:rw]

volumes:
  workspace:
```

### 4.3 容器分层设计

| 镜像层 | 用途 | 扩展性 |
|---|---|---|
| `aicodingaicodingagentteam-base` | Go 运行时 + 公共库 | — |
| `coordinator` | 主控编排引擎 | 单实例（可 HA） |
| `agent-{role}` | 评审角色 Agent | 水平扩展（QA/安全等） |
| `host-{cli}` | AI 编码 CLI 驱动 | 按需启动 |
| `tui` | TS TUI 客户端 | 客户端本地运行 |


---

## 5. A2A 协议设计

### 5.1 协议概述

AiCodingAgentTeam 的 Agent 间协作基于 **A2A（Agent-to-Agent）协议**，采用 JSON-RPC over HTTP 或 Redis Pub/Sub 实现：

- **Agent Card**：每个 Agent 暴露能力声明（角色、能力、端点）；
- **Task Delegation**：Coordinator 委派任务给专业 Agent；
- **Result Return**：Agent 返回结构化 Result + Verdict；
- **Status Query**：实时查询 Agent 执行状态。

### 5.2 A2A 消息格式

```json
{
  "jsonrpc": "2.0",
  "id": "task-uuid",
  "method": "agent.execute",
  "params": {
    "task_id": "dag-node-3",
    "role": "qa",
    "payload": {
      "artifacts": ["output/app-prd.md", "output/app-architecture.md"],
      "source_paths": ["src/"],
      "instruction": "基于 PRD 验收标准编写测试用例"
    },
    "deadline": "2026-09-02T12:00:00Z"
  }
}
```

### 5.3 Agent Card 声明

```json
{
  "id": "agent-qa-001",
  "name": "QA Engineer",
  "role": "qa",
  "capabilities": ["test-generation", "runtime-probe", "coverage-analysis"],
  "endpoint": "http://agent-qa:9090/a2a",
  "max_concurrent": 1,
  "timeout_default": 300
}
```

### 5.4 A2A 通信模式

```
Coordinator                    Agent (QA)
    │                             │
    │── agent.execute(task) ─────▶│
    │                             │ (处理中)
    │◀── agent.progress(event) ───│
    │                             │
    │◀── agent.result(verdict) ───│
    │                             │
```

**关键约束**：
1. Agent 间**禁止直接对话**，所有通信经 Coordinator 路由；
2. Agent 返回的 Verdict 是**结构化**的（accept/blocking/advisory）；
3. 评审 Agent 失败（超时/解析失败）**不伪造成功**，标记 park 暂停；
4. A2A Bus（Redis）作为消息中间件，解耦 Coordinator 与 Agent 容器。

### 5.5 Verdict 结构

```json
{
  "task_id": "dag-node-3",
  "role": "qa",
  "verdict": "blocking",
  "severity": "critical",
  "findings": [
    {
      "check": "test-coverage",
      "status": "fail",
      "detail": "覆盖率 65%，阈值 80%",
      "evidence": "coverage-report.json"
    }
  ],
  "artifacts_produced": ["output/app-test-report.md"]
}
```

---

## 6. 宿主驱动层

### 6.1 四款宿主 CLI 适配

| 宿主 | 协议 | 容器 | 说明 |
|---|---|---|---|
| Claude-Code | 私有流协议 | `host-claude` | stdio 流式交互 |
| Codex | JSON-RPC | `host-codex` | OpenAI Codex CLI 协议 |
| OpenCode | ACP stdio JSON-RPC | `host-opencode` | notifications/session/update 流 (ADR-0021) |
| DeepSeek-DSH | ACP v1 / 私有 | `host-dsh` | DeepSeek CLI 驱动 |

### 6.2 Runtime Trait（Go 接口）

```go
// 每个宿主驱动实现此接口
type Runtime interface {
    StartSession(ctx context.Context, opts SessionOpts) (SessionID, error)
    DestroySession(ctx context.Context, id SessionID) error
    SendTask(ctx context.Context, id SessionID, task TaskPayload) (<-chan Event, error)
    Capabilities() HostCapabilities
    ModelInfo() ModelInfo
    Pause(ctx context.Context, id SessionID) error
    Resume(ctx context.Context, id SessionID) error
    AuthStatus(ctx context.Context, id SessionID) AuthStatus
}
```

### 6.3 能力差异处理

**核心原则：不抹平能力差异，不具备的能力直接报告，禁止模拟伪造。**

| 能力 | Claude-Code | Codex | OpenCode | DSH |
|---|---|---|---|---|
| 会话 Resume | ✓ | ✓ | ✗ | ✓ |
| 工具调用事件 | ✓ | ✓ | ✓ | ✓ |
| 联网搜索 | ✓ | ✓ | ✗ | ✓ |
| 文件写入钩子 | ✓ | ✗ | ✓ | ✗ |

> 当某宿主不支持 Resume 时，Coordinator 自动新建会话并移交上下文工件，而非伪造恢复。

### 6.4 鉴权策略

- AiCodingAgentTeam **不存储任何 API Key / 凭证**；
- 每个宿主容器自行管理鉴权（环境变量 / 配置文件挂载）；
- Coordinator 仅查询 `AuthStatus` 判断是否就绪。


---

## 7. 角色编排引擎

### 7.1 调度模型

```
plan.json (DAG)
    │
    ▼
┌─────────────────────────────────────────┐
│            Scheduler                       │
│                                           │
│  ┌─────────┐  写角色串行队列               │
│  │ Writer  │──────────────────▶ host-cli   │
│  │ Queue   │  (前端→后端 顺序执行)         │
│  └─────────┘                              │
│                                           │
│  ┌─────────┐  评审角色并行池               │
│  │ Review  │──┬─▶ PM Agent    (A2A)       │
│  │ Pool    │  ├─▶ Architect    (A2A)       │
│  │         │  ├─▶ QA           (A2A)       │
│  │         │  ├─▶ Security     (A2A)       │
│  │         │  └─▶ DevOps       (A2A)      │
│  └─────────┘                              │
│                                           │
│  汇聚: 收集所有 Verdict → Coordinator 判定  │
└─────────────────────────────────────────┘
```

### 7.2 DAG 任务节点结构

```json
{
  "nodes": [
    {
      "id": "n1-clarify",
      "phase": "clarify",
      "role": "coordinator",
      "depends_on": [],
      "artifacts_out": ["output/app-clarify.md"]
    },
    {
      "id": "n3-prd",
      "phase": "docs",
      "role": "pm",
      "depends_on": ["n1-clarify", "n2-research"],
      "artifacts_in": ["output/app-clarify.md", "output/app-research.md"],
      "artifacts_out": ["output/app-prd.md"]
    },
    {
      "id": "n6-frontend",
      "phase": "frontend",
      "role": "frontend",
      "depends_on": ["n5-spec", "n3-prd", "n4-arch"],
      "artifacts_in": ["output/app-architecture.md", ".aicodingaicodingaicodingagentteam/contracts/openapi.json"],
      "artifacts_out": ["src/frontend/"],
      "writer": true
    }
  ],
  "gates": [
    { "id": "g1-docs", "after": "n4-arch", "type": "human" },
    { "id": "g2-preview", "after": "n6-frontend", "type": "human" },
    { "id": "g3-quality", "after": "n7-backend", "type": "auto" }
  ]
}
```

### 7.3 规模自适应

| 任务规模 | 启动角色 | 流程 |
|---|---|---|
| QuickEdit（小修改） | 无（直接 host） | 轻量治理校验 |
| Feature（功能增强） | Frontend/Backend + QA | 裁剪流水线 |
| Build（绿场项目） | 全部 9 角色 | 完整 9 阶段 |

判定依据：需求复杂度评分 + 路由器意图分类 + 用户显式指定。

---

## 8. 质量门禁与治理

### 8.1 Quality-Gate 校验清单

```
┌─────────────────────────────────────────┐
│         Quality-Gate Engine              │
│                                         │
│  ① PRD 完整性校验                        │
│  ② 架构/API/数据模型校验                  │
│  ③ 前端 API 调用 vs OpenAPI 契约交叉校验  │
│  ④ UI 坏味道检测 (emoji/硬编码颜色)       │
│  ⑤ 编译 / 测试 / lint / 类型检查          │
│  ⑥ Dockerfile / CI / .env.example       │
│  ⑦ 密钥泄露扫描                          │
│  ⑧ Runtime 探针 (启动应用访问路由)        │
│  ⑨ 审计日志完整性                         │
│  ⑩ 合规映射 (SOC-2/ISO27001/EU-AI-Act)   │
│                                         │
│  输出: quality-gate.json + scorecard    │
└─────────────────────────────────────────┘
```

### 8.2 治理规则引擎

- **113+ 条治理规则**，覆盖 UI、安全、前后端工程风险；
- 每条规则可独立配置：启用/关闭、路径排除、严重级别；
- **fail-open 设计**：治理引擎内部异常返回通过，不阻断开发。

```toml
# .aicodingaicodingaicodingagentteam/rules.toml
[disabled]
clauses = ["ui-emoji-icon"]

[exclusions]
paths = ["src/legacy/**", "**/*.test.ts"]

[quality]
threshold = 90
skip_checks = ["runtime-probe"]
```

### 8.3 治理触发入口

| 入口 | 时机 |
|---|---|
| Pre-write 钩子 | 代码写入文件前拦截 |
| CI / Pre-commit | 提交前扫描 |
| MCP govern_file | 外部客户端调用 |
| Quality-Gate | 交付前全套扫描 |

---

## 9. 知识库与检索

### 9.1 混合检索架构

```
用户需求 + 当前阶段
        │
        ▼
    HyDE 查询扩展（生成假想文档）
        │
   ┌────┴────┐
   ▼         ▼
 路径A      路径B
 BM25      向量检索
(CJK       (本地 e5-small
 bigram)    或云端 embedding)
   │         │
   └────┬────┘
        ▼
    RRF 融合排序
        │
        ▼
    Top-K 块 → 注入宿主上下文
```

### 9.2 安全限制

- BM25 是**保底必选**，纯 Go 实现，零外部依赖；
- 云端 embedding 需双环境变量：`OPENAI_EMBED_KEY` + `AICODINGAGENTTEAM_ALLOW_CLOUD_EMBED=1`；
- 向量不可用时自动降级纯 BM25，不报错。

### 9.3 repo-map 符号索引

- 扫描项目源码符号，计算重要度排序；
- 做上下文预算裁剪，优先注入高重要度符号。


---

## 10. 数据存储设计

### 10.1 工作目录结构

```
your-project/
├── output/                      # 产出文档
│   ├─ *-clarify.md
│   ├─ *-research.md
│   ├─ *-prd.md
│   ├─ *-architecture.md
│   ├─ *-uiux.md
│   ├─ *-execution-plan.md
│   ├─ *-quality-gate.md / .json
│   └─ ...
├── .aicodingaicodingaicodingagentteam/
│   ├─ plan.json                 # DAG 任务计划
│   ├─ workflow-state.json       # 运行时状态
│   ├─ rules.toml                # 治理规则配置
│   ├─ contracts/
│   │   └─ openapi.json/yaml     # API 契约
│   ├─ audit/
│   │   ├─ tool-calls.jsonl      # 工具调用审计
│   │   ├─ verify.jsonl          # 校验审计
│   │   ├─ a2a-messages.jsonl    # A2A 消息审计
│   │   └─ frontend-api-calls.jsonl
│   ├─ memory/
│   │   ├─ facts.jsonl           # 项目事实
│   │   ├─ recipes.jsonl         # 历史方案
│   │   ├─ learned-skills/       # 经验规则
│   │   └─ dev-errors.jsonl      # 失败事件
│   ├─ proof/                    # 证据包
│   │   ├─ runtime-proof.json
│   │   └─ deploy-proof.json
│   └─ run-notes.md             # 运行时笔记
└── src/                         # 用户源码
```

### 10.2 审计日志格式

```json
{"ts":"2026-09-02T10:00:00Z","type":"tool_call","agent":"qa","task":"n8","tool":"run_tests","result":"pass","duration_ms":3200}
{"ts":"2026-09-02T10:05:00Z","type":"a2a_message","from":"coordinator","to":"agent-architect","method":"agent.execute","status":"ok"}
{"ts":"2026-09-02T10:10:00Z","type":"verify","check":"api-contract","status":"fail","detail":"路径不匹配 /api/v1/users","severity":"blocking"}
```

### 10.3 容器卷映射

| 宿主路径 | 容器内 | 说明 |
|---|---|---|
| `{workspace}/` | `/workspace` | 共享工件卷（读写） |
| `~/.aicodingaicodingaicodingagentteam/config.toml` | `/config/config.toml` | 全局配置（只读） |

---

## 11. 对外协议层

### 11.1 三协议架构

```
┌─────────────────────────────────────────────┐
│              Protocol Gateway                 │
│                                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │ MCP Server│  │ ACP Server│  │ A2A Server│  │
│  │  :8081    │  │  :8082    │  │  :8083    │  │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  │
│        │              │              │        │
│        └──────────────┼──────────────┘        │
│                       ▼                       │
│              Coordinator API                   │
└───────────────────────────────────────────────┘
```

### 11.2 MCP Server

`aicodingagentteam mcp serve` 暴露治理与编排能力：

| MCP Tool | 功能 |
|---|---|
| `govern_file` | 对文件执行治理扫描 |
| `run_pipeline` | 触发完整流水线 |
| `quick_edit` | 轻量快速编辑 |
| `verify` | 执行质量校验 |
| `get_plan` | 获取当前 DAG 计划 |
| `query_memory` | 查询项目记忆 |

### 11.3 ACP Server

Agent Client Protocol，对接标准 Agent 客户端：
- 会话管理（创建/恢复/销毁）；
- 消息收发；
- 工具调用事件转发。

### 11.4 A2A Server

对外暴露 A2A 接口，支持：
- 外部 AiCodingAgentTeam 实例互联；
- Agent Card 发现（`/.well-known/agent.json`）；
- 跨团队任务委派。

---

## 12. TUI 客户端架构

### 12.1 技术栈

```
┌─────────────────────────────┐
│     TypeScript TUI Client    │
│  (npm 包，终端运行)            │
│                              │
│  ┌─────────┐  ┌───────────┐ │
│  │ Ink/     │  │ gRPC-Web  │ │
│  │ Blessed  │  │ /WebSocket│ │
│  │ 渲染层   │  │ 通信层    │ │
│  └────┬────┘  └─────┬─────┘ │
│       │             │       │
│  ┌────▼─────────────▼────┐  │
│  │   Slash 命令路由器     │  │
│  │ /run /quick /plan ... │  │
│  └───────────────────────┘  │
└─────────────────────────────┘
          │ gRPC / WebSocket
          ▼
   Coordinator :8080
```

### 12.2 技术选型

| 组件 | 选型 | 理由 |
|---|---|---|
| TUI 渲染 | Ink（React for CLI） | 生态成熟，TS 原生 |
| 通信协议 | gRPC-Web / WebSocket | 实时双向流式 |
| 状态管理 | Zustand | 轻量，适合 TUI |
| 分发 | npm `aicodingaicodingagentteam-tui` | 跨平台，`npx` 直接运行 |

### 12.3 Slash 命令

| 命令 | 功能 |
|---|---|
| `/run [需求]` | 启动完整流水线 |
| `/goal` | 目标模式 |
| `/quick [修改]` | 轻量快速编辑 |
| `/plan` | 查看/编辑 DAG 计划 |
| `/continue` | 继续暂停的任务 |
| `/revise` | 修订文档 |
| `/backend [name]` | 切换宿主 CLI |
| `/verify` | 执行质量校验 |
| `/memory` | 记忆管理 |
| `/knowledge` | 知识库管理 |

---

## 13. 安全设计

| 维度 | 策略 |
|---|---|
| 密钥管理 | 不持有任何 API Key；宿主容器自行鉴权 |
| 代码隐私 | 默认本地处理，不上传云端 |
| 云端 embedding | 需双环境变量显式开启 |
| 容器隔离 | 角色 Agent 容器隔离，网络命名空间 |
| 工件卷权限 | 读写隔离：评审角色只读，写角色读写 |
| 审计完整性 | 所有 A2A 消息 + 工具调用 jsonl 留痕 |

---

## 14. 可观测性

| 维度 | 实现 |
|---|---|
| 日志 | 结构化 JSON 日志（zerolog） |
| 指标 | Prometheus metrics（任务耗时、门禁通过率） |
| 追踪 | OpenTelemetry trace（Coordinator→Agent→Host 链路） |
| 审计 | `.aicodingaicodingaicodingagentteam/audit/*.jsonl` 持久化 |
| 证据包 | `proof-pack-*.zip` + `scorecard-*.html` 可交付 |
