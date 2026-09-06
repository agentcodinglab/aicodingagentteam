# ADR-0007: 宿主 CLI 非交互调用可行性 Spike 结论

> 状态：已接受（OpenCode 部分结论已被 ADR-0021 证伪并修订——`serve` 非 JSON API，改用 `opencode acp`）
> 日期：2026-09-02
> 关联计划：`docs/plan/mvp-preparation.md` A1

## 背景

AiCodingAgentTeam 的架构核心假设是「能通过程序化方式（Go exec.Command）调用外部 AI 编码 CLI 的非交互模式」。在投入 MVP 实现前，需验证这一假设是否成立。若不成立，整个宿主驱动层设计需改方案（如 pty/expect 或 REST API）。

本机已安装两款 CLI：
- `codex`（@openai/codex v0.149.0）路径 `C:\nvm4w\nodejs\node_cache\codex.cmd`
- `opencode`（opencode-ai v1.18.18）路径 `C:\nvm4w\nodejs\node_cache\opencode.cmd`

claude-code 和 deepseek-dsh 未安装，待后续补充验证。

## 决策

### Codex — 非交互调用可行 ✅

- **子命令**：`codex exec [PROMPT]` 专为非交互执行设计
- **Go exec.Command 调用**：成功，26s 完成，stdout 输出正常，退出码 0
- **关键参数**：
  - `--skip-git-repo-check`：跳过 git 仓库信任检查（非 git 目录也可用）
  - `-m <MODEL>`：指定模型
  - stdin：未提供 prompt 参数时从 stdin 读取，支持管道
  - `-c key=value`：覆盖配置
- **能力**：支持 `exec resume`（会话恢复）、`exec fork`（会话分叉）
- **注意事项**：
  - stderr 输出大量 WARN 日志（插件同步、模型元数据），需过滤
  - sandbox 默认 read-only，写文件需配置 `sandbox_permissions`
  - 退出后进程干净结束，无残留

### OpenCode — 非交互调用部分可行 ⚠️

- **子命令**：`opencode run [message]`
- **Go exec.Command 调用**：stdout 成功输出内容（`Hello! 👋`），但进程不主动退出，110s 后超时
- **问题**：`run` 命令在输出响应后仍保持运行（可能等待后续输入），不适合 exec.Command 同步等待模式
- **替代方案**：
  - 方案 A：`opencode serve` 启动 headless HTTP 服务器，通过 API 调用（推荐）
  - 方案 B：`opencode run --auto` + 强制超时 kill 进程
  - 方案 C：使用 `opencode acp`（ACP 协议 server 模式，stdio JSON-RPC）
- **结论**：OpenCode 宿主驱动应采用 `serve` 模式（HTTP API）而非 `run` 模式
  
> ⚠️ SUPERSEDED by ADR-0021：spike 证伪 `opencode serve` 是 Web UI 而非 JSON API（v1.18.18）。OpenCode 驱动最终改用 `opencode acp`（stdio JSON-RPC, ACP 2025-03-26）。

### Claude-Code / DeepSeek-DSH — 待验证 ⏳

本机未安装，根据文档调研：
- Claude-Code：支持 `claude -p "prompt"` 非交互模式（已知行为）
- DeepSeek-DSH：需安装后补充验证

## 后果

### 正面后果

1. **架构假设成立**：Go exec.Command 可程序化调用 AI 编码 CLI，宿主驱动层设计无需推翻
2. **Codex 可作为 MVP 首个实现的宿主**：非交互模式完整可用
3. **明确了各宿主的适配模式差异**，为 Runtime trait 各实现提供指导

### 负面后果与风险

1. **OpenCode 不能用简单 exec.Command**，需实现 HTTP serve 模式的驱动，增加复杂度
2. **Codex stderr 噪声大**，需在驱动层过滤日志、提取有效输出
3. **模型端不稳定**（测试中遇到 503），宿主驱动需实现重试与超时降级
4. **Windows 上 CLI 是 .cmd 包装脚本**，exec.Command 需注意 Windows 上 .cmd 调用方式

### 后续需关注的事项

1. 安装 claude-code 和 dsh 后补充 spike
2. 宿主驱动实现时按此结论选择调用模式
3. codex exec 的 sandbox 写权限配置需进一步验证
4. OpenCode serve 模式的 API 格式需调研记录

## 验证产物

- `spike/host-cli/main.go` — Go spike 程序
- `spike/host-cli/go.mod` — spike 模块
- 运行日志见本 ADR 背景 section