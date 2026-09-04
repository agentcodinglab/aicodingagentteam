# ADR-0012: 质量门 CheckDetails 全链透传

> 状态：已接受
> 日期：2026-09-04
> 关联规范：`docs/CONSTRAINTS.md` 确定性优先原则

## 背景

质量门 Engine 的 `Result.Details`（每项检查的 name/status/output）在传递链中多处丢失：
- `api.VerifyResponse` 缺 `Details` 字段 → scorecard 的 "Check Details" 永远为空
- `Director.QuickEdit`/`RunPipeline` 未透传 `Score`/`Details` → gRPC 客户端收到零值
- `coordinatorAdapter` 的 QuickEdit/Verify/RunPipeline 未映射 `Details` 到 pb 响应
- CLI `quick`/`run` 只打印 `passed=true/false` 不显示失败原因

这导致偶发的质量门失败无法诊断——`passed=false` 但不知哪个检查失败、为什么失败。

## 决策

1. **Delivery 增加 CheckDetails**：`types.Delivery.CheckDetails []CheckSummary`，从 `verdict.Details` 转换。
2. **proto 新增 CheckDetailInfo**：`RunPipelineResponse`/`QuickEditResponse`/`VerifyResponse` 都加 `details` 字段。
3. **全链透传**：`Director.Handle` → `Delivery.CheckDetails` → `Director.QuickEdit/RunPipeline` → `api.QuickResponse/RunResponse.Details` → `coordinatorAdapter` → `pb.*Response.Details`。
4. **CLI 输出**：`quick`/`run` 失败时打印每项检查的 `[FAIL] name: output`。
5. **输出截断放宽**：`runCheck` 输出截断从 300→2000 字符，确保失败诊断信息不被截断。

## 后果

- 正面：质量门失败原因从 server 内部贯穿到 gRPC 客户端和 CLI，不再需要重跑 verify 来诊断。
- 正面：TUI 客户端可直接渲染检查详情（`resp.details`）。
- 负面：proto 变更需重新生成 stubs 并同步 `tui/proto/` 副本。
- 排障记录：发现 npm 子进程环境缺 `LOCALAPPDATA` 导致 Go 工具链找不到 build cache（`GOCACHE is not defined`），这是之前所有"质量门飘红"的真正根因。修复方式：e2e 脚本 spawn server 时补 `LOCALAPPDATA` fallback。