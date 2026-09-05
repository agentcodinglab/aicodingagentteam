# ADR-0013: RAG 知识库与项目记忆接入 Director 编排闭环

> 状态：已接受
> 日期：2026-09-05
> 关联规范：`docs/spec/rag-knowledge-memory-integration.md`
> 关联 ADR：ADR-0005（本地优先/不持有密钥）、ADR-0009（质量门确定性）

## 背景

知识库（`internal/knowledge`）与项目记忆（`internal/memory`）的包级实现早已完成且测试覆盖 87–91%，但 Director 编排核心从未调用它们。`newDirector` 中 `_ = knowledge.New(false)` 和 `_ = memory.New(...)` 是丢弃实例的死代码——创建了 Engine 和 Store 却从未接线到编排流。

这导致：
- 调度宿主 CLI 前无 RAG 上下文注入，生成质量受限。
- 每次会话从零开始，历史失败教训不沉淀、不召回。
- 质量门无记忆支撑，无法从历史 pitfall 中预防同类问题。

## 决策

1. **Director 结构体扩展**：新增 `knowledge *knowledge.Engine` 和 `memory *memory.Store` 可选字段，通过 `WithKnowledge`/`WithMemory` Option setter 注入。nil 时自动跳过，保持向后兼容。

2. **Handle 五层流接入**：
   - ① Route 之后：`enhanceWithKnowledge` — 检索 top-5 chunk 路径注入 intent 上下文。
   - ② Plan 阶段：`recallMemory` — RecallFacts 召回已知事实，日志记录。
   - ⑤ Finalize 之后：`captureMemory` — 成功捕获 fact，失败捕获 pitfall。

3. **Fail-safe 原则**：所有 RAG/记忆操作失败时降级跳过，不 panic、不阻塞主编排流。这保证确定性优先原则不变——RAG 是增强而非阻断。

4. **CLI `knowledge` 子命令**：新增 `index`/`search`/`demo` 三个子命令，`demo` 端到端演示索引→检索→记忆写入→召回闭环。

5. **不自动注入宿主 prompt**：MVP 阶段 RAG 检索结果仅日志记录 + Delivery 携带，不注入宿主 CLI 的 prompt，保持编排-执行分离原则（AGENTS.md 第 5 节第 1 条）。

## 后果

- **正面**：知识库与记忆从死代码变为编排闭环一等公民，RAG 检索和记忆召回/捕获全链路可用。
- **正面**：端到端 demo 可验证（`knowledge demo`），无需真实 codex binary 即可展示 RAG + memory 闭环。
- **正面**：codex driver 覆盖率从 83.9% 提升至 94.9%（补齐 7 个 0% 函数测试）。
- **正面**：CI golangci-lint-action 从 v8 升级到 v9，消除 Node 20 deprecation warning。
- **负面**：Director 多了两个可选依赖，构造器链更长；但 nil-safe 设计保证不增加认知负担。
- **风险**：索引大目录耗时 → demo 用小目录；生产索引由用户显式 `knowledge index` 触发。