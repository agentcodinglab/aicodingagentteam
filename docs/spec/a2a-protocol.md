# 功能：A2A 协议（Agent-to-Agent Protocol）

> 文件名：`docs/spec/a2a-protocol.md`
> 遵循模板：`aicoding_docs/docs/templates/SPEC.md`
> 关联 ADR：ADR-0002、ADR-0003

---

## 背景与目标

Agent Team 的角色间协作基于 A2A（Agent-to-Agent）协议。Coordinator 作为 A2A Hub 调度各专业角色 Agent，每个角色独立容器/会话，不共享主会话历史。已在 `internal/a2a/a2a.go` 中定义 Bus/Agent/Task/Result/AgentCard 接口和 in-process stub。

### 为谁解决什么问题

- Coordinator 需要将任务委派给专业角色 Agent 并收集结构化评审结果；
- 角色 Agent 间禁止直接对话，所有通信经 Coordinator 路由，保证可审计。

## 用户故事

- 作为 Coordinator，我希望通过 A2A Bus 向 QA Agent 委派测试任务并接收 Verdict，以便评审交付质量。
- 作为 Coordinator，我希望查询所有注册 Agent 的能力声明（AgentCard），以便按能力分派任务。
- 作为外部 Agent，我希望通过 A2A Server 发现 AiCodingAgentTeam 的 Agent Card 并委派跨团队任务。

## 功能描述

### 做什么

1. **Agent Card 发现**
   - 每个 Agent 暴露 AgentCard（ID/Name/Role/Capabilities/Endpoint/MaxConcurrent/TimeoutDefault）
   - HTTP 端点 `/.well-known/agent.json` 返回 Coordinator 的 Card（已在 `pkg/api/server.go` 实现）
   - `Bus.Discover()` 返回所有已注册 Agent 的 Card 列表

2. **任务委派**
   - `Bus.Delegate(ctx, Task)` 根据 `Task.Role` 路由到对应 Agent
   - Task 结构：TaskID、Role、Payload（artifacts + instruction）、Deadline
   - 消息格式：JSON-RPC 2.0 over HTTP / Redis Pub/Sub

3. **结果回传**
   - Agent 返回 `Result{TaskID, Verdict}`
   - Verdict 结构：Decision（accept/blocking/advisory）、Severity、Findings[]、Artifacts[]

4. **进度事件**
   - Agent 执行中发送 `agent.progress` 事件（TaskID、Phase、Status、Message）
   - Coordinator 订阅进度流，实时推送给 TUI

5. **状态查询**
   - `Agent.Status(ctx)` 返回当前 Agent 状态（idle/running/parked）

6. **通信约束**
   - Agent 间禁止直接对话，所有通信经 Coordinator 路由
   - Agent 返回的 Verdict 是结构化的（不允许自由文本评审）
   - 评审 Agent 失败（超时/解析失败）标记 park，不伪造成功

7. **消息总线实现**
   - MVP 阶段：in-process channel（`internal/a2a/a2a.go` 已 stub）
   - 容器化阶段：Redis Pub/Sub（ADR-0002），跨容器通信
   - 降级：Redis 不可用时回退 in-process，不报错失败

### 通信模式

```
Coordinator                    Agent (QA)
    │                             │
    │── agent.execute(task) ─────▶│
    │                             │ (处理中)
    │◀── agent.progress(event) ───│
    │                             │
    │◀── agent.result(verdict) ───│
```

### JSON-RPC 消息格式

委派消息（`schemas/a2a-message.json` 已定义）：
```json
{"jsonrpc":"2.0","id":"task-uuid","method":"agent.execute","params":{"task_id":"n8","role":"qa","payload":{"artifacts":["output/app-prd.md"],"instruction":"基于验收标准编写测试用例"},"deadline":"2026-09-02T12:00:00Z"}}
```

结果消息：
```json
{"jsonrpc":"2.0","id":"task-uuid","result":{"task_id":"n8","verdict":{"decision":"blocking","severity":"critical","findings":[{"check":"test-coverage","status":"fail","detail":"覆盖率 65%，阈值 80%","evidence":"coverage.json"}],"artifacts":["output/app-test-report.md"]}}}
```

## 验收标准

- [x] `Bus.Register(agent)` 注册后 `Bus.Discover()` 能返回该 Agent 的 Card
- [x] `Bus.Delegate(ctx, task)` 能按 `task.Role` 路由到对应 Agent
- [x] 未注册角色的 `Delegate` 返回明确 error（"no agent registered for role X"）
- [x] Agent 返回的 Verdict 含 Decision 字段（accept/blocking/advisory）
- [x] blocking Verdict 的 Severity 非空
- [x] Agent 超时时 Coordinator 收到 error 而非阻塞
- [x] Agent 超时不伪造 accept Verdict
- [x] HTTP `/.well-known/agent.json` 返回合法 AgentCard JSON
- [x] AgentCard JSON 符合 `schemas/agent-card.json` Schema
- [x] A2A 消息符合 `schemas/a2a-message.json` Schema
- [x] MVP in-process Bus 所有测试通过，无 Redis 依赖

## 非目标

- 不实现 Redis Pub/Sub（MVP 用 in-process，容器化阶段替换）。
- 不实现 Agent 间直接通信（所有通信经 Coordinator）。
- 不实现消息持久化（Redis Pub/Sub 非持久，Agent 离线消息丢失，配合任务重试）。
- 不实现跨团队 A2A 互联（MVP 仅 Coordinator 内部调度，外部 A2A 互联后续迭代）。

## 依赖与约束

- 前置依赖：`internal/types`（Role/Verdict/Finding 定义）、`schemas/agent-card.json`、`schemas/a2a-message.json`
- 技术约束：JSON-RPC 2.0 格式、Agent 间无直接对话
- 风险与缓解：
  - Redis 不可用 → 降级 in-process（ADR-0008）
  - Agent 消息丢失 → 任务重试 + park 暂停

## 关联

- 实现计划：`docs/plan/mvp-preparation.md`
- 相关 ADR：ADR-0002（Redis 总线）、ADR-0003（容器化）、ADR-0008（基础设施）
- 契约：`schemas/a2a-message.json`、`schemas/agent-card.json`
- 源码：`internal/a2a/a2a.go`、`pkg/api/server.go`