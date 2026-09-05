# 功能：RAG 知识库与项目记忆编排集成（Director RAG/Memory Wiring）

> 文件名：`docs/spec/rag-knowledge-memory-integration.md`
> 遵循模板：`aicoding_docs/docs/templates/SPEC.md`
> 关联 ADR：ADR-0005、ADR-0009
> 关联 spec：`docs/spec/knowledge-memory.md`（包级规格）、`docs/spec/coordinator.md`

---

## 背景与目标

知识库（`internal/knowledge`）与项目记忆（`internal/memory`）的包级实现已完成且有 87–91% 测试覆盖，但 **Director 编排核心从未调用它们**。当前 `newDirector` 中 `_ = knowledge.New(false)` 和 `_ = memory.New(...)` 是丢弃实例的死代码，Director 结构体无 knowledge/memory 字段，`Handle` 五层流（Route→Plan→Schedule→Verify→Finalize）中无任何 RAG 检索或记忆召回/捕获。

本规格将知识库与记忆 **接入 Director 编排闭环**，使 RAG 检索与记忆召回/捕获成为编排流的一等公民，并交付一个端到端 demo。

### 为谁解决什么问题

- **Coordinator**：调度宿主 CLI 前缺乏项目上下文注入，生成质量受限。
- **用户**：每次会话从零开始，历史失败教训不沉淀、不召回。
- **质量门**：无记忆支撑，无法从历史 pitfall 中预防同类问题。

## 用户故事

- 作为 Coordinator，在 Route 阶段后检索知识库相关文档/代码片段，注入 Plan 上下文。
- 作为 Coordinator，在 Plan 阶段召回相关 recipe（历史方案）作为调度建议。
- 作为 Coordinator，在 Finalize 后捕获本次交付的 fact（环境事实）与 pitfall（若失败）。
- 作为用户，运行 `aicodingagentteam knowledge` 命令索引目录并 demo 检索 + 记忆写入/召回闭环。

## 功能描述

### 做什么

#### 1. Director 结构体扩展

Director 新增两个可选字段：

```go
type Director struct {
    router   *router.Router
    planner  *planner.Planner
    sched    *scheduler.Scheduler
    gate     *qualitygate.Engine
    bus      a2a.Bus
    knowledge *knowledge.Engine  // 可选；nil 时跳过 RAG
    memory   *memory.Store      // 可选；nil 时跳过记忆
}
```

- 提供 `WithKnowledge(*knowledge.Engine)` 和 `WithMemory(*memory.Store)` Option 模式 setter。
- 现有 `New` / `NewWithBus` 构造器保持兼容（knowledge/memory 默认 nil）。

#### 2. Handle 流接入

在 `Handle(ctx, req)` 中：

| 阶段 | 接入点 | 行为 |
|---|---|---|
| ① Route 之后、② Plan 之前 | **RAG 检索** | 若 `d.knowledge != nil`，对 `req.Message` 执行 `Retrieve(ctx, req.Message, 5)`，将 top-k chunk 路径拼接到 intent 的上下文（不改路由结果，只增强上下文）。 |
| ② Plan 时 | **记忆召回** | 若 `d.memory != nil`，`RecallFacts(ctx)` 取已有 facts；若 intent 有栈信息则 `Recall(ctx, stack)` 取 recipe。打印召回的 facts 数量到审计日志（不阻塞主流程）。 |
| ⑤ Finalize 之后 | **记忆捕获** | 若 `d.memory != nil`：成功时捕获 fact（`backend`/`score`/`passed`）；失败时捕获 pitfall（`plan-id`+`score`+`check failures`）。 |

- 所有 RAG/记忆操作失败时 **降级跳过**，不影响主编排流（确定性优先原则不变）。
- RAG 检索结果不自动注入宿主 CLI 的 prompt（MVP 阶段仅日志记录 + Delivery 携带），避免破坏编排-执行分离原则。

#### 3. CLI 命令：`knowledge`

新增 `aicodingagentteam knowledge` 子命令：

```
aicodingagentteam knowledge index [dir]   # 索引目录（默认 .）
aicodingagentteam knowledge search "query" # 检索 top-5
aicodingagentteam knowledge demo          # 端到端 demo：索引自身→检索→写记忆→召回记忆
```

#### 4. 端到端 Demo

`knowledge demo` 执行：
1. 创建临时知识目录，写入 2 个 `.go` 文件
2. `IndexDirectory` 索引
3. `Retrieve("router", 3)` 检索并打印结果
4. `memory.Capture(fact)` 写入一条 fact
5. `memory.RecallFacts()` 召回并打印
6. 打印 `DocCount` 和 `IsCloudEmbed`

## 验收标准

- [ ] Director 结构体含 `knowledge`/`memory` 可选字段
- [ ] `WithKnowledge`/`WithMemory` setter 可用
- [ ] `Handle` 在 Route 后执行 RAG 检索（knowledge 非 nil 时）
- [ ] `Handle` 在 Plan 时召回 facts（memory 非 nil 时）
- [ ] `Handle` 在 Finalize 后捕获 fact/pitfall（memory 非 nil 时）
- [ ] RAG/记忆操作失败时降级跳过，不 panic、不阻塞主流程
- [ ] `newDirector` 不再有 `_ = knowledge.New` / `_ = memory.New` 死代码
- [ ] `knowledge index/search/demo` 子命令可用
- [ ] `knowledge demo` 端到端输出检索结果 + 记忆写入/召回
- [ ] `go test ./internal/coordinator/...` 通过且覆盖不降
- [ ] `go test ./internal/host/codex/...` 覆盖率 ≥ 90%
- [ ] golangci-lint 0 issue
- [ ] CI 三 job 全绿

## 非目标

- 不实现向量检索（MVP 纯 BM25，向量后续迭代）
- 不实现 HyDE 查询扩展
- 不实现 RAG 结果自动注入宿主 prompt（保持编排-执行分离）
- 不实现记忆跨项目共享（ADR-0005 禁止）
- 不改变 Route 的路由决策逻辑（只增强上下文）

## 依赖与约束

- 前置：`internal/knowledge`（已实现）、`internal/memory`（已实现）
- 约束：RAG/记忆操作须 fail-safe，不得成为主流程的阻断点
- 风险：索引大目录耗时 → demo 用小目录；生产索引由用户显式触发

## 关联

- 实现计划：`docs/plan/rag-knowledge-memory-integration.md`
- 相关 ADR：ADR-0005（本地优先）、ADR-0009（质量门确定性）
- 源码：`internal/coordinator/director.go`、`cmd/aicodingagentteam/main.go`