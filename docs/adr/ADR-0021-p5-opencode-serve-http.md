# ADR-0021：P5 — OpenCode 驱动集成方案（spike 后修订）

> 状态：Accepted（选项 B）
> 日期：2026-09-06
> 关联：ADR-0007（宿主 CLI 可行性）、ADR-0019（P4 流式）、ADR-0020（P6 ACP）
> 计划：`docs/plan/p5-opencode-serve-http.md`
> Spike：`spike/opencode-serve/notes.md`

## 背景

原 P5 范围基于假设：`opencode serve` 是长驻 JSON API，通过 HTTP/SSE 接受任务。spec/host-driver.md 明确标注 serve 模式 API 「需补充调研」。

## spike 结果（P5.0）

跑 `opencode serve --port 14148 --hostname 127.0.0.1`，端点探测结果：

- `GET /` 返回 HTML 单页应用（`<title>OpenCode</title>` + `<div id=root>`），不是 JSON
- `GET /api`/`/session`/`/sessions`/`/v1`/`/openapi.json` 等所有路径都返回该 SPA HTML（SPA fallback）
- 服务器输出：`Warning: OPENCODE_SERVER_PASSWORD is not set` 与 `listening on http://127.0.0.1:14148`

**结论**：OpenCode v1.18.18 的 `serve` 是 Web UI HTTP server，不是 JSON API。与 spec 假设不一致。

OpenCode 提供的编程接口是 `opencode acp`（stdio JSON-RPC ACP），与我们刚做的 P6 ACP server 同质。

## 决策选项（待选择）

详见 `docs/plan/p5-opencode-serve-http.md`。推荐选项 A：维持现状 + 修订 spec。

- **A（推荐）**：opencode driver 仍用 `opencode run --format json`，配合 P4 流式。仅修订 spec 文档，不动代码。
- **B**：opencode driver 改调 `opencode acp`。会与 P6 ACP server 叠加。
- **C**：丢弃 P5，后续推进方向 B（Claude/DSH 真实 exec）。

## 后果（选 A）

- 正面：不动代码、不增加复杂度；P4 流式已解决实时性；与 OpenCode 现状一致
- 负面：不能复用 session（每次 run 是新 session）
- 中性：OpenCode 未来如增加编程 API，可启 P5 跟进

## 验收（选 A）

- spec 修改后与代码现状一致
- `go build ./...` + `go test ./...` 全绿（无代码变动，应默认返绿）