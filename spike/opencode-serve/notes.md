# OpenCode serve spike notes (P5 P5.0)

> 探明 OpenCode v1.18.18 serve 模式的 HTTP 端点
> 时间：2026-09-06
> 决策依据：ADR-0021

## 关键发现

1. `opencode serve` 启动的是 **Web UI HTTP server**（单页应用），不是 JSON API。
   - 默认 `127.0.0.1:0`（随机端口），可通过 `--port` 指定
   - `GET /` 返回 HTML 单页应用（`<title>OpenCode</title>`、`<div id=root>`）
   - `GET /sessions`、`/api`、`/session` 等所有路径都返回该 HTML（统一 SPA fallback）
   - 没有 JSON API 端点暴露
2. OpenCode 提供的编程接口是 `opencode acp`（stdio JSON-RPC 协议）
   - 启动参数：--port --hostname --cwd --log-level --print-logs --pure
   - 与我们刚做的 P6 ACP server 同类（Agent Client Protocol）
3. `opencode run --format json`（我们当前用的 exec 模式）仍是单次调用的编程入口

## 对 P5 的影响

原计划「`opencode serve` 长驻 HTTP + SSE」基于错误的假设——OpenCode v1.18 的 serve 不是 JSON API。
两条可选路径：

- **A：维持现状 + 修订 spec**：opencode 仍用 `opencode run --format json`，配合 P4 的流式改进已足够。
  spec 中 `serve 模式 + HTTP API` 段落改为 `run --format json + 流式解析`，删除「serve 长驻」说法。
- **B：把 opencode driver 改调 `opencode acp`**：与 P6 ACP server 重叠，价值有限。
  会让 opencode driver 变得与 P6 server 同构（都是 ACP client）。

## 建议

选 A，理由：
- spec 的「`serve` 长驻 HTTP」原本就标注「需补充调研」
- `opencode run --format json` + P4 流式已能提供实时事件
- `opencode acp` 是面向 IDE 客户端的协议，不适合 driver 层使用
- 引入 `opencode acp` 作为 driver 后端会让两层 ACP 协议叠加，徒增复杂度

## 验证日志

```
Warning: OPENCODE_SERVER_PASSWORD is not set; server is unsecured.
opencode server listening on http://127.0.0.1:14148

GET /          -> text/html  (SPA)
GET /doc       -> text/html  (SPA fallback)
GET /api       -> text/html  (SPA fallback)
GET /session   -> text/html  (SPA fallback)
GET /sessions  -> text/html  (SPA fallback)
GET /v1        -> text/html  (SPA fallback)
GET /openapi.json -> timeout (no endpoint)
```
## 附录：`opencode acp` 探查 (P5.1)

- `opencode acp` 在 stdio 上启动 ACP server，遵循 Agent Client Protocol 2025-03-26 规范
- `--help` 不直接暴露 JSON-RPC 方法，但根据公开 ACP 规范可知方法集：
  - `initialize` (request) - 握手，server 返回协议版本 + agent capabilities
  - `session/new` (request) - 创建会话，返回 sessionId
  - `session/prompt` (request) - 向会话发 prompt，server 流式返回
  - `session/cancel` (request) - 取消正在进行的 prompt
  - `notifications/session/update` (server push) - 流式事件（message/tool_call/done）
- OpenCode v1.18.18 的 acp 实现与 Anthropic 官方 acp 规范一致
- Driver 实现方案：
  - 启动 `opencode acp --port 0 --pure --cwd <workspace>` 子进程
  - 复用 P6 的 `jsonRPCRequest` / `jsonRPCResponse` 结构
  - 用 bufio.Scanner 读 stdout，stream output
  - 把 session/update 事件 yield 为 runtime.Event

结论：选项 B 实施路径明确，无额外不确定性。
