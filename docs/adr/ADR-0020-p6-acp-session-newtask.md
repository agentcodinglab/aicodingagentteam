# ADR-0020：P6 — ACP v1 `session/newTask` 与流式事件

> 状态：Accepted
> 日期：2026-09-06
> 关联：ADR-0014（ACP/MCP 真实实现）、ADR-0019（P4 流式）
> 计划：`docs/plan/p6-acp-session-newtask.md`

## 背景

ACP server 实现了 `initialize`/`session/start`/`session/stop`/`session/list`，但缺 `session/newTask`——ACP v1 的核心方法。外部客户端（IDE 插件等）无法向 session 发任务并拿流式事件。同时 Server 未接 Director，无法把任务交给编排引擎。

## 决策

1. 实现 `session/newTask` method：接收 `agentId`/`prompt`，返回 task ID。
2. `session/newTask` 调 `Director.Handle`（同步 goroutine 中），把 delivery 的 ProgressEvent 转成 ACP 事件（start/message/tool_call/done/error）。
3. 通过 JSON-RPC `notifications/session/update`（无 ID）推送流式事件，事件含 `sessionId`+`taskId` 字段以供客户端关联。
4. `NewWithDirector` 构造器让 ACP server 指向 Director。

## 备选方案

- 同步调 Handle 再发事件：事件流要等整个 delivery 完成才开始，不是流式，否决。
- 只发 done 不发中间事件：外部客户端体验差，否决。

## 后果

- 正面：外部 ACP 客户端可驱动编排引擎并拿流式进度；ACP 协议覆盖度提升。
- 负面：JSON-RPC notification 无 ID，客户端需靠 sessionId+taskId 关联（已在事件中带）。
- 中性：现有 initialize/start/stop/list 不变。

## 验收

- `TestACP_SessionNewTask_StreamsEvents` 跑通
- 现有 ACP 测试无回归
- `go build ./...` + `go test ./...` 全绿