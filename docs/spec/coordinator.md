# 功能：Coordinator 五层调度循环（Orchestration Engine）

> 文件名：`docs/spec/coordinator.md`
> 遵循模板：`aicoding_docs/docs/templates/SPEC.md`
> 关联 ADR：ADR-0001、ADR-0003、ADR-0004、ADR-0006

---

## 背景与目标

Coordinator 是编排引擎的核心，模拟软件开发团队的技术负责人角色。它不写代码，负责：意图路由 → DAG 计划构建 → 角色调度执行 → 确定性质量验证 → 交付打包。

已在 `internal/coordinator/director.go` 中搭建骨架，实现了五层串联（Route→Plan→Schedule→Verify→Finalize），但下游组件均为 stub。本 spec 定义每层的精确行为契约。

### 为谁解决什么问题

- 用户提交需求后，Coordinator 自动决定是走完整流水线还是轻量编辑；
- 编排引擎各组件通过 Coordinator 串联，避免组件间直接耦合。

## 用户故事

- 作为用户，我希望提交一条需求后 Coordinator 自动路由并执行，以便无需手动选择流程。
- 作为 Coordinator，我希望将写角色串行执行、评审角色并行执行，以便保证代码完整性同时最大化并行度。
- 作为 Coordinator，我希望评审失败时 park 任务而非伪造成功，以便保证交付质量。

## 功能描述

### 做什么

#### 第一层：Route 意图路由

- 输入：`types.UserRequest{Message, Backend}`
- 处理：规则匹配判定 `IntentType`（Chat/Explain/QuickEdit/Debug/Build），评估深度、写权限、范围
- 输出：`types.Intent{Type, Depth, WriteAccess, Scope}`
- 规则（`internal/router/router.go` 已实现）：
  - 包含 "build"/"搭建"/"创建项目" → Build，深度 Build，写权限 true
  - 包含 "fix"/"修复"/"bug" → Debug，深度 Feature，写权限 true
  - 包含 "修改"/"change"/"update" → QuickEdit，深度 Trivial，写权限 true
  - 包含 "explain"/"解释" → Explain，深度 Trivial，写权限 false
  - 其他 → Chat，深度 Trivial，写权限 false
- 后续迭代：加入知识检索辅助判定 + 一问一答澄清（~95% 置信度）

#### 第二层：Plan 计划构建

- 输入：`types.Intent`
- 处理：
  - QuickEdit/Chat → 返回空 plan（直接执行，不走 DAG）
  - Debug/Build → 生成 9 节点 DAG（clarify→research→docs→spec→frontend→backend→quality→delivery）
  - DAG 节点含：ID、Phase、Role、DependsOn、ArtifactsIn、ArtifactsOut、Writer 标记
  - 3 个 Gate：docs_confirm(human)、preview_confirm(human)、quality(auto)
- 输出：`*types.Plan`（须持久化到 `.aicodingagentteam/plan.json`，MVP 后续迭代）
- 用户可通过 `/plan` 命令增删、调整任务顺序

#### 第三层：Schedule 调度执行

- 输入：`*types.Plan`
- 处理：
  - **写角色节点**（Writer=true）：串行执行，单写者模型（ADR-0004），mutex 保护
  - **评审角色节点**：并行启动独立 Agent 会话，通过 A2A Bus 委派
  - 角色间禁止自由对话，通信媒介仅限工件文件 + 结构化 Verdict
  - 评审角色返回 Verdict（accept/blocking/advisory）
- 输出：`*scheduler.Result{PlanID, Verdicts, Artifacts, Parked}`
- park 条件：评审角色超时、解析失败、或返回 blocking 且无法自动修复 → 标记 Parked=true

#### 第四层：Verify 自校正验证

- 输入：`[]string`（artifacts 路径列表）
- 处理：确定性校验，不依赖模型自评（ADR-0006 的 fail-open 在治理层，此层是硬校验）
  - go build、go test、go vet（ADR-0009 验证可行）
  - 前后端 API 契约交叉校验（`pkg/contracts`）
  - 密钥泄露扫描
  - Runtime 探针（启动应用访问路由）
- 阻塞项生成修复方案；指纹快照防止无限循环修复
- 输出：`qualitygate.Result{Score, Passed, Blocking, Advisory}`

#### 第五层：Finalize 交付产物

- 输入：`qualitygate.Result` + `scheduler.Result`
- 处理：
  - 输出对应深度工件（Build 输出全套 PRD/架构/质量报告；QuickEdit 输出轻量校验结果）
  - 更新项目记忆库
  - 生成审计证据包（`proof-pack-*.zip` + `scorecard-*.html`）
- 输出：`*types.Delivery{PlanID, Artifacts, Score, Passed, CreatedAt}`

### 规模自适应

| 任务规模 | 启动角色 | 流程 |
|---|---|---|
| QuickEdit（trivial） | 无（直接调宿主） | 轻量治理校验 |
| Debug（feature） | Frontend/Backend + QA | 裁剪流水线 |
| Build（build） | 全部 9 角色 | 完整 9 阶段 |

判定依据：Router 输出的 `Intent.Depth`。

## 验收标准

- [x] `Director.Handle(ctx, req)` 能接收 `UserRequest` 并返回 `*Delivery` 无 panic
- [x] Build 意图生成 9 节点 DAG + 3 门禁，QuickEdit 意图生成空 plan
- [x] 写角色节点（frontend/backend）在 DAG 中标记 `Writer=true`
- [x] 同一时刻仅一个 Writer 节点在执行（单写者模型，ADR-0004）
- [x] 评审角色节点并行执行，返回 Verdict
- [x] 评审角色返回 blocking 时，Coordinator 标记 `Parked=true` 而非伪造成功
- [x] 质量门禁 Score 低于阈值时 `Passed=false`
- [x] `Director` 实现 `api.Handler` 接口（RunPipeline/QuickEdit/Verify/GetPlan）
- [x] Gate（docs_confirm/preview_confirm）在 auto_approve=false 时暂停等待用户 /continue
- [x] 最终 Delivery 含 PlanID、Artifacts、Score、Passed、CreatedAt

## 非目标

- Coordinator 不直接写代码（写代码是宿主 CLI 的职责）。
- 不实现 AI 驱动的意图判定（MVP 阶段用规则匹配，后续迭代再引入知识检索）。
- 不实现 plan.json 持久化（MVP 阶段 plan 在内存中，后续迭代加持久化）。
- 不实现指纹快照防循环（MVP 阶段质量门禁单次执行，后续迭代加重试上限）。

## 依赖与约束

- 前置依赖：Router、Planner、Scheduler、QualityGate 组件均可实例化
- 技术约束：五层必须串行，不可跳层（Route→Plan→Schedule→Verify→Finalize 顺序固定）
- 风险与缓解：
  - 评审角色超时 → park 暂停，不伪造成功
  - 质量门禁组件故障 → fail-open（ADR-0006），返回默认通过

## 关联

- 实现计划：`docs/plan/mvp-preparation.md`
- 相关 ADR：ADR-0001、ADR-0003、ADR-0004、ADR-0006
- 契约：`proto/aicodingagentteam.proto`（gRPC API）、`schemas/plan.json`
- 源码：`internal/coordinator/director.go`、`internal/router/router.go`、`internal/planner/planner.go`、`internal/scheduler/scheduler.go`、`internal/qualitygate/engine.go`