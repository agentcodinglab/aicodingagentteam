# 功能：知识库与项目记忆（Knowledge & Memory）

> 文件名：`docs/spec/knowledge-memory.md`
> 遵循模板：`aicoding_docs/docs/templates/SPEC.md`
> 关联 ADR：ADR-0005

---

## 背景与目标

知识库负责检索增强（RAG），为宿主 CLI 注入项目上下文；项目记忆负责记录历史决策与失败教训，减少重复错误。两者均在 `internal/` 下搭骨架，当前为空实现。

### 为谁解决什么问题

- 宿主 CLI 缺乏项目上下文，知识库注入相关文档/代码片段；
- 需求迭代后上下文丢失，记忆库保存项目事实和历史方案；
- 团队规范难注入 AI 上下文，知识库可加载自定义知识。

## 用户故事

- 作为 Coordinator，我希望在调度宿主前检索相关知识注入上下文，以便提高生成质量。
- 作为用户，我希望记忆库记住项目的环境事实和历史失败，以便减少重复踩坑。
- 作为用户，我希望默认不上传我的代码到云端，以便保护隐私。

## 功能描述

### 做什么

#### 知识库（Knowledge）

1. **混合检索架构**（BM25 + 向量 RRF + HyDE）
   - HyDE 查询扩展：用用户需求生成假想文档，扩展检索
   - 路径 A：纯 Go BM25（CJK bigram 分词），零外部依赖，**保底必选**
   - 路径 B：本地向量模型（multilingual-e5-small f16, 224MB）或云端 embedding
   - RRF 融合两路召回结果，返回 top-k 块

2. **降级策略**
   - 云端 embedding 需双环境变量：`OPENAI_EMBED_KEY` + `AICODINGAGENTTEAM_ALLOW_CLOUD_EMBED=1`
   - 向量不可用时自动降级纯 BM25，不报错失败
   - 本地向量模型可选 feature（`vector-local`），默认关闭

3. **repo-map 符号索引**
   - 扫描项目源码符号，计算重要度排序
   - 上下文预算裁剪：优先注入高重要度符号

4. **自定义知识库**
   - `aicodingagentteam knowledge-manage add ./docs` 添加外部知识
   - 知识源：项目文档、代码注释、外部规范

#### 项目记忆（Memory）

5. **存储对象**（全部项目本地，不跨项目共享）

   | 对象 | 位置 | 说明 |
   |---|---|---|
   | pitfalls | `.aicodingagentteam/memory/dev-errors.jsonl` | 失败事件库 |
   | lessons | `.aicodingagentteam/memory/learned-skills/` | 经验证的规则 |
   | facts | `.aicodingagentteam/memory/facts.jsonl` | 项目环境事实 |
   | recipes | `.aicodingagentteam/memory/recipes.jsonl` | 历史交付方案 |

6. **记忆机制**
   - pitfalls：记录运行失败事件，**重复发生才生成待验证规则**
   - lessons：只有修复后**验证器确认修复有效**，规则才正式生效
   - facts：提取项目环境事实（语言、框架、数据库等），带来源证据；过期标记 tombstone
   - recipes：历史交付方案，**严格匹配栈**，仅作为建议

7. **开关控制**
   ```bash
   aicodingagentteam memory capture off --scope project --store facts
   aicodingagentteam memory recall off --scope project --store recipes
   ```

## 验收标准

- [x] `Engine.Retrieve(ctx, query, topK)` 返回 `[]Chunk` 无 panic
- [x] BM25 检索在无向量依赖时正常工作（降级不报错）
- [x] 云端 embedding 未设双环境变量时不触发上传
- [x] `AICODINGAGENTTEAM_ALLOW_CLOUD_EMBED` 未设时纯 BM25 降级
- [x] `Store.Capture(ctx, fact)` 存储后可 Recall
- [x] `Store.Recall(ctx, stack)` 按 stack 匹配返回 recipes
- [x] 记忆不跨项目共享（不同工作目录的 facts 隔离）
- [x] facts 过期后标记 tombstone
- [x] lessons 规则在验证器确认后才生效（未验证的不影响行为）
- [x] `memory capture off` / `memory recall off` 开关生效
- [x] repo-map 符号索引能扫描 `.go` 文件的函数/类型声明

## 非目标

- 不实现本地向量模型集成（MVP 阶段纯 BM25，向量后续迭代加）。
- 不实现云端 embedding（需双环境变量，MVP 默认关闭）。
- 不实现 HyDE 查询扩展（MVP 阶段直接用原始 query 检索 BM25）。
- 不实现记忆的自动跨项目共享（设计上禁止，ADR-0005）。
- 不保证记忆消除所有模型错误（辅助减少重复错误，非保证）。

## 依赖与约束

- 前置依赖：`internal/config`（cloud_embed 配置）
- 技术约束：BM25 纯 Go 实现，零外部依赖；本地向量模型 224MB 磁盘
- 风险与缓解：
  - 向量模型加载失败 → 降级 BM25
  - 记忆噪音 → lessons 须验证器确认才生效
  - 隐私 → 默认本地，云端需双环境变量显式开启

## 关联

- 实现计划：`docs/plan/mvp-preparation.md`
- 相关 ADR：ADR-0005（不持有密钥/本地优先）
- 源码：`internal/knowledge/knowledge.go`、`internal/memory/memory.go`