# 需求分析文档（PRD）— AiCodingAgentTeam

> 项目名称：AiCodingAgentTeam（AI 编码编排 AiCodingAgentTeam）
> 语言/平台：Golang + Docker 容器化
> 文档版本：v1.0
> 编写日期：2026-09-02
> 本文档精心原创，全部为项目自身需求描述与架构决策。

---

## 1. 产品定位

AiCodingAgentTeam 是一个 **基于 Golang 开发的 AI 编码编排平台**，核心定位为 **AI 软件交付流水线编排者（Orchestrator）**：

- **本身不拥有大模型**，不内置 LLM API 端点；
- **调度外部 AI 编码 CLI**（Claude-Code、Codex、OpenCode、DeepSeek-DSH）作为执行器；
- **模拟真实软件开发团队角色**（PM、架构师、前后端、QA、安全、DevOps），按角色编排任务；
- **Agent 间通过 A2A 协议协作**，对外暴露 MCP / ACP / A2A 三种协议接口；
- **容器化部署**，每个角色/宿主可独立伸缩；
- 提供 **TypeScript TUI 客户端**，面向软件开发者交互式使用。

> 比喻：AiCodingAgentTeam 是「导演 + 制片」，AI 编码 CLI 是「演员」，导演制定计划、分配角色、审查质量门禁，演员负责写代码。

---

## 2. 解决的行业痛点

| 痛点 | 现状 | AiCodingAgentTeam 方案 |
|---|---|---|
| 直接编码缺文档 | AI 拿需求就写代码，无 PRD/架构/验收标准 | 标准化流水线，先产出文档再编码 |
| API 契约不一致 | 前后端接口对不上 | 自动生成 OpenAPI 契约并交叉校验 |
| 劣质 AI 代码 | 硬编码颜色、emoji 图标、模板代码 | 确定性质量门禁机器硬校验 |
| 假完成 | TODO 占位、假数据标记完成 | 指纹快照 + 运行时探针验证 |
| 上下文丢失 | 迭代后历史决策遗忘 | 项目记忆库 + 知识检索 |
| 缺审计证据 | 无交付报告 | 完整审计日志 + 证据包 |
| 规范难注入 | 团队规范无法注入 AI | 治理规则配置 + 知识库注入 |

---

## 3. 与同领域方案的差异

| 维度 | 同领域方案（对照） | AiCodingAgentTeam（本项目） |
|---|---|---|
| 开发语言 | Rust | **Golang** |
| 调度 CLI | Claude-Code、Codex、OpenCode、Grok-Build、Kimi-Code（5款） | Claude-Code、Codex、OpenCode、DeepSeek-DSH（**4款**） |
| 部署模式 | 单二进制 | **容器化部署**（Docker/Compose） |
| Agent 协作 | 进程内结构化工件通信 | **A2A 协议**（跨容器/跨进程） |
| 对外协议 | MCP | **MCP + ACP + A2A** 三协议 |
| 客户端 | 内置 TUI（Rust） | **TypeScript TUI 客户端**（独立进程） |
| 分发 | npm shell + Rust 二进制 | Docker 镜像 + npm TUI 客户端 |

---

## 4. 核心功能需求

### 4.1 团队编排（角色模型）

9 个角色席位：8 个专业角色 + Coordinator 协调器

| 角色 | 产出工件 | 模式 |
|---|---|---|
| Product Manager | *-prd.md（用户故事、EARS 验收标准） | 评审角色（并行子会话） |
| Architect | *-architecture.md + openapi.* | 评审角色（并行子会话） |
| UI/UX Designer | *-uiux.md（设计 Token、组件状态） | 评审角色（并行子会话） |
| Frontend Engineer | 前端组件页面源码 | 写角色（串行） |
| Backend Engineer | 后端接口业务逻辑源码 | 写角色（串行） |
| QA Engineer | 测试用例 + runtime-proof.json | 评审角色（并行子会话） |
| Security Engineer | 威胁模型、SAST 扫描 | 评审角色（并行子会话） |
| DevOps | Dockerfile、CI 配置、deploy-proof.json | 评审角色（并行子会话） |
| **Coordinator** | 计划调度、门禁控制、审计日志 | 主控（不写代码） |

