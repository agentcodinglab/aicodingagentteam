# 方向 C — 真实宿主驱动端到端验证（Codex stub binary + 本地真实 codex）

> 决策依据：`docs/adr/ADR-0018-direction-c-real-host-e2e.md`
> 目标版本：v0.5.0
> 前置依赖：方向 D（v0.4.0）
> 后续：P4/P5/P6（流式解析 / OpenCode serve API / ACP v1）

## 0. 范围

**做**：
1. 让 Scheduler 的 writer 节点真正调用宿主 driver（`pkg/runtime.Runtime.SendTask`），把 stdout 工件落盘 + Verdict 如实记录。
2. 给 Scheduler 注入 driver（`WithDriver` option）。
3. 在 repo 内放一个 codex stub binary 脚本，模拟 `codex exec --skip-git-repo-check <prompt>`：往 stdout 写含工件引用输出、退出 0。
4. CI 用 stub binary 跑真实 `Director.Handle -> scheduler -> codex driver -> stub binary` 端到端（无 API key，不违反 ADR-0005）。
5. 本地真实 codex：加 `make e2e-codex` / 脚本 + 文档说明。

**不做**：
- 不在 CI 注入 OpenAI API key（违反 ADR-0005）。
- 不实现流式增量解析（属 P4）。
- 不动 OpenCode serve API / ACP v1（属 P5/P6）。
- 不改 Claude / DSH stub driver。

## 1. 当前状态（缺口分析）

- `internal/host/codex.Driver.SendTask` 调真实 codex，但从未被调用。
- `internal/scheduler/Scheduler.Execute` 的 writer 节点只 acquireWriteLock，**从不调用 driver**。
- `internal/coordinator/Director` 没有 driver 注入口。
- 现有 E2E 测试 `Backend: codex` 走 router 快路径，没触发真实子进程。

## 2. 实施步骤

| # | 任务 | 产出物 | 估时 | 关键路径 |
|---|---|---|---|---|
| C1 | Scheduler 注入 driver + writer 节点调 SendTask | `internal/scheduler/scheduler.go` | 40 min | ✓ |
| C2 | Director 暴露 driver 注入（WithDriver DirectorOption） | `internal/coordinator/director.go` | 20 min | ✓ |
| C3 | codex stub binary 脚本 | `testdata/stubbin/` | 20 min | ✓ |
| C4 | 端到端测试：stub binary 跑真实 Director.Handle | `internal/scheduler/host_e2e_test.go` | 40 min | ✓ |
| C5 | CI governance host-e2e job（soft-fail，stub binary） | `.github/workflows/governance.yml` | 20 min | ✓ |
| C6 | 本地真实 codex 脚本 + 文档 | `scripts/e2e-real-codex.ps1` + docs | 20 min | |
| C7 | ADR-0018 + 计划 | `docs/adr/ADR-0018-*.md` | 15 min | ✓ |
| C8 | CHANGELOG + v0.5.0 tag | `CHANGELOG.md` + git tag | 10 min | 末 |

总估时 ~3 小时。

## 3. 验收标准

- [ ] `Scheduler.Execute` writer 节点调用 `driver.SendTask`，stdout 工件落盘到 workspace
- [ ] driver=nil 降级为旧行为（直接列 artifacts），不 panic
- [ ] codex stub binary 接收 `exec --skip-git-repo-check <prompt>`，stdout 写含工件引用输出，退出 0
- [ ] `TestScheduler_HostE2E_StubBinary` 跑通：delivery.Passed=true，artifacts 非空
- [ ] `TestScheduler_HostE2E_NilDriver_NoPanic`：无 driver 降级不 panic
- [ ] governance.yml host-e2e job（soft-fail）用 stub binary 跑端到端
- [ ] `make e2e-codex` 本地真实 codex手动验证（文档说明）
- [ ] ADR-0018 记录 stub-binary 与 ADR-0005 不冲突论证
- [ ] v0.5.0 release 含 ADR-0018 引用
- [ ] `go build ./...` + `go test ./...` 全绿
- [ ] 无 umadev/umacloud/goder.ai 字样

## 4. 不在范围

- 流式增量解析 -> P4
- OpenCode serve HTTP API -> P5
- ACP v1 stdio JSON-RPC -> P6
- Claude / DSH 真实 exec -> 方向 B 后续

## 5. 风险与缓解

| 风险 | 缓解 |
|---|---|
| stub binary Win 上 .cmd 包装 | 提供 codex.cmd + codex 两份，test 按 runtime.GOOS 选 |
| exec.LookPath 找不到 stub | driver 已支持 WithBinary(abs path) |
| stub 输出与真实 codex 不一致 | stub 输出固定工件引用串，driver 不解析内容 |
| 真实 codex 503/超时 | 本地脚本加重试提示；CI 只用 stub |
| scheduler 改动破现有 E2E | driver=nil 降级，现有测试不注入 driver |

## 6. 文件清单

| 操作 | 路径 |
|---|---|
| 改写 | `internal/scheduler/scheduler.go` |
| 方法 | `internal/scheduler/host_e2e_test.go` |
| 修改 | `internal/coordinator/director.go`（WithDriver） |
| 新建 | `testdata/stubbin/codex` + `codex.cmd` |
| 修改 | `.github/workflows/governance.yml` |
| 新建 | `scripts/e2e-real-codex.ps1` |
| 新建 | `docs/adr/ADR-0018-direction-c-real-host-e2e.md` |
| 修改 | `CHANGELOG.md` + tag v0.5.0 |