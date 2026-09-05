# ADR-0017: 端到端 RAG Demo 升级为 Director.Handle 全链路

> 状态：已接受
> 日期：2026-09-05
> 关联规范：`docs/spec/rag-knowledge-memory-integration.md`、`docs/plan/direction-d-rag-demo.md`

## 背景

方向 A（ADR-0016）落地了治理门禁。项目现有三个独立能力：

1. `internal/knowledge.Engine`：BM25 检索 + 文件索引（本地优先，不依赖云端 embedding）
2. `internal/memory.Store`：fact / pitfall / lesson / recipe 持久化（JSONL）
3. `internal/coordinator.Director`：5 层流水线 Route→Plan→Schedule→Verify→Finalize

ADR-0013 已经把三者通过 `WithKnowledge` / `WithMemory` 选项接进 Director.Handle。但 CLI 命令 `aicodingagentteam knowledge demo` 仍然只**手工拼装** knowledge + memory 调用，不走 Director.Handle，无法验证完整闭环。

## 决策

### 1. demo 升级为 Director.Handle 全链路

`knowledgeDemo` 不再手工调 IndexDirectory / Retrieve / Capture / Recall，而是：

1. 在当前仓库根目录上索引（加 `indexMaxFiles` 上限防超时）
2. 构造 `Director`（带 `WithKnowledge` + `WithMemory`），用最小 stub backend 避免 fork 真实子进程
3. 调用 `Director.Handle(ctx, UserRequest{Message: "..."})`，走完整 Route→Plan→Schedule→Verify→Finalize
4. 从返回的 `Delivery` 提取 verdict、score、artifacts
5. 输出 demo 报告

### 2. demo 报告产物

报告写到 `.aicodingagentteam/demo-report.md`（已在 .gitignore），包含：

- 索引文档数 + 来源路径
- Retrieve top-K chunks（路径 + 分数）
- RecallFacts 结果
- Delivery verdict（score / passed / check details）
- 时间戳 + 版本号

报告同时输出 JSON 副本 `.aicodingagentteam/demo-report.json` 供 CI 解析。

### 3. 端到端集成测试

新建 `internal/coordinator/demo_e2e_test.go`：

- `TestKnowledgeDemo_E2E_ThroughDirector`：临时目录写 2 个 .go 文件 → 索引 → 构造 Director → Handle → 验证 Delivery 非 nil + verdict 有值
- `TestKnowledgeDemo_ReportWritten`：验证 demo-report.md 被写出且含关键字段
- `TestKnowledgeDemo_NilKnowledge_NoPanic`：nil engine 时 Handle 不 panic

### 4. governance.yml 集成

godocgen job 增加一步 `go run ./cmd/aicodingagentteam knowledge demo`，把 `.aicodingagentteam/demo-report.*` 作为 artifact 上传。**不作为硬门禁**——demo 失败只 warning。

### 5. indexMaxFiles 安全阀

`knowledge.IndexDirectory` 新增 `IndexDirectoryWithLimit(ctx, root, maxFiles int)` 重载，默认上限 500 文件。超过上限只索引前 500 个 + 在报告标注 truncated。

## 后果

### 正面

- demo 真正验证"知识检索 + 记忆召回 + 编排 + 记忆捕获"完整闭环
- demo 报告作为 CI artifact 可审计
- 外部用户能跑 `knowledge demo` 看到真实输出
- e2e 测试防止 Director.Handle 与 knowledge/memory 集成回归

### 负面 / 权衡

- demo 仍用 stub backend（不 fork 真实 codex）—— 真实宿主验证属方向 C
- `indexMaxFiles` 上限可能让大型仓库索引不完整 —— 已在报告标注 truncated
- governance.yml 增 ~20s（demo 跑一次）—— 可接受

### 风险与缓解

| 风险 | 缓解 |
|---|---|
| 索引大仓库超时 | indexMaxFiles=500 + context 30s timeout |
| demo 报告污染 git | 路径固定 `.aicodingagentteam/`（已 gitignore） |
| Director.Handle 走真实 schedule 缺宿主 | demo 用 stub backend，不触发子进程 |
| governance 集成引入时延 | demo 仅作 artifact，不进 lint/test gate |

## 后续

- v0.4.0 发布时同步宣告本 ADR
- 方向 C 将把 demo 升级为真实 codex binary 端到端
- P4/P5/P6 将在 demo 稳定后叠加流式解析、OpenCode serve、ACP v1
