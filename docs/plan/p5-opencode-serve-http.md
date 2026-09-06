# P5 — OpenCode driver 改调 `opencode acp`（选项 B）

> 决策依据：`docs/adr/ADR-0021-p5-opencode-serve-http.md`
> 状态：Accepted（选项 B）
> 目标版本：v0.8.0
> 前置依赖：P6（v0.7.0，已落地）
> 后续：方向 B 后续（Claude/DSH 真实 exec）

## 0. 背景

P5.0 spike（`spike/opencode-serve/notes.md`）披露 `opencode serve` v1.18.18 是 Web UI HTTP，不是 JSON API。
OpenCode 提供的编程接口是 `opencode acp`（stdio JSON-RPC），与我们刚做的 P6 ACP server 同质。
选项 B：opencode driver 改调 `opencode acp`。

## 1. 范围

**做**：
1. opencode driver 重构：预启动 `opencode acp` 子进程，通过 stdio 发 JSON-RPC 调用，从 stdout 读流式事件。
2. ACP 调用走流程：`initialize` -> `session/new` -> `session/prompt` → 逐行读返回的 JSON-RPC 响应 / notification。
3. 事件转换：ACP 事件 -> `runtime.Event` yield，与 P4 橁式一致。
4. 进程生命周期管理：启动/健康检查/优雅退出。
5. CI 用 stub JSON-RPC server（in-process，模拟 acp 响应）跑驱动端到端。本地真实 opencode：`scripts/e2e-real-opencode.ps1`。

**不做**：
- 不改 codex / claude / dsh driver。
- 不动 P6 ACP server（已做）。
- 不补 P5 原范围（「serve 长驻 HTTP + SSE」）——与 OpenCode v1.18 实际不符。

## 2. 实施步骤

| # | 任务 | 产出物 | 估时 | 关键路径 |
|---|---|---|---|---|
| P5.1 | 探明 opencode acp 的 JSON-RPC 方法集（使用 `opencode acp --pure` 跑下）| spike 附录 | 30 min | ✓ |
| P5.2 | opencode driver 重构：调用 `opencode acp` 子进程 | `internal/host/opencode/opencode.go` | 60 min | ✓ |
| P5.3 | stub JSON-RPC server 测试（in-process）| `internal/host/opencode/acp_stub_test.go` | 40 min | ✓ |
| P5.4 | 本地真实 opencode 脚本 | `scripts/e2e-real-opencode.ps1` | 15 min | |
| P5.5 | ADR-0021 状态改 Accepted + 本计划为选 B | 已 | 0 |  |
| P5.6 | CHANGELOG + v0.8.0 tag | `CHANGELOG.md` + git tag | 10 min | 末 |

总估时 ~2.5 小时。

## 3. 验收标准

- [ ] opencode driver 预启动 `opencode acp` 子进程，stdio 上调 JSON-RPC
- [ ] 事件流：ACP notification 逐条 yield `runtime.EventMessage` / `EventToolCall` / `EventDone`
- [ ] 会话复用：多个 SendTask 复用同一 session
- [ ] `TestOpenCode_ACP_StubServer`：用 stub JSON-RPC server 跑驱动端到端。
- [ ] `scripts/e2e-real-opencode.ps1` 本地验证。
- [ ] 现有 scheduler/coordinator 测试无回归。
- [ ] `go build ./...` + `go test ./...` 全绿。
- [ ] 无 umadev/umacloud/goder.ai 字样。

## 4. 不在范围

- 修订 spec 中「serve 长驻 HTTP + SSE」说法（与 B 方案交叉，交给后续）
- codex / claude / dsh driver
- P6 ACP server（已做）

## 5. 风险与缓解

| 风险 | 缓解 |
|---|---|
| ACP 响应格式与预期不一致（某些版本不同）| P5.1 先跑验证；不能验证则调整 |
| 与 P6 ACP server 叠加（同为 ACP 协议）| driver 仅使用 stdio RPC，不复用 ACP server 代码；独立维护 |
| 子进程崩溃 / 占用端口 | 用 `--port 0`（随机端口）；健康检查 + 会话重连 |
| Stub server 难以验证与 P6 server 错位 | stub server 使用 P5 本身封装的 JSON-RPC 类型，与 P6 server 无关 |

## 6. 文件清单

| 操作 | 路径 |
|---|---|
| 改写 | `internal/host/opencode/opencode.go` |
| 新建 | `internal/host/opencode/acp_stub_test.go` |
| 新建 | `scripts/e2e-real-opencode.ps1` |
| 修改 | `docs/adr/ADR-0021-p5-opencode-serve-http.md`（状态 Accepted） |
| 修改 | `CHANGELOG.md` + tag v0.8.0 |