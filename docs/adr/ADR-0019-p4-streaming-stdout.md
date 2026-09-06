# ADR-0019：P4 — 宿主 stdout 流式增量解析

> 状态：Accepted
> 日期：2026-09-06
> 关联：ADR-0007（宿主 CLI 可行性）、ADR-0018（方向 C 真实宿主 e2e）、ADR-0013（RAG/Memory wiring）
> 计划：`docs/plan/p4-streaming-stdout.md`

## 背景

方向 C 落地后，编排引擎 → 驱动层 → 子进程 → 事件流 → 工件落盘 链路首次被 stub binary e2e 测试覆盖。但驱动层仍一次性收集 stdout：

- codex driver 用 `cmd.Run()` + `bytes.Buffer`，子进程跑完才发 `EventMessage`。
- opencode driver 同样一次性收，再解析 JSON Lines。

结果：TUI/gRPC streaming 要等 codex 整个任务跑完才能看到输出，不是实时。`runtime.Event` channel 架构上支持流式（buffered），但驱动没用。

## 决策

重构 codex 与 opencode 驱动的 `SendTask`，采用 `cmd.StdoutPipe()` + `bufio.Scanner` 边读边 yield：

1. codex：Scanner 按默认行分割，每读到一段就 yield `EventMessage`；子进程退出后发 `EventDone`。不解析语义（自然语言无结构化边界）。
2. opencode：JSON Lines 天然有行边界，每行 JSON 解析后立即发对应 `EventMessage`/`EventToolCall`，不等全部输出完。
3. 超时/e0a下文取消时 `cmd.Process.Kill()` 子进程，管道关闭无残留。
4. 向后兼容：`EventMessage` 聚合语义不变（消费端拼接后仍可聚合得完整 stdout），现有 scheduler/coordinator 测试不破坏。

## 备选方案

- 保持一次性收集：不满足实时体验，否决。
- 只流式 opencode，codex 保持一次性：两者架构不一致，维护成本高，否决。

## 后果

- 正面：TUI/gRPC streaming 实时显示宿主输出；超长任务可提前看到进度；超时可提前识别卡住。
- 负面：codex stdout 是自然语言流，分块边界是 Scanner 行结束，不是语义边界（但对实时体验足够）。
- 中性：驱动代码重构，但 `Event` channel 接口不变，消费端不需改。

## 验收

- `TestCodex_SendTask_StreamingEvents` + `TestOpenCode_SendTask_StreamingJSONLines` 跑通
- 现有 scheduler/coordinator 测试无回归
- `go build ./...` + `go test ./...` 全绿