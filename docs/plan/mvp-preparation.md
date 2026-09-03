# 计划：MVP 准备阶段（A-E）

> 关联规格：`docs/spec/`（待编写）
> 关联文档：`02-技术架构设计.md`、`03-系统设计与实施规划.md`
> 遵循模板：`aicoding_docs/docs/templates/PLAN.md`

---

> 状态：**MVP 功能实现已完成**（2026-09-03）
> A-E 全部完成，P0-P6 MVP 实现全部完成。
> 7 个 spec 加权完成率 86%，功能性验收标准全部通过。

---

## 目标

在进入 MVP 实现之前，完成风险验证（Spike）、功能规格（Spec）、决策记录（ADR）、契约固化（Proto/Schema）、工程基线（Lint/CI），确保后续 MVP 实现返工率最低。

---

## 执行顺序

```
A (Spike, 1-2天) → B (Spec, 3-4天) → C (ADR, 1天) → D (契约, 1-2天) → E (基线, 1天)
                                                                        ↓
                                                                    MVP 实现
```

> A 必须最先：验证「能否做」。B 紧随其后：定义「做什么」。C/D/E 可部分并行。全部完成后 MVP 实现变为「按 spec 填实现」。

---

## A. 风险 Spike（最高优先，验证最大不确定性）

脚手架最致命的假设是「能调用真实 CLI」。先用最小代价验证 4 个风险点是否成立。

### A1. 宿主 CLI 非交互调用可行性

- 完成标准：4 款 CLI 各用 50 行 spike 脚本试调，确认是否支持 stdin/stdout 流式管道、非 TTY 环境运行、退出码语义
- 涉及文件：`spike/host-cli/cmd_codex.go`、`spike/host-cli/cmd_claude.go`、`spike/host-cli/cmd_opencode.go`、`spike/host-cli/cmd_dsh.go`、`docs/adr/ADR-0007-host-cli-feasibility.md`
- 输出：可行性报告 + 不支持的 CLI 的替代方案（expect/pty/REST API）
- 风险：若某款 CLI 不支持非交互模式，宿主驱动设计需改方案

### A2. A2A 跨容器通信延迟

- 完成标准：起两个容器 + Redis，测 Coordinator→Agent 委派往返 p95，验证 500ms 阈值是否现实
- 涉及文件：`spike/a2a-latency/`
- 输出：延迟报告 + 阈值调整建议

### A3. 质量门禁真实执行

- 完成标准：确认容器内能否跑 go build/test、golangci-lint、宿主项目测试，产物路径如何
- 涉及文件：`spike/quality-exec/`
- 输出：执行报告 + 超时策略

### A4. 工件卷并发读写

- 完成标准：验证单写者约束在 Docker Volume 上是否可靠，是否需要文件锁
- 涉及文件：`spike/volume-concurrency/`
- 输出：并发报告 + 锁策略

---

## B. 功能规格（Spec-first）

为 7 个核心模块各写 spec，写入 `docs/spec/`，每份含可验证验收标准。

| 编号 | Spec 文件 | 内容 |
|---|---|---|
| B1 | `host-driver.md` | 宿主驱动层：Runtime trait 各方法行为契约 + 能力差异处理 |
| B2 | `coordinator.md` | 五层调度循环：每层输入/输出/判定逻辑 |
| B3 | `a2a-protocol.md` | A2A 协议：消息格式、Agent Card 发现、Verdict 结构 |
| B4 | `quality-gate.md` | 质量门禁：10 项校验 + 阈值 + 阻塞规则 |
| B5 | `governance.md` | 治理引擎：规则结构 + fail-open + 触发入口 |
| B6 | `knowledge-memory.md` | 知识库 + 记忆：检索降级 + 不跨项目共享 |
| B7 | `tui-client.md` | TS TUI：Slash 命令 + gRPC 通信 + 状态展示 |

---

## C. 关键 ADR（记录决策理由）

把隐性决策写进 `docs/adr/`，避免后续被推翻时没记录。

| 编号 | ADR 文件 | 决策 |
|---|---|---|
| C1 | `ADR-0001-go-not-rust.md` | 为何选 Go 而非 Rust |
| C2 | `ADR-0002-redis-a2a-bus.md` | 为何用 Redis Pub/Sub 做 A2A 总线 |
| C3 | `ADR-0003-container-per-role.md` | 为何容器化而非单二进制 |
| C4 | `ADR-0004-single-writer-model.md` | 为何写角色串行 |
| C5 | `ADR-0005-no-key-custody.md` | 为何不持有密钥 |
| C6 | `ADR-0006-fail-open-governance.md` | 为何治理 fail-open |

---

## D. 契约固化（接口先行）

把接口定义固化为机器可校验的契约文件。

| 编号 | 文件 | 内容 |
|---|---|---|
| D1 | `proto/aicodingagentteam.proto` | gRPC API 定义 |
| D2 | `schemas/a2a-message.json` | A2A 消息 JSON Schema |
| D3 | `schemas/agent-card.json` | Agent Card 发现格式 |
| D4 | `schemas/plan.json` | DAG 计划文件格式 |

---

## E. 脚手架补强（工程基线）

让骨架达到团队可协作基线。

| 编号 | 文件 | 内容 |
|---|---|---|
| E1 | `.golangci.yml` | lint 规则（对齐 CONSTRAINTS.md） |
| E2 | `Makefile` | 统一 build/test/lint/vet/run 入口 |
| E3 | `.github/workflows/ci.yml` | CI 流水线 |
| E4 | `*_test.go` | 各包单元测试骨架 |
| E5 | `CONTRIBUTING.md` | 协作约定 |

---

## 并行与关键路径

- 可并行：C（ADR）与 D（契约）与 E（基线）三者互不依赖
- 关键路径（串行）：A → B → MVP（A 验证可行性，B 定义规格，才能开始实现）
- A 内部：A1 必须最先（最大风险），A2/A3/A4 可并行

---

## 风险与回滚

- 风险点：A1 若发现 CLI 不支持非交互，需改用 pty/expect 方案，ADR-0007 记录
- 回滚策略：Spike 失败不回滚代码（spike 代码隔离在 spike/ 目录），仅更新设计文档