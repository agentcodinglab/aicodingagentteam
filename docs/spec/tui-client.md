# 功能：TypeScript TUI 客户端（TUI Client）

> 文件名：`docs/spec/tui-client.md`
> 遵循模板：`aicoding_docs/docs/templates/SPEC.md`
> 关联 ADR：ADR-0001

---

## 背景与目标

TUI 客户端是开发者与 AiCodingAgentTeam 交互的终端界面，独立 npm 包分发，通过 gRPC/WebSocket 连接 Coordinator。架构设计已在 `02-技术架构设计.md` 第 12 节定义。

### 为谁解决什么问题

- 开发者需要一个交互式终端界面驱动编排流水线；
- 需要实时查看任务进度、Agent 状态、审计日志。

## 用户故事

- 作为开发者，我希望在终端输入 `/run "需求"` 启动完整流水线，以便快速驱动 AI 编码。
- 作为开发者，我希望看到任务 DAG 的实时进度（哪个角色在执行、哪个门禁暂停），以便掌握流程状态。
- 作为开发者，我希望通过 `/plan` 查看/编辑 DAG 计划，以便干预任务编排。

## 功能描述

### 做什么

1. **技术栈**
   - 渲染：Ink（React for CLI），TypeScript 原生
   - 通信：gRPC-Web / WebSocket（实时双向流式）
   - 状态管理：Zustand
   - 分发：npm 包 `aicodingagentteam-tui`，`npx` 直接运行

2. **Slash 命令系统**

   | 分类 | 命令 | 功能 |
   |---|---|---|
   | 流程控制 | `/run [需求]` | 启动完整流水线 |
   | | `/goal` | 目标模式 |
   | | `/quick [修改]` | 轻量快速编辑 |
   | | `/plan` | 查看/编辑 DAG 计划 |
   | | `/continue` | 继续暂停的任务 |
   | | `/revise` | 修订文档 |
   | 宿主切换 | `/backend [name]` | 切换宿主 CLI |
   | 质量检查 | `/verify` | 执行质量校验 |
   | | `/report` | 输出合规报告 |
   | 预览 | `/preview` | 启动前端预览 |
   | 知识管理 | `/knowledge add` | 添加知识库 |
   | | `/knowledge list` | 列出知识库 |
   | 记忆管理 | `/memory on/off` | 记忆开关 |
   | | `/memory show` | 查看记忆 |
   | 其他 | `/exit` | 退出 TUI |

3. **实时状态展示**
   - 任务 DAG 可视化（节点状态：pending/running/completed/failed/parked）
   - Agent 状态面板（哪个角色在执行、进度百分比）
   - 审计日志流（滚动显示 tool_call / a2a_message / verify 事件）
   - 门禁暂停提示（docs_confirm / preview_confirm 等待用户确认）

4. **通信协议**
   - 通过 `proto/aicodingagentteam.proto` 的 gRPC API 连接 Coordinator
   - `RunPipelineStream` 返回流式 `ProgressEvent`，TUI 实时渲染
   - 环境变量配置连接地址：`AICODINGAGENTTEAM_HOST` + `AICODINGAGENTTEAM_PORT`

5. **安装与运行**
   ```bash
   npm install -g aicodingagentteam-tui
   aicodingagentteam-tui                           # 连接本地 Coordinator
   AICODINGAGENTTEAM_HOST=x.x.x.x npx aicodingagentteam-tui  # 连接远程
   ```

## 验收标准

- [x] `npx aicodingagentteam-tui` 能启动 TUI 并连接本地 Coordinator :8080
- [x] `/run "需求"` 能触发 `RunPipelineStream` gRPC 调用
- [x] TUI 实时显示 `ProgressEvent` 流（节点状态变化）
- [x] `/quick "修改"` 能触发 `QuickEdit` gRPC 调用
- [x] `/verify` 能触发 `Verify` gRPC 调用并显示 Score + Blocking
- [x] `/plan` 能显示当前 DAG 节点列表和状态
- [x] `/continue` 能恢复 parked 任务
- [x] `/backend [name]` 能切换宿主 CLI
- [x] 门禁暂停时 TUI 显示等待提示，`/continue` 后继续
- [x] Coordinator 不可达时 TUI 显示明确错误而非崩溃
- [x] TUI 代码遵循 `aicoding_docs/docs/standards/languages/typescript.md`
- [x] npm 包可发布到 npm registry

## 非目标

- 不实现图形界面（TUI 是终端文本界面，非 Web/GUI）。
- 不实现 DAG 的拖拽编辑（MVP 阶段 `/plan` 只读展示，后续迭代加编辑）。
- 不实现 TUI 内直接编辑代码（代码编辑由宿主 CLI 在工作目录执行）。
- 不实现离线模式（TUI 必须连接 Coordinator，无离线能力）。

## 依赖与约束

- 前置依赖：`proto/aicodingagentteam.proto`（gRPC 契约）、Coordinator gRPC 服务可用
- 技术约束：Node.js ≥ 18；Ink 需要 React；gRPC-Web 需要 protobuf 编译
- 风险与缓解：
  - gRPC-Web 兼容性 → 可降级 WebSocket
  - 终端渲染性能 → Ink 虚拟滚动 + 限流更新

## 关联

- 实现计划：`docs/plan/mvp-preparation.md`
- 相关 ADR：ADR-0001（Go + TS 双语言）
- 契约：`proto/aicodingagentteam.proto`
- 规范：`aicoding_docs/docs/standards/languages/typescript.md`
- 架构设计：`02-技术架构设计.md` 第 12 节
- 源码：`tui/`（待创建）