运行约束：
1. **写角色串行**：同一时刻仅一个写角色修改源码（单写者模型）；
2. **评审角色并行独立会话**：每个评审角色独立 Agent，不共享主会话历史；
3. **角色间禁止自由对话**：通信媒介仅限工件文件 + 结构化 Verdict；
4. **任务规模自适应**：小 bug 不拉起全团队，仅完整绿场构建启动全部角色。

### 4.2 计划驱动执行（DAG）

- 需求 → 意图路由 → 生成 DAG 依赖任务图 → 保存 `.aicodingaicodingaicodingagentteam/plan.json`；
- 用户可干预：`/plan` 查看、增删、调整任务顺序；
- 任务节点携带角色分配、依赖关系、验收标准。

### 4.3 确定性质量门禁（Quality-Gate）

不依赖模型自评，机器硬校验：
1. PRD 需求与验收标准完整性；
2. 架构 API/数据模型/鉴权设计；
3. 前后端 API 契约交叉校验（OpenAPI）；
4. UI 坏味道检测（emoji、硬编码颜色）；
5. 编译/测试/lint/类型检查；
6. Dockerfile/CI/.env.example 完整性；
7. 密钥泄露扫描；
8. Runtime 探针：启动应用访问路由生成 runtime-proof.json；
9. 审计日志完整性 + 合规映射（SOC-2 / ISO27001 / EU-AI-Act）。

### 4.4 Agent 间协作（A2A 协议）

- 基于 Agent-to-Agent (A2A) 协议实现跨 Agent 通信；
- 每个 Agent 暴露 Agent Card（能力声明）；
- 任务委派、结果回传、状态查询走 A2A RPC；
- Coordinator 作为 A2A Hub 调度各专业角色 Agent。

### 4.5 对外协议暴露

| 协议 | 用途 |
|---|---|
| MCP | 暴露治理/编排能力给 MCP 客户端（govern_file、run_pipeline 等） |
| ACP | Agent Client Protocol，对接标准 Agent 客户端 |
| A2A | Agent 间协作协议，支持外部 AiCodingAgentTeam 实例互联 |

### 4.6 知识库与检索

- BM25 必选（保底检索）；
- 向量检索可选（本地模型优先，云端需双环境变量开启）；
- RRF + HyDE 混合检索；
- repo-map 代码符号索引 + 上下文预算裁剪。

### 4.7 项目记忆

| 存储对象 | 位置 | 说明 |
|---|---|---|
| pitfalls | `.aicodingaicodingaicodingagentteam/learned/dev-errors.jsonl` | 失败事件库 |
| lessons | `.aicodingaicodingaicodingagentteam/memory/learned-skills/` | 经验证的规则 |
| facts | `.aicodingaicodingaicodingagentteam/memory/facts.jsonl` | 项目环境事实 |
| recipes | `.aicodingaicodingaicodingagentteam/memory/recipes.jsonl` | 历史交付方案 |

### 4.8 TypeScript TUI 客户端

- 独立 npm 包分发；
- Slash 命令交互（`/run /goal /quick /plan /continue /revise`）；
- 通过 gRPC/HTTP 连接 AiCodingAgentTeam 后端；
- 实时显示任务进度、Agent 状态、审计日志。

---

## 5. 业务流水线

### 5.1 完整交付流程（绿场项目）

```
clarify → research → docs(PRD/架构/UIUX) → docs_confirm【人工门禁】
→ spec(计划) → frontend → preview_confirm【预览门禁】
→ backend → quality(门禁) → delivery(交付打包)
```

### 5.2 轻量化快速编辑

`/quick` 命令 → 路由 QuickEdit → 直接宿主改文件 → 轻量治理校验 → 完成

### 5.3 治理子系统

