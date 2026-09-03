# 领域模型与统一语言 — AiCodingAgentTeam

> 本文件记录项目核心业务术语与统一语言，确保人与 AI、角色间使用一致词汇。
> 范本：`E:\javaproject\my\2026\aicoding_docs\docs\domain.md`
> 遵循 `aicoding_docs/docs/writing-guide.md` 编写规范。

---

## 1. 术语表

| 术语 | 英文 | 含义 | 备注 |
|---|---|---|---|
| 编排 | Orchestration | Coordinator 负责的路由/计划/调度/验证/交付流程 | 平台核心能力 |
| 协调器 | Coordinator | 技术负责人角色，掌握计划与门禁，不写代码 | 主控逻辑 |
| 角色 | Role | 软件团队席位（PM/架构/UIUX/前端/后端/QA/安全/DevOps） | 9 席位 |
| 宿主 | Host | 被调度的 AI 编码 CLI（Claude-Code/Codex/OpenCode/DeepSeek-DSH） | 执行器 |
| 工件 | Artifact | 角色间通信与产出的文件（PRD/架构/UIUX/源码/报告） | 通信媒介 |
| Verdict | Verdict | 结构化评审结论（accept/blocking/advisory） | 评审角色输出 |
| A2A | Agent-to-Agent | Agent 间协作通信协议 | 跨容器通信 |
| ACP | Agent Client Protocol | Agent 客户端标准协议 | 对外接口 |
| MCP | Model Context Protocol | 模型上下文协议 | 对外接口 |
| DAG | Directed Acyclic Graph | 有向无环图任务计划 | 计划驱动 |
| Quality-Gate | Quality-Gate | 确定性质量门禁引擎 | 交付前校验 |
| 治理 | Governance | 113+ 规则的代码治理引擎 | fail-open |
| 门禁 | Gate | 流程中的人工/自动检查点 | docs_confirm/preview_confirm/quality |
| 记忆 | Memory | 项目本地事实/经验/方案存储 | 不跨项目共享 |
| 知识库 | Knowledge | BM25 + 向量混合检索 | 本地优先 |
| park | Park | 任务暂停状态，等待用户 /continue | 失败不伪造成功 |
| 写角色 | Writer Role | 可修改源码的角色（前端/后端），串行执行 | 单写者模型 |
| 评审角色 | Reviewer Role | 并行独立会话的评审角色 | 返回 Verdict |

---

## 2. 业务边界

### 2.1 解决的问题

为软件开发团队提供基于 AI 编码 CLI 的**可编排、可审计、带质量门禁的软件交付流水线**：
- 不替代大模型，而是调度外部 AI 编码 CLI；
- 模拟真实团队角色，按角色分工编排；
- 确定性校验保证交付质量，不依赖模型自评。

### 2.2 不解决的问题

- 不内置 LLM，不提供模型 API 端点；
- 不持有用户的 AI 编码 CLI 凭证；
- 不接管宿主 CLI 的鉴权与计费；
- 不替代人工需求澄清与最终验收决策。

---

## 3. 核心流程

### 3.1 完整交付流水线（绿场项目）

```
flowchart LR
  A[需求输入] --> B[clarify 澄清]
  B --> C[research 调研]
  C --> D[docs 文档产出]
  D --> G1{docs_confirm 门禁}
  G1 -->|确认| E[spec 计划]
  E --> F[frontend 前端开发]
  F --> G2{preview_confirm 门禁}
  G2 -->|确认| H[backend 后端开发]
  H --> I[quality 质量门禁]
  I --> J[delivery 交付打包]
  G1 -->|拒绝| B
  G2 -->|拒绝| F
```

### 3.2 轻量快速编辑

```
flowchart LR
  A[小修改需求] --> B[路由 QuickEdit]
  B --> C[直接调用宿主改文件]
  C --> D[轻量治理校验]
  D --> E[完成]
```

### 3.3 A2A 协作闭环

```
flowchart LR
  CO[Coordinator] -->|委派| BUS[A2A Bus]
  BUS --> PM[PM Agent]
  BUS --> AR[Architect Agent]
  BUS --> QA[QA Agent]
  PM -->|Verdict| BUS
  AR -->|Verdict| BUS
  QA -->|Verdict| BUS
  BUS -->|汇聚| CO
  CO -->|判定| DEC{全部 accept?}
  DEC -->|是| DEL[交付]
  DEC -->|有 blocking| PARK[park 暂停]
```

---

## 4. 聚合与实体

| 聚合根 | 包含实体 | 不变式（业务规则） |
|---|---|---|
| Plan（任务计划） | TaskNode, Gate | DAG 无环；写角色节点串行；门禁顺序不可跳过 |
| Session（会话） | AgentSession, HostSession | 同一时刻仅一个写角色修改源码 |
| Verdict（评审结论） | Finding, Evidence | blocking 必须有 severity；失败不伪造 accept |
| Memory（记忆库） | Fact, Recipe, Pitfall, Lesson | 规则须验证器确认后才生效；不跨项目共享 |

---

## 5. 状态机

### 5.1 任务节点状态

| 状态 | 流转 | 说明 |
|---|---|---|
| pending | → scheduled | 等待依赖完成 |
| scheduled | → running | Coordinator 调度中 |
| running | → completed / failed / parked | 执行中 |
| completed | → (下游 pending) | 产出工件供下游 |
| failed | → running（重试） / parked | 超过重试上限则 parked |
| parked | → running（用户 /continue） | 等待人工介入 |

### 5.2 流水线状态

| 状态 | 流转 |
|---|---|
| clarify → research → docs → docs_confirm → spec → frontend → preview_confirm → backend → quality → delivery |
