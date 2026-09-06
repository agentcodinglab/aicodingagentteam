# 功能：宿主驱动层（Host Driver Layer）

> 文件名：`docs/spec/host-driver.md`
> 遵循模板：`aicoding_docs/docs/templates/SPEC.md`
> 关联 ADR：ADR-0001、ADR-0005、ADR-0007、ADR-0019（P4 流式）、ADR-0021（P5 opencode acp）

---

## 背景与目标

AiCodingAgentTeam 不拥有大模型，通过调用外部 AI 编码 CLI（Claude-Code、Codex、OpenCode、DeepSeek-DSH）作为执行器。宿主驱动层是编排引擎与 AI CLI 之间的唯一桥梁，负责：子进程生命周期管理、消息协议转换、能力差异处理、鉴权状态查询。

> ⚠️ OUTDATED: 以下关于 OpenCode `serve` HTTP 模式的描述基于 ADR-0007 时的初步假设，已被 ADR-0021（P5 spike）证伪——OpenCode v1.18.18 的 `serve` 实为 Web UI HTTP server（SPA fallback），并非 JSON API。OpenCode 真正的编程接口是 `opencode acp`（stdio JSON-RPC, Agent Client Protocol 2025-03-26）。driver 现据此实现（ADR-0021 选项 B）。

ADR-0007 Spike 已验证：Codex 支持 `exec` 非交互模式（26s 正常退出，stdout 正常）；OpenCode 的 `run` 不干净退出，原计划改 `serve` HTTP 模式，**后经 ADR-0021 spike 证伪，最终改用 `opencode acp` stdio JSON-RPC 模式**。

### 为谁解决什么问题

- Coordinator 调度宿主时需要一个统一抽象，不关心各 CLI 协议差异；
- 需要保证宿主不具备的能力被如实报告，而非模拟伪造。

## 用户故事

- 作为 Coordinator，我希望通过统一接口调用任意宿主 CLI，以便切换 backend 不影响编排逻辑。
- 作为 Coordinator，我希望查询宿主鉴权状态，以便在调度前决定是否跳过未就绪的宿主。
- 作为开发者，我希望新增宿主只需实现一个 trait 接口，以便快速接入新 AI CLI。

## 功能描述

### 做什么

1. **Runtime trait 接口**：定义 8 个方法（StartSession/DestroySession/SendTask/Capabilities/ModelInfo/Pause/Resume/AuthStatus），已在 `pkg/runtime/runtime.go` 中定义。
2. **4 款驱动实现**：
   - Codex：`exec.Command("codex", "exec", "--skip-git-repo-check", prompt)`，stdout/stderr 分离，stderr 日志过滤，sandbox 写权限配置。
   - Claude-Code：`exec.Command("claude", "-p", prompt)`，流协议适配。
   - OpenCode：启动 `opencode acp --pure` 子进程，通过 stdio JSON-RPC（Agent Client Protocol 2025-03-26）调用——发送 initialize/session/new/session/prompt，从 stdout 流式读取 notifications/session/update 事件（ADR-0021 选项 B；原 `serve` HTTP 假设已被 spike 证伪）。
   - DeepSeek-DSH：待安装后补充验证，预期 ACP v1 stdio JSON-RPC。
3. **事件流**：`SendTask` 返回 `<-chan Event`，事件类型 start/message/tool_call/done/error。
4. **能力差异处理**：`Capabilities()` 返回真实能力矩阵，不支持的能力返回错误而非伪造。例如 OpenCode 不支持 Resume → `Resume()` 返回 error。
5. **鉴权**：`AuthStatus()` 查询宿主是否就绪，不持有密钥（ADR-0005）。
6. **Registry**：`host.NewRegistry()` 注册 4 款驱动，提供 `Get(backend)` 和 `AuthCheck(ctx, backend)`。

### 各宿主调用模式（依据 ADR-0007）

| 宿主 | 调用模式 | 进程管理 | 事件获取 |
|---|---|---|---|
| Codex | `exec` 子进程，同步等待 | 单次 exec | stdout 流式 |
| Claude-Code | `-p` 非交互子进程 | 单次 exec | stdout 流式 |
| OpenCode | `opencode acp` stdio JSON-RPC | 单次会话进程 | stdout 流式（notifications/session/update） |
| DeepSeek-DSH | ACP v1 stdio JSON-RPC | 长驻 stdio | JSON-RPC 流 |

## 验收标准

- [x] `Runtime` 接口有 4 个实现，全部编译通过且 `go vet` 无警告
- [x] Codex 驱动 `SendTask` 能用 `exec.Command` 启动真实 codex CLI，stdout 返回模型输出，退出码 0
- [x] Codex 驱动 stderr 中的 WARN 日志被过滤，不污染事件流
- [x] Codex 驱动支持 `--skip-git-repo-check` 和 sandbox 写权限配置
- [x] OpenCode 驱动使用 `opencode acp`（stdio JSON-RPC）模式而非 `run`/`serve` 模式（ADR-0021 选项 B，spike 证伪 `serve`）
- [x] OpenCode 驱动 `Resume()` 返回 nil（ACP 模式支持会话恢复，ADR-0021）；Capabilities().SessionResume = true
- [x] `Capabilities()` 返回的矩阵与 ADR-0007 实测一致
- [x] `AuthStatus()` 在未登录时返回 `Ready=false`，不 panic
- [x] `Registry.Get()` 对未注册的 backend 返回明确 error
- [x] `Registry.AuthCheck()` 在宿主未就绪时返回 error 且包含 Detail
- [x] `SendTask` 超时后 context 取消，子进程被 kill，无残留
- [x] 驱动层不硬编码任何 API Key（ADR-0005）

## 非目标

- 不统一抹平各宿主能力差异（不支持的能力直接报告，不模拟）。
- 不接管宿主的计费、配额、模型选择逻辑（由宿主 CLI 自行管理）。
- 不实现 Claude-Code 和 DSH 的真实调用（MVP 阶段仅 Codex 和 OpenCode 真实实现，其余为可编译 stub）。
- ~~不实现流式增量解析（MVP 阶段 stdout 一次性收集，后续迭代再改流式）。~~ **已实现（P4，ADR-0019）：codex + opencode driver 用 StdoutPipe + Scanner 边读边 yield EventMessage。**

## 依赖与约束

- 前置依赖：A1 Spike 已完成（ADR-0007），确认调用模式可行性。
- 技术约束：
  - Windows 上 CLI 是 `.cmd` 包装脚本，`exec.Command` 需正确处理。
  - Codex stderr 噪声大，需白名单/黑名单过滤。
  - ~~OpenCode serve 模式的 API 格式需补充调研。~~ **已调研（ADR-0021 spike）：serve 是 Web UI 非 JSON API；改用 `opencode acp` stdio JSON-RPC，协议为 Agent Client Protocol 2025-03-26。**
  - 子进程超时需 `context.WithTimeout` + 进程 kill。
- 风险与缓解：
  - 模型端 503（ADR-0007 实测遇到）→ 驱动层实现重试（max 3 次）+ 超时降级。
  - 宿主 CLI 版本变更 → 驱动层隔离变更，Coordinator 不感知。

## 关联

- 实现计划：`docs/plan/mvp-preparation.md`（MVP 实现从此 spec 开始）
- 相关 ADR：ADR-0001（Go 语言）、ADR-0005（不持有密钥）、ADR-0007（CLI 可行性）
- 契约：`pkg/runtime/runtime.go`（Runtime trait）、`schemas/agent-card.json`