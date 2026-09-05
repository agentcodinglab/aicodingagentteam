# 方向 D — 端到端 RAG/知识库 Demo (Lv 4-5 特性)

> 决策依据：`docs/adr/ADR-0017-direction-d-rag-demo.md`
> 目标版本：v0.4.0
> 前置依赖：方向 A（ADR-0016，已落地）
> 后续：方向 C → P4/P5/P6

## 0. 范围

**做**：把 `aicodingagentteam knowledge demo` 从"手工拼装 knowledge + memory"升级为"通过 Director.Handle 走完整 5 层流水线，并产出可验证报告"。

**不做**：CI 装真实 codex binary（属方向 C）；新功能特性（属方向 B / P4-6）。

## 1. 当前状态

- `internal/knowledge.Engine` + `internal/memory.Store` 已实现且覆盖率达标
- Director.Handle 已通过 `WithKnowledge` / `WithMemory` 选项把知识检索、记忆召回、记忆捕获接入 5 层流水线（ADR-0013）
- `cmd/aicodingagentteam/knowledge demo` 已存在，但**只手工调用 knowledge/memory，不走 Director.Handle**
- `internal/coordinator/director_rag_test.go` 有 5 个针对性单测，但**没有端到端集成测试覆盖完整 demo 链路**

## 2. 实施步骤

| # | 任务 | 产出物 | 估时 | 关键路径 |
|---|---|---|---|---|
| D1 | 重写 `knowledgeDemo` 让它通过 Director.Handle | `cmd/aicodingagentteam/main.go` | 30 min | ✓ |
| D2 | 让 demo 在**当前仓库根目录**上索引（而非临时目录） | `main.go` | 15 min | ✓ |
| D3 | 产出 demo 报告（JSON / Markdown）写到 `.aicodingagentteam/demo-report.md` | `main.go` + 报告 schema | 30 min | ✓ |
| D4 | 新增 e2e 集成测试覆盖 demo 完整链路 | `internal/coordinator/demo_e2e_test.go` | 40 min | ✓ |
| D5 | ADR-0017 记录决策 | `docs/adr/ADR-0017-direction-d-rag-demo.md` | 15 min | ✓ |
| D6 | demo-report.md 加进 governance summary（governance.yml 增量） | `.github/workflows/governance.yml` | 15 min | ✓ |
| D7 | v0.4.0 tag + goreleaser | git tag + release | 5 min | 末 |

总估时 ~2.5 小时，单一 commit 落地（`feat(demo): end-to-end RAG demo through Director.Handle`）。

## 3. 验收标准

- `aicodingagentteam knowledge demo` 退出码 0
- demo 报告写入 `.aicodingagentteam/demo-report.md`，含：
  - 索引文档数、检索 top-K chunk、recall facts 数、delivery verdict
  - 真实使用当前仓库作为索引源（非临时目录硬编码文件）
- 集成测试 `TestKnowledgeDemo_E2E_ThroughDirector` 跑通
- `governance.yml` 中 godocgen job 把 demo-report 作为 artifact 上传
- v0.4.0 release 页面含 ADR-0017 引用

## 4. 不在范围 / 推迟到下阶段

- CI 装真实 codex binary 跑端到端 → 方向 C
- 新宿主（claude/dsh 真实 exec） → 方向 B
- 进阶特性（流式解析 / OpenCode serve API / ACP v1） → P4/P5/P6

## 5. 风险与缓解

| 风险 | 缓解 |
|---|---|
| 索引当前仓库体量大（>1000 文件）触发超时 | 加 `indexMaxFiles` 上限（默认 500）+ timeout 30s |
| demo 报告写入当前目录污染 git | 报告路径固定在 `.aicodingagentteam/`（已 gitignore） |
| Director.Handle 走真实 schedule 时缺宿主驱动 | demo 用最小 stub backend（codex mock），不触发真实子进程 |
| 与 governance workflow 集成可能引入时延 | demo-report 仅作 artifact，不进 lint / test gate |

## 6. 文件清单（预计）

| 操作 | 路径 |
|---|---|
| 改写 | `cmd/aicodingagentteam/main.go`（`knowledgeDemo` 函数） |
| 新建 | `internal/coordinator/demo_e2e_test.go` |
| 新建 | `docs/adr/ADR-0017-direction-d-rag-demo.md` |
| 修改 | `.github/workflows/governance.yml`（godocgen job 加 demo-report artifact） |
| 新建 | `.aicodingagentteam/demo-report.md`（产物，已 gitignore） |
