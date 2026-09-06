# P6 — ACP v1 stdio JSON-RPC 完整实现

> 决策依据：`docs/adr/ADR-0020-p6-acp-session-newtask.md`
> 目标版本：v0.7.0
> 前置依赖：P4（v0.6.0，已落地）
> 后续：P5（OpenCode serve HTTP API）

## 0. 范围

**做**：
1. 在 `internal/acp` 实现 `session/newTask` method：接收 `agentId`/`prompt`，返回 task ID。
2. 在 session 运行时，通过 JSON-RPC `notifications/session/update` 推送流式事件（start/message/tool_call/done/error）。
3. 把 ACP server 接到 Director：`session/newTask` 调 `Director.Handle` 或快编路径，把 delivery 的 ProgressEvent 转成 ACP 事件。
4. 完整的涋出测试：模拟 stdio JSON-RPC 客户端，验证 newTask → 事件流 → done 。

**不做**：
- 不动 OpenCode serve HTTP API（属 P5）。
- 不改 MCP server（已实现 tools/list + tools/call）。
- 不实现 ACP 的 `session/prompt`（别名越界限混乱，本期只 newTask）。

## 1. 当前状态（缺口分析）

- `internal/acp/acp.go` 实现了 `initialize`/`session/start`/`session/stop`/`session/list`。
- 缺 `session/newTask`：ACP v1 核心方法，外部客户端无法驱动编排引擎。
- Server 未接 Director，无法把任务交给编排。
- 现有测试只覆盖 initialize/start/stop/list 入口，未覆盖 newTask 事件流。

## 2. 实施步骤

| # | 任务 | 产出物 | 估时 | 关键路径 |
|---|---|---|---|---|
| P6.1 | `session/newTask` method + task ID 生成 | `internal/acp/acp.go` | 30 min | ✓ |
| P6.2 | `notifications/session/update` 推送流式事件 | `internal/acp/acp.go` | 40 min | ✓ |
| P6.3 | ACP server 接 Director（`NewWithDirector`） | `internal/acp/acp.go` + `cmd/aicodingagentteam/main.go` | 30 min | ✓ |
| P6.4 | 涋出测试：模拟 stdio 客户端验证 newTask → 事件流 | `internal/acp/acp_test.go` | 40 min | ✓ |
| P6.5 | ADR-0020 + 计划 | `docs/adr/ADR-0020-*.md` | 15 min | ✓ |
| P6.6 | CHANGELOG + v0.7.0 tag | `CHANGELOG.md` + git tag | 10 min | 末 |

总估时 ~2.5 小时。

## 3. 验收标准

- [ ] `session/newTask` method 返回 task ID
- [ ] `notifications/session/update` 推送 start/message/tool_call/done/error
- [ ] ACP server 能接 Director（`NewWithDirector`）
- [ ] `TestACP_SessionNewTask_StreamsEvents`：模拟 stdio 客户端，验证 newTask → 事件流 → done
- [ ] 现有 ACP initialize/start/stop/list 测试不破坏
- [ ] `go build ./...` + `go test ./...` 全绿
- [ ] 无 umadev/umacloud/goder.ai 字样

## 4. 不在范围

- OpenCode serve HTTP API -> P5
- MCP server（已实现）
- `session/prompt`（本期不做，避免越界混乱）
- A2A HTTP server（已实现）

## 5. 风险与罓解

| 风险 | 罓解 |
|---|---|
| JSON-RPC notification 无 ID，客户端不易关联 | 事件含 sessionId + taskId 字段 |
| Director.Handle 同步阻塞，事件流会等完才发 | 用 goroutine 调 Handle，事件通 channel 推送 |
| stdio 写入与响应交织 | 响应用 mutex 保护，notification 单独线写 |

## 6. 文件清单

| 操作 | 路径 |
|---|---|
| 修改 | `internal/acp/acp.go` |
| 修改 | `internal/acp/acp_test.go` |
| 修改 | `cmd/aicodingagentteam/main.go`（acp 命令接 Director） |
| 新建 | `docs/adr/ADR-0020-p6-acp-session-newtask.md` |
| 修改 | `CHANGELOG.md` + tag v0.7.0 |