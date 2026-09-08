# SPEC: P10 生产化硬化 + P13 开发者体验（init 向导 + 插件机制）

> 关联 ADR：ADR-0022（覆盖率路线图）、ADR-0005（不持有密钥）、ADR-0007（宿主可行性）
> 关联计划：`docs/plan/p10-p13-init-wizard-plugin.md`

## 1. 背景

v0.9.0 MVP 功能全部完成。盘点发现：

- RedisBus 已实现（`internal/a2a/redis.go`），`NewBusFromEnv` 已消费 `AICODINGAGENTTEAM_A2A_BUS` 环境变量，fail-open 回退到 InProcBus ✅
- gRPC Server 已实现（`pkg/api/server.go`），启动后监听端口 ✅
- TUI gRPC Client 已实现（`tui/src/grpc/client.ts`）✅
- Host Driver Registry 已支持 `Register(backend, runtime.Runtime)` 动态注册 ✅

**实际缺口：**

1. **`cmdInit` 过于简陋**：仅创建目录，不生成配置文件、不交互选择宿主 CLI、不配置质量门禁阈值
2. **无插件发现机制**：`Registry.Register()` 存在但只能编译时调用，社区无法在不修改主仓库源码的情况下贡献新宿主驱动
3. **无 TUI↔Coordinator 连接验证测试**：gRPC server + client 都实现了但从未端到端测过

## 2. 功能需求

### P10.1 — init 向导增强

**用户故事**：作为开发者，我运行 `aicodingagentteam init` 时，系统交互式引导我选择宿主 CLI 后端、配置质量门禁阈值、生成 `.aicodingagentteam/config.json`。

**验收标准：**
- [ ] `init` 无参数时进入交互模式（stdin 读取），有 `--non-interactive` 跳过交互用默认值
- [ ] 交互步骤：1) 选择默认宿主 CLI（codex/opencode/claude-code/deepseek-dsh）2) 质量门禁阈值（默认 90）3) 自动确认门禁（默认 false）
- [ ] 生成 `.aicodingagentteam/config.json`，含 selected backend + threshold + auto-approve
- [ ] 非交互模式生成与 `config.Default()` 一致的配置
- [ ] 已存在 config.json 时提示覆盖确认

### P10.2 — TUI↔Coordinator 连接验证测试

**验收标准：**
- [ ] Go 集成测试：启动 `api.Server`（真实 gRPC listen），用 `grpc.DialContext` 调用 RunPipeline/GetPlan/Verify，断言响应
- [ ] 测试覆盖 gRPC server 正常启动 + 响应 + context cancel

### P13.1 — 插件发现机制（Plugin Discovery）

**用户故事**：作为社区贡献者，我可以实现 `runtime.Runtime` 接口，编译为独立 Go 包，通过 `init()` 自注册到全局 Registry，用户只需 `import _ "github.com/community/my-driver"` 即可启用。

**验收标准：**
- [ ] 新增 `pkg/plugin` 包，提供 `Register(backend, runtime.Runtime)` 全局注册函数
- [ ] `Registry` 改为从全局插件表自动加载所有注册的驱动
- [ ] 提供 `pkg/plugin/examples/stub-driver` 示例驱动，演示自注册模式
- [ ] 文档：`CONTRIBUTING.md` 补充「如何贡献新宿主驱动」章节
- [ ] 不破坏现有 `host.NewRegistry()` 的 4 个内置驱动

### P13.2 — 宿主能力自省 CLI

**验收标准：**
- [ ] `aicodingagentteam backends` 命令列出所有已注册宿主驱动 + 各自 Capabilities + AuthStatus
- [ ] 输出格式：`backend name | session_resume | tool_calls | web_search | auth_ready`

## 3. 非功能约束

- 不持有任何 API Key（ADR-0005）
- 插件机制不引入外部依赖（纯 Go 标准库 + 已有依赖）
- init 向导的 stdin 读取不阻塞 CI（`--non-interactive` 模式用于测试）
- 覆盖率：新增代码 ≥ 85%

## 4. 完成定义

- [ ] `go test ./... -cover` 全包达标
- [ ] `go build ./...` 通过
- [ ] `go vet ./...` 通过
- [ ] ADR + CHANGELOG 更新
