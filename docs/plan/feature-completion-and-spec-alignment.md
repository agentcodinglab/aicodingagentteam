# 实现计划：功能补齐与规格对齐（方向 I）

> 文件名：`docs/plan/feature-completion-and-spec-alignment.md`
> 目标：消除文档与代码不一致的债务，补齐 spec/部署文档承诺但未实现的 CLI 命令和 stub

---

## I1: CLI 缺失命令补齐
- 文件：`cmd/aicodingagentteam/main.go`
- 新增命令：
  - `continue` — 继续 parked/paused 工作流（调用 `Director.ContinuePlan`）
  - `mcp serve` — 启动 MCP JSON-RPC 服务（stdio）
  - `a2a serve` — 启动 A2A HTTP 服务（复用 serve 逻辑，仅暴露 A2A 端口）
  - `acp serve` — 启动 ACP JSON-RPC 服务（stdio）
  - `memory show` — 显示项目记忆（facts/pitfalls/lessons）
  - `ci` — 治理扫描别名（等价 `govern --ci`）
- 验证：各命令可执行不 panic

## I2: ACP Server 真实实现
- 文件：`internal/acp/acp.go`、`internal/acp/acp_test.go`
- 改动：参照 MCP 模式，实现 stdio JSON-RPC 2.0（initialize/session/start/stop）
- 目标：acp 覆盖率 0% → 80%+
- 验证：`go test ./internal/acp/...` 通过

## I3: stub 驱动基础测试
- 文件：`internal/host/claude/claude_test.go`、`internal/host/dsh/dsh_test.go`
- 补齐：TestModelInfo、TestCapabilities、TestSendTask、TestAuthStatus、TestStartDestroySession、TestPauseResume
- 目标：claude/dsh 0% → 80%+
- 验证：`go test ./internal/host/...` 通过

## I4: mcp 覆盖率补齐
- 文件：`internal/mcp/mcp_test.go`
- 补齐：GovernDirectory error path、toolDefinitions 内容验证
- 目标：mcp 77.9% → 80%+
- 验证：`go test ./internal/mcp/... -cover` ≥ 80%

## I5: 文档与代码对齐
- 文件：`docs/05-快速上手部署.md`、`README.md`
- 改动：命令速查表补齐新增命令，标注未实现项
- 验证：文档命令与 CLI 一致

## 关键路径
```
I1 (CLI) → I2 (ACP) → I3 (stub tests) ∥ I4 (mcp tests) → I5 (docs)
```