- 规则库覆盖 UI、安全、工程风险（可配置关闭、路径排除）；
- **fail-open 设计**：治理引擎异常不阻断开发；
- 触发入口：Pre-write 钩子、CI/Pre-commit、MCP govern_file、Quality-Gate。

---

## 6. 非功能需求

| 维度 | 要求 |
|---|---|
| 容器化 | Docker/Compose，角色 Agent 独立容器 |
| 可伸缩 | 评审角色 Agent 可水平扩展 |
| 高可用 | Coordinator 无状态化，状态外置 |
| 安全 | 不持有密钥；鉴权交底层 CLI；本地优先不上传代码 |
| 可审计 | 关键操作留证据，产出可交付证据包 |
| 跨平台 | Linux/macOS/Windows（ARM 走 x64 兼容） |
| 性能 | 单任务超时可配（默认 300s），校验超时 120s |

---

## 7. 交付物规范

```
your-project/
├── output/                     # 产出文档
│   ├─ *-prd.md
│   ├─ *-architecture.md
│   ├─ *-uiux.md
│   ├─ *-execution-plan.md
│   ├─ *-quality-gate.md
│   └─ ...
├── .aicodingaicodingaicodingagentteam/
│   ├─ plan.json                 # DAG 任务计划
│   ├─ workflow-state.json       # 运行时状态
│   ├─ rules.toml                # 治理规则配置
│   ├─ contracts/openapi.json    # API 契约
│   ├─ audit/*.jsonl             # 审计日志
│   ├─ memory/                   # 记忆库
│   └─ proof/                    # 证据包
└── src/                         # 用户源码
```

---

## 8. 命令系统

### 8.1 TUI Slash 命令
- 流程控制：`/run /goal /quick /plan /continue /revise`
- 宿主切换、预览、检查、知识/记忆管理

### 8.2 CLI 子命令
```bash
aicodingagentteam init                  # 初始化项目
aicodingagentteam adopt                 # 导入存量项目
aicodingagentteam run "需求" --backend codex
aicodingagentteam quick "小修改"
aicodingagentteam verify --runtime
aicodingagentteam report
aicodingagentteam ci
aicodingagentteam mcp serve
aicodingagentteam a2a serve
aicodingagentteam knowledge-manage add ./docs
```

---

## 9. 关键设计原则

1. **模型只是工人，编排层掌握计划与验收**：不相信模型自评，重要校验机器硬执行；
2. **本地优先**：默认不上传用户代码/文档到云端，云端向量需双环境变量开启；
3. **fail-open 治理**：治理故障不阻断开发；
4. **工件通信**：角色间仅通过文件 + 结构化 Verdict，禁止自由对话，失败不伪造成功；
5. **规模自适应**：小任务轻量，大项目全团队；
6. **不持有密钥**：鉴权交底层 CLI；
7. **一切可审计**：关键操作留证据。

---

## 10. 限制与风险

1. 依赖外部 AI 编码 CLI 正常安装登录，平台不提供模型能力；
2. 容器化部署需要编排多个容器，资源占用高于单二进制；
3. A2A 协议跨容器通信引入网络延迟，需优化序列化与超时；
4. 本地向量模型约 224MB 磁盘空间；
5. NFS/SMB 网络文件系统锁语义不完全可靠；
6. 记忆学习机制是辅助，不能保证完全消除模型错误。

---

## 附录：术语表

| 术语 | 说明 |
|---|---|
| A2A | Agent-to-Agent Protocol，Agent 间协作通信协议 |
| ACP | Agent Client Protocol，Agent 客户端标准协议 |
| MCP | Model Context Protocol，模型上下文协议 |
| DAG | Directed Acyclic Graph，有向无环图任务计划 |
| RRF | Reciprocal Rank Fusion，倒数排名融合 |
| HyDE | Hypothetical Document Embedding，假想文档嵌入 |
| Verdict | 结构化评审结论（accept/blocking/advisory） |
| fail-open | 治理组件故障时默认放行，不阻断业务 |
