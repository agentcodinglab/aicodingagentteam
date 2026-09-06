# ADR-0018：方向 C — 真实宿主驱动端到端验证（Codex stub binary）

> 状态：Accepted
> 日期：2026-09-06
> 关联：ADR-0005（不持有密钥）、ADR-0007（宿主 CLI 可行性）、ADR-0013（RAG/Memory wiring）、ADR-0017（方向 D）
> 计划：`docs/plan/direction-c-real-host-e2e.md`

## 背景

方向 D 落地了 RAG + memory demo，但 demo 用 stub backend，从不 fork 真实 codex 子进程。代码审查发现：

- `internal/host/codex.Driver.SendTask` 能调真实 `codex exec`，但**从未被调用**。
- `internal/scheduler/Scheduler.Execute` 的 writer 节点只 acquireWriteLock 后直接列 `node.ArtifactsOut`，**绕过了宿主驱动**。
- `internal/coordinator/Director` 没有 driver 注入口。
- 现有 E2E 测试 `Backend: codex` 走 router 快速路径（PlanID=quick），没触发真实子进程。

即编排引擎 → 驱动层 → 子进程 → 事件流这条链路从未被任何测试覆盖，是真实的覆盖盲区。

## 决策

采用 **codex stub binary 方案**：

1. 在仓库内放一个 codex stub 脚本（`testdata/stubbin/codex` Unix + `codex.cmd` Windows），模拟 `codex exec --skip-git-repo-check <prompt>`：往 stdout 写一段含工件引用的输出、退出 0。
2. 给 Scheduler 注入 driver（`WithDriver` option + `SetDriver` 方法），让 writer 节点真正调用 `driver.SendTask`，把 stdout 作为工件落盘到 `.aicodingagentteam/host/<nodeID>.txt`，并把工件路径记入 Verdict。
3. 给 Director 暴露 `WithDriver` DirectorOption，向下传给 scheduler。
4. CI 用 stub binary 跑真实 `Director.Handle → scheduler → codex driver → stub binary` 端到端，governance.yml 新增 `host-e2e` job（soft-fail）。
5. driver=nil 时降级为旧行为（直接列 artifacts），保证现有测试不破坏。
6. 本地真实 codex：`scripts/e2e-real-codex.ps1` + 文档说明，不在 CI 注入 API key。

## 备选方案

- **CI 注入 OpenAI_API_KEY 跑真实 codex exec**：覆盖最真实，但违反 ADR-0005「不持有密钥」精神（需补 ADR 豁免），且有成本/503/速率限制风险，CI 不稳定。否决。
- **跳过方向 C 直接进 P4/P5/P6**：编排→驱动→子进程链路仍是盲区，进阶特性（流式解析）会建在未验证的地基上。否决。

## 后果

- 正面：编排引擎 → 驱动层 → 子进程 → 事件流 → 工件落盘 全链路首次被测试覆盖；CI 稳定可复现；不碰密钥原则。
- 负面：stub binary 输出与真实 codex 不完全一致（但 driver 只收 stdout 不解析具体内容，影响有限）；真实 codex 的 503/超时只靠本地脚本手动验证。
- 中性：scheduler 行为对 driver=nil 保持兼容，现有 E2E 测试无需改动。

## 与 ADR-0005 的关系

ADR-0005「不持有密钥」要求鉴权全部交底层 CLI。本方案完全符合：

- CI 的 stub binary 不需要任何 key。
- 本地真实 codex 的 key 由 codex CLI 自持，AiCodingAgentTeam 仅查询 AuthStatus，不持有。
- Coordinator/Driver 代码无任何 API key 字段或注入路径。

## 验收

- `TestScheduler_HostE2E_StubBinary` 跑通：delivery 不 park、host 工件文件非空、planned artifacts 在结果中。
- `TestScheduler_HostE2E_NilDriver_NoPanic` 跑通：无 driver 降级不 panic。
- governance.yml `host-e2e` job（soft-fail）用 stub binary 跑端到端。
- `scripts/e2e-real-codex.ps1` 本地可跑真实 codex。
- `go build ./...` + `go test ./...` 全绿。