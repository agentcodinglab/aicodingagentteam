# AGENTS.md — AiCodingAgentTeam 项目开发规范总纲

> 本文件是 AiCodingAgentTeam 项目的**唯一权威入口**与全局规范。
> 本规范基于 Vibe Coding 规范总纲（`aicoding_docs/`）定制，结合本项目 Golang + 容器化 + A2A 协作 + TS TUI 的技术特征适配。
> 本文件为活文档，随项目演进持续更新；过时内容标注而非删除。

---

## 0. 如何使用本规范

### 0.1 规范来源

本项目的开发规范继承自 Vibe Coding 规范仓库，统一引用路径为：

```
E:\javaproject\my\2026\aicoding_docs\
├── AGENTS.md                         ← 规范总纲（范本）
├── docs\CONSTRAINTS.md               ← 质量红线
├── docs\domain.md                    ← 领域模型范本
├── docs\writing-guide.md             ← 统一文档编写规范
├── docs\standards\general.md         ← 通用开发规范（全栈）
├── docs\standards\languages\go.md    ← Go 开发规范（本项目后端主语言）
├── docs\standards\languages\typescript.md ← TypeScript 规范（TUI 客户端）
├── docs\standards\testing\testing.md ← 测试规范
├── docs\standards\security\security.md ← 安全规范
├── docs\standards\git\git-workflow.md ← Git 工作流
├── docs\templates\ADR.md             ← 架构决策记录模板
├── docs\templates\SPEC.md            ← 功能规格模板
└── docs\templates\PLAN.md            ← 实现计划模板
```

> 本项目在 `docs/` 目录下维护项目专属的 CONSTRAINTS.md、domain.md 及 adr/spec/plan 子目录，规范条款沿用 `aicoding_docs/` 范本。

### 0.2 快速应用

1. 阅读本 `AGENTS.md` 全文（必读）。
2. 按项目实际技术栈，激活 `aicoding_docs/docs/standards/` 下对应语言/框架规范（Go + TypeScript 强制激活）。
3. 在 `docs/CONSTRAINTS.md` 调整本项目质量阈值。
4. 在 `docs/domain.md` 写入业务术语与统一语言。
5. 团队 / AI 通读后即可开始 vibe coding。

### 0.3 规范优先级

AI 在处理任务时，按以下优先级合并指令（高 → 低）：

1. **用户当轮指令**：本轮对话中用户直接给出的要求，优先级最高。
2. **AGENTS.md（本文件）**：全局规范与协作流程总纲。
3. **`docs/CONSTRAINTS.md`**：质量红线，不可静默降级。
4. **`aicoding_docs/docs/standards/*`**：分层开发规范（通用/语言/框架）。
5. **`docs/domain.md`**：统一语言与业务术语。
6. **`docs/adr/*`**：历史决策记录，解释「为什么这样做」。
7. **`docs/spec/*`**：功能规格与验收标准。
8. **代码自身**：以上均未覆盖时，以现有代码风格与约定为准。

> 冲突规则：当文档与代码冲突时，以文档为准并向用户确认是否修正代码或更新文档。深层目录的规范覆盖浅层（如 `languages/go.md` 细节优先于 `general.md` 通用原则）。

---

## 1. 文档编写规范

所有项目文档（spec/plan/adr/domain 等）的编写，必须遵循 **统一文档编写规范**，详见 `aicoding_docs/docs/writing-guide.md`。

核心要求速览：
- **结构化**：每个文档有明确的章节骨架，不写流水账。
- **验收驱动**：spec 必须含可验证的「完成定义」清单。
- **决策留痕**：非平凡技术选型必须记 ADR（背景/决策/后果）。
- **过时标注**：过时内容用 `> ⚠️ OUTDATED:` 标注，不得静默删除。
- **变更同步**：改变行为的改动必须同步更新对应文档。

### 1.1 本项目文档结构

```
agent_team/
├── AGENTS.md                          ← 你在这里：项目规范总纲
├── docs/
│   ├── 01-需求分析文档.md             ← 已产出
│   ├── 02-技术架构设计.md             ← 已产出
│   ├── 03-系统设计与实施规划.md        ← 已产出
│   ├── 04-系统架构图.md               ← 已产出（PlantUML）
│   ├── 05-快速上手部署.md             ← 已产出
│   ├── README.md                      ← 文档导读
│   ├── CONSTRAINTS.md                 ← 项目质量红线
│   ├── domain.md                      ← 领域模型与统一语言
│   ├── adr/                           ← 架构决策记录
│   ├── spec/                          ← 功能规格与验收标准
│   ├── plan/                          ← 实现计划
│   └── standards/                     ← 项目专属规范（可选覆写）
├── internal/                          ← Go 源码
├── cmd/                               ← Go 入口
├── tui/                               ← TypeScript TUI 源码
└── deploy/                            ← 容器化部署
```

---

## 2. Vibe Coding 协作流程

核心理念：**人对意图负责，AI 对实现负责**。每个功能遵循统一的三阶段闭环。

### 2.1 三阶段闭环

```
规格(Spec) → 计划(Plan) → 实现(Implement)
   ↑                                    │
   └──────── 反馈验证(Verify) ──────────┘
```

1. **规格（Spec-first）**：先写清楚「做什么、为谁做、验收标准」，存入 `docs/spec/`。
   - 清晰的需求优先；模糊的需求先用「一问一答」澄清到 ~95% 置信度。
2. **计划（Plan）**：拆解为可独立验证的原子步骤，存入 `docs/plan/`。
   - 标注并行与关键路径；复杂步骤拆到单文件单职责。
3. **实现（Implement）**：按计划逐步实现，每步完成后跑验证。
   - 实现前先写测试；不确定时问用户，不静默假设。
4. **反馈验证（Verify）**：跑测试、lint、构建、质量门禁；红则回 Spec/Plan 修正。

### 2.2 模板

| 类型 | 模板路径 | 输出目录 |
|---|---|---|
| 功能规格 | `aicoding_docs/docs/templates/SPEC.md` | `docs/spec/{功能名}.md` |
| 实现计划 | `aicoding_docs/docs/templates/PLAN.md` | `docs/plan/{功能名}.md` |
| 架构决策 | `aicoding_docs/docs/templates/ADR.md` | `docs/adr/ADR-{编号}-{短标题}.md` |

---

## 3. 开发规范引用

本项目主语言为 **Go**（后端编排引擎），客户端为 **TypeScript**（TUI）。以下规范为强制生效（叠加 `general.md` 通用原则）。

### 3.1 通用开发规范

详见 `aicoding_docs/docs/standards/general.md`。
涵盖：根因优先、最小改动、复杂度控制、知识显性化、命名规范。

### 3.2 Go 开发规范（后端强制）

详见 `aicoding_docs/docs/standards/languages/go.md`。
涵盖：Go 工具链、项目布局（cmd/internal/pkg）、包命名、错误处理、goroutine 并发、接口设计。

本项目 Go 源码须遵循 `internal/` + `pkg/` + `cmd/` 标准布局，包名小写单数短名，业务代码禁止放 `cmd/`。

### 3.3 TypeScript 规范（TUI 客户端强制）

详见 `aicoding_docs/docs/standards/languages/typescript.md`。
涵盖：严格模式、类型设计、`any` 禁令、泛型、模块边界。

### 3.4 测试规范

详见 `aicoding_docs/docs/standards/testing/testing.md`。
涵盖：测试金字塔、TDD 流程、单元/集成/E2E 分层、Mock 边界、覆盖率目标。

### 3.5 安全规范

详见 `aicoding_docs/docs/standards/security/security.md`。
涵盖：输入校验、鉴权与授权、密钥管理、注入防御、依赖审计、合规。

本项目特别注意：**不持有任何 AI 编码 CLI 的 API Key**；鉴权全部交底层宿主 CLI；本地优先，默认不上传代码到云端。

### 3.6 Git 工作流规范

详见 `aicoding_docs/docs/standards/git/git-workflow.md`。
涵盖：分支模型、提交信息规范、合并策略、版本号、发布与回滚、AI 操作边界。

### 3.7 规范完整索引

| 维度 | 规范文件 | 覆盖要点 |
|---|---|---|
| 通用 | `aicoding_docs/docs/standards/general.md` | 根因优先、最小改动、复杂度、命名 |
| Go | `aicoding_docs/docs/standards/languages/go.md` | 布局、包命名、错误处理、并发、接口 |
| TypeScript | `aicoding_docs/docs/standards/languages/typescript.md` | 严格模式、类型设计、any 禁令 |
| 测试 | `aicoding_docs/docs/standards/testing/testing.md` | 测试金字塔、TDD、分层、覆盖率 |
| 安全 | `aicoding_docs/docs/standards/security/security.md` | 鉴权、密钥、注入、依赖审计 |
| Git | `aicoding_docs/docs/standards/git/git-workflow.md` | 分支、提交、合并、版本 |

> 后端若引入 Spring Boot / FastAPI / Node-Express 等框架，参照 `aicoding_docs/docs/standards/frameworks/` 对应规范；本项目编排引擎为 Go，框架规范仅在宿主驱动涉及外部集成时参照。

---

## 4. 质量约束

详见 `docs/CONSTRAINTS.md`。默认阈值速览：

| 维度 | 默认阈值 | 备注 |
|---|---|---|
| Go 单元测试覆盖率 | ≥ 80% | 核心编排逻辑 ≥ 90% |
| 端到端关键路径 | 100% 通过 | A2A 协作闭环 / 完整流水线 |
| Lint | 0 新增警告 | golangci-lint + eslint |
| 构建时长 | ≤ 5 分钟 | CI 全流程含测试 |
| A2A 消息处理 p95 | ≤ 500ms | Coordinator→Agent |
| 容器启动到就绪 | ≤ 30s | 各角色 Agent |
| 依赖漏洞 | 0 高危 | govulncheck + npm audit |
| 质量门禁得分 | ≥ 90 | AiCodingAgentTeam Quality-Gate |

> 红线规则：不得为通过检查而调低阈值或静默跳过检查；确需调整须经用户确认并记 ADR。

---

## 5. 项目专属设计原则

基于本项目编排引擎 + 容器化 + A2A 协作的架构特征，在通用规范之上补充以下原则：

1. **编排与执行分离**：Coordinator 只编排不写代码，宿主 CLI 只执行不决策。违反此原则的代码须重构。
2. **容器即角色**：每个角色 Agent 独立容器，可独立伸缩/替换；角色逻辑不跨容器直接调用。
3. **协议即契约**：Agent 间 A2A 通信，对外 MCP/ACP/A2A 三协议；接口变更须先更新契约再改实现。
4. **工件即通信**：角色间仅通过文件 + 结构化 Verdict 通信，禁止自由对话；评审失败不伪造成功。
5. **确定性优先**：质量校验机器硬执行，不依赖模型自评；指纹快照防无限循环修复。
6. **本地优先**：代码/文档默认不出容器，云端能力需显式开启。
7. **不持有密钥**：鉴权全部交底层 CLI，Coordinator 仅查询 AuthStatus。
8. **规模自适应**：小任务轻量流程，大项目全团队；不做过度编排。

---

## 6. AI 行为公约

### 该做的

- 动手前先读相关文档与代码，基于事实而非猜测。
- 遵循三阶段闭环：规格 → 计划 → 实现。
- 实现前先写测试，完成前先跑验证。
- 长任务定期一句话进度同步。
- 不确定时直接问用户，不静默假设。
- 编辑代码前先确认相关规范（`aicoding_docs/docs/standards/`）的适用范围。
- 新增宿主 Backend 须实现 `pkg/runtime/runtime.go` 的 `Runtime` trait，禁止伪造能力。

### 不该做的

- 不静默删除或降级质量检查。
- 不擅自提交、推送、合并、发版。
- 不为「让测试过」而改断言或删用例。
- 不一次性倾倒大块未经拆分的代码。
- 不输出臆造的文件路径、API 或「已验证」结论。
- 不擅自修复无关 bug（可提，不擅改）。
- 不在 Coordinator 中持有任何 API Key / 凭证。
- 不让角色 Agent 间自由对话（只走 A2A + 工件）。

---

## 附录 A：落地检查清单

新功能落地前，逐项确认：

- [ ] 规格已写入 `docs/spec/` 且验收标准明确
- [ ] 计划已写入 `docs/plan/` 且步骤可验证
- [ ] 已读取对应语言/框架规范（Go/TypeScript）
- [ ] 测试先行且通过
- [ ] 根因已定位，非表面补丁
- [ ] 构建与 lint 通过（golangci-lint + eslint）
- [ ] 质量红线未降级
- [ ] 文档已同步
- [ ] 改动聚焦，无无关修改
- [ ] 涉及 Agent 间通信时，A2A 消息审计已记录

## 附录 B：应用到你的项目

1. 阅读本 `AGENTS.md`。
2. 激活 `aicoding_docs/docs/standards/languages/go.md` 与 `typescript.md`。
3. 调整 `docs/CONSTRAINTS.md` 阈值。
4. 填写 `docs/domain.md`。
5. 通读后即可开始 vibe coding。
