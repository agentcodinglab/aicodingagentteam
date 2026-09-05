# ADR-0014: ACP/MCP Server 真实 JSON-RPC 实现

> 状态：已接受
> 日期：2026-09-05
> 关联规范：`docs/spec/coordinator.md`（MCP/ACP 协议）

## 背景

MCP Server 的 `Serve()` 方法和 ACP Server 的 `Serve()` 方法此前都是纯阻塞 stub：

```go
// 旧 stub
func (s *Server) Serve(ctx context.Context) error {
    <-ctx.Done()
    return ctx.Err()
}
```

外部 MCP/ACP 客户端无法通过协议与 Coordinator 交互——MCP 只能通过 Go API 调用 `GovernFile`/`GovernDirectory`，ACP 完全不可用。部署文档中承诺了 `mcp serve` / `acp serve` CLI 命令，但实现不存在。

## 决策

1. **MCP Server**：实现 stdio JSON-RPC 2.0 服务器，支持 `initialize`、`tools/list`、`tools/call`（govern_file、govern_directory）。外部 MCP 客户端可通过 stdio 与 governance 引擎交互。

2. **ACP Server**：实现 stdio JSON-RPC 2.0 服务器，支持 `initialize`、`session/start`、`session/stop`、`session/list`。外部 ACP 客户端可管理 Agent 会话生命周期。

3. **可测试设计**：两个 Server 都提供 `ServeReader(ctx, reader, writer)` 变体，使 JSON-RPC 逻辑可在单元测试中通过 buffer 驱动，无需真实 stdio。

4. **CLI 命令对齐**：新增 `mcp`、`acp`、`a2a serve`、`continue`、`memory show/capture/recall`、`ci` 命令，使 CLI 命令集与部署文档规格一致。

## 后果

- 正面：MCP/ACP 从 stub 升级为可用协议，外部客户端可集成。
- 正面：ACP 覆盖率 0% → 93.0%，MCP 覆盖率 77.9% → 89.6%。
- 正面：claude/dsh stub 驱动补齐测试，覆盖率 0% → 100%。
- 正面：CLI 命令集与文档规格一致，消除文档-代码不一致债务。
- 负面：MCP `Serve()` 本身仍 0% 覆盖（依赖真实 stdio），但 `ServeReader` 覆盖了所有逻辑路径。