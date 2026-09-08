# ADR-0023：插件发现机制与 init 向导增强

> 状态：Accepted
> 日期：2026-09-08
> 关联：ADR-0005（不持有密钥）、ADR-0007（宿主可行性）、ADR-0022（覆盖率路线图）
> 计划：`docs/plan/p10-p13-init-wizard-plugin.md`

## 背景

v0.9.0 MVP 完成后，盘点发现两个开发者体验缺口：

1. `cmdInit` 仅创建目录，不生成配置文件，不交互引导用户选择宿主 CLI 和质量门禁阈值。新用户不知道需要配置什么才能运行。
2. `Registry.Register()` 虽然支持动态注册驱动，但只能在编译时在 `NewRegistry()` 内部调用。社区贡献者无法在不修改主仓库源码的情况下添加新宿主驱动。

## 决策

### 决策 1：全局插件注册表（pkg/plugin）

创建 `pkg/plugin` 包，提供全局 `Register(backend, runtime.Runtime)` 函数。社区驱动包通过 `init()` 自注册：

```go
package mystubdriver

func init() {
    plugin.Register("my-backend", &myDriver{})
}
```

用户只需 blank import 即可激活：
```go
import _ "github.com/community/my-driver"
```

`host.NewRegistry()` 在注册 4 个内置驱动后，自动加载 `plugin.All()` 中所有插件驱动。

### 决策 2：init 向导交互式增强

`cmdInit` 支持 `--non-interactive` flag。交互模式引导用户选择：
- 默认宿主 CLI 后端（codex/opencode/claude-code/deepseek-dsh）
- 质量门禁阈值（默认 90）
- 自动确认门禁（默认 false）

生成 `.aicodingagentteam/config.json`，与 `config.Load()` 兼容。

### 决策 3：backends 自省命令

新增 `aicodingagentteam backends` 命令，列出所有已注册驱动的 Capabilities + AuthStatus，帮助用户验证环境配置。

### 决策 4：gRPC 端到端集成测试

新增 `pkg/api/server_grpc_test.go`，用真实 gRPC listener + client 验证 RunPipeline/GetPlan/Verify/QuickEdit 四个 RPC 全链路。

## 后果

- 社区可贡献新宿主驱动，无需 fork 主仓库。
- 新用户通过 `init` 向导快速配置项目，降低上手门槛。
- `backends` 命令提供环境自省，便于调试。
- gRPC 集成测试覆盖 TUI↔Coordinator 连接路径，消除「实现了但没测过」的盲区。
