# P4 — 宿主 stdout 流式增量解析

> 决策依据：`docs/adr/ADR-0019-p4-streaming-stdout.md`
> 目标版本：v0.6.0
> 前置依赖：方向 C（v0.5.0，已落地）
> 后续：P6（ACP v1）→ P5（OpenCode serve API）

## 0. 范围

**做**：
1. codex driver：`cmd.Run()` 一次性收 stdout → 改为 `cmd.StdoutPipe()` + `bufio.Scanner`，边读边 yield `EventMessage`，子进程退出后发 `EventDone`。
2. opencode driver：同样改流式，且 JSON Lines 天然有行边界，每行 JSON 解析后立即发对应 `EventMessage`/`EventToolCall`，不等全部输出完。
3. 保证超时/e0a下文取消时子进程被 kill，管道关闭，无残留。
4. 保负向兼容：`Event` channel 缓冲大小不变，消费端不改（旧 `EventMessage` 聚合语义保持不变）。

**不做**：
- 不改 gRPC streaming 层（`RunPipelineStream` 已实现）。
- 不动 OpenCode serve HTTP API（属 P5）。
- 不解析 codex stdout 的语义结构（自然语言无结构化边界，按 Scanner buffer 分块 yield 即可）。
- 不动 Claude / DSH driver（仍为 stub）。

## 1. 当前状态（缺口分析）

- `internal/host/codex/driver.SendTask` 用 `cmd.Run()` + `bytes.Buffer` 一次性收 stdout，事件在子进程结束后才发。
- `internal/host/opencode/driver.SendTask` 同样一次性收，然后才解析 JSON Lines。
- 结果：TUI/gRPC streaming 要等 codex 跑完才看到输出，不是实时。
- `runtime.Event` channel 已是 buffered（codex cap=4，opencode cap=8），架构上支持流式，只是 driver 没用。

## 2. 实施步骤

| # | 任务 | 产出物 | 估时 | 关键路径 |
|---|---|---|---|---|
| P4.1 | codex driver 流式重构：StdoutPipe + Scanner yield | `internal/host/codex/codex.go` | 40 min | ✓ |
| P4.2 | opencode driver 流式重构：逐行 JSON 解析 yield | `internal/host/opencode/opencode.go` | 30 min | ✓ |
| P4.3 | 流式单测：stub binary + 真实模拟 stdout 分段 | `internal/host/codex/stream_test.go` | 30 min | ✓ |
| P4.4 | ADR-0019 + 计划 | `docs/adr/ADR-0019-p4-streaming-stdout.md` | 15 min | ✓ |
| P4.5 | CHANGELOG + v0.6.0 tag | `CHANGELOG.md` + git tag | 10 min | 末 |

总估时 ~2 小时。

## 3. 验收标准

- [ ] codex driver `SendTask` 用 `StdoutPipe` + `Scanner` yield，不再 `cmd.Run()` 一次性收
- [ ] opencode driver 逐行 JSON 解析 yield，不等全部输出完
- [ ] `TestCodex_SendTask_StreamingEvents`：模拟 stdout 分段，验证多个 `EventMessage` 在 `EventDone` 前到达
- [ ] `TestOpenCode_SendTask_StreamingJSONLines`：逐行 JSON 验证流式事件
- [ ] 超时/e0a下文取消时子进程 kill，管道关闭无残留
- [ ] 向后兼容：旧 `EventMessage` 聚合语义不变，现有 scheduler/coordinator 测试不破坏
- [ ] `go build ./...` + `go test ./...` 全绿
- [ ] 无 umadev/umacloud/goder.ai 字样

## 4. 不在范围

- OpenCode serve HTTP API -> P5
- ACP v1 `session/newTask` -> P6
- Claude / DSH 真实 exec -> 方向 B 后续
- gRPC streaming 层（已实现）

## 5. 风险与缓解

| 风险 | 缓解 |
|---|---|
| Scanner 默认 buffer 小，长行截断 | Scanner.Buffer 提升到 1MB |
| 管道写入快于读取时阻塞 | channel cap 足够大（codex 16）+ 非阻塞写（select default 丢弃最旧）或 blocking（当前选 blocking） |
| 流式改动破坏现有非流式测试 | 聚合语义保持不变，现有测试不改 |
| stub binary 输出太少不易模拟分段 | 新增多段 stub 脚本 `testdata/stubbin/codex-stream` |

## 6. 文件清单

| 操作 | 路径 |
|---|---|
| 改写 | `internal/host/codex/codex.go` |
| 改写 | `internal/host/opencode/opencode.go` |
| 新建 | `internal/host/codex/stream_test.go` |
| 可选 | `testdata/stubbin/codex-stream` 多段输出 stub |
| 新建 | `docs/adr/ADR-0019-p4-streaming-stdout.md` |
| 修改 | `CHANGELOG.md` + tag v0.6.0 |