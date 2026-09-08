# PLAN: P10 生产化硬化 + P13 开发者体验

> 关联 SPEC：`docs/spec/p10-p13-init-wizard-plugin.md`
> 基线：v0.9.0 已发版

## 实施步骤

### P10.1 — init 向导增强

| # | 任务 | 文件 | 估时 |
|---|---|---|---|
| 10.1.1 | 重构 `cmdInit`：支持 `--non-interactive` flag，交互式选择 backend/threshold/auto-approve | `cmd/aicodingagentteam/main.go` | 40min |
| 10.1.2 | 生成 `.aicodingagentteam/config.json`，含用户选择 | `cmd/aicodingagentteam/main.go` | 20min |
| 10.1.3 | 测试：交互模式 + 非交互模式 + 覆盖确认 | `cmd/aicodingagentteam/coverage_test.go` | 30min |

### P10.2 — TUI↔Coordinator gRPC 连接验证

| # | 任务 | 文件 | 估时 |
|---|---|---|---|
| 10.2.1 | 集成测试：启动真实 gRPC server + client 调 RunPipeline/GetPlan/Verify | `pkg/api/server_grpc_test.go` | 40min |

### P13.1 — 插件发现机制

| # | 任务 | 文件 | 估时 |
|---|---|---|---|
| 13.1.1 | 新建 `pkg/plugin` 包：全局注册表 + `Register()` + `All()` | `pkg/plugin/plugin.go` | 20min |
| 13.1.2 | `Registry` 改为自动加载 `plugin.All()` 注册的驱动 | `internal/host/registry.go` | 15min |
| 13.1.3 | 示例驱动：`pkg/plugin/examples/stub-driver/` 自注册演示 | `pkg/plugin/examples/` | 20min |
| 13.1.4 | 测试：插件注册 + Registry 加载 + 覆盖 | `pkg/plugin/plugin_test.go` | 20min |

### P13.2 — 宿主能力自省 CLI

| # | 任务 | 文件 | 估时 |
|---|---|---|---|
| 13.2.1 | `backends` 子命令：列出驱动 + Capabilities + AuthStatus | `cmd/aicodingagentteam/main.go` | 25min |
| 13.2.2 | 测试 + printUsage 更新 | `cmd/aicodingagentteam/coverage_test.go` | 10min |

### 收尾

| # | 任务 | 文件 |
|---|---|---|
| F1 | ADR-0023 记录插件机制决策 | `docs/adr/` |
| F2 | CHANGELOG 更新 | `CHANGELOG.md` |
| F3 | CONTRIBUTING 补充「贡献新宿主驱动」章节 | `CONTRIBUTING.md` |

## 关键路径

```
13.1.1 → 13.1.2 → 13.1.3 → 13.1.4（插件机制，其他任务的前置）
10.1.1 → 10.1.2 → 10.1.3（init 向导，独立）
10.2.1（gRPC 集成测试，独立）
13.2.1 → 13.2.2（backends 命令，依赖插件机制）
F1→F2→F3（收尾）
```
