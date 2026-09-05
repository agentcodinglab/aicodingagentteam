# 实现计划：RAG 知识库与项目记忆编排集成

> 文件名：`docs/plan/rag-knowledge-memory-integration.md`
> 遵循模板：`aicoding_docs/docs/templates/PLAN.md`
> 关联规格：`docs/spec/rag-knowledge-memory-integration.md`

---

## 任务分解

### Step 1: Director 结构体扩展 + Option setter

- 文件：`internal/coordinator/director.go`
- 改动：
  - import `internal/knowledge`、`internal/memory`
  - Director 结构体新增 `knowledge *knowledge.Engine`、`memory *memory.Store` 字段
  - 新增 `WithKnowledge(e *knowledge.Engine) Option`、`WithMemory(s *memory.Store) Option`（函数式 setter，返回 `func(*Director)`）
  - 新增 `NewWithKnowledge`/`NewWithMemory` 组合 setter 或在现有构造器链式调用
- 验证：编译通过；现有构造器 `New`/`NewWithBus` 行为不变（knowledge/memory 为 nil）

### Step 2: Handle 流接入 RAG 检索 + 记忆召回/捕获

- 文件：`internal/coordinator/director.go` — `Handle` 方法
- 改动：
  - Route 之后：`d.enhanceWithKnowledge(ctx, req.Message)` — 检索 top-5 chunk，append 到 intent.Scope（或新增 intent context 字段）
  - Plan 阶段：`d.recallMemory(ctx, intent)` — RecallFacts + Recall(stack)，日志记录
  - Finalize 之后：`d.captureMemory(ctx, delivery)` — 成功捕获 fact，失败捕获 pitfall
  - 三个辅助方法均 fail-safe（err 只 log 不返回）
- 验证：单元测试覆盖 nil-skip、检索、召回、捕获四条路径

### Step 3: newDirector 接线（去除死代码）

- 文件：`cmd/aicodingagentteam/main.go` — `newDirector`
- 改动：
  - `keng := knowledge.New(false)`（替代 `_ = knowledge.New(false)`）
  - `mem := memory.New(".aicodingagentteam/memory")`（替代 `_ = memory.New(...)`）
  - 用 `WithKnowledge(keng)`/`WithMemory(mem)` 传入 Director
  - serve/run/quick 路径可选自动 `keng.IndexDirectory(".")` 索引当前目录
- 验证：`go vet ./...` 无 unused import；`go build ./...` 通过

### Step 4: CLI `knowledge` 子命令

- 文件：`cmd/aicodingagentteam/main.go`
- 改动：
  - `case "knowledge": cmdKnowledge(ctx, os.Args[2:])`
  - `cmdKnowledge`：flag subcommand `index`/`search`/`demo`
  - `demo`：端到端演示（临时目录→索引→检索→记忆写入→召回）
- 验证：`go run ./cmd/aicodingagentteam knowledge demo` 输出检索结果 + 记忆召回

### Step 5: Coordinator 单元测试

- 文件：`internal/coordinator/director_test.go`（新建或追加）
- 改动：
  - `TestHandle_KnowledgeRetrieve` — 注入 knowledge engine，验证 Handle 不 panic
  - `TestHandle_MemoryCaptureAndRecall` — 注入 memory store，验证 fact 捕获 + 召回
  - `TestHandle_NilKnowledgeMemory_NoPanic` — nil 时正常执行
  - `TestHandle_MemoryCapturePitfall` — 失败 delivery 捕获 pitfall
- 验证：`go test ./internal/coordinator/... -race` 通过

### Step 6: Codex driver 测试补齐（0% → 100%）

- 文件：`internal/host/codex/codex_test.go`（追加）
- 改动：参照 opencode_test.go 模式，补齐：
  - `TestCodex_WithOptions` — WithBinary/WithTimeout/WithMaxRetries
  - `TestCodex_StartDestroySession` — StartSession/DestroySession
  - `TestCodex_Capabilities` — Capabilities 字段
  - `TestCodex_ModelInfo` — ModelInfo 字段
  - `TestCodex_PauseResume` — Pause/Resume 返回 nil
  - `TestCodex_AuthStatusMissingBinary` — 缺失 binary 报 not ready
  - `TestCodex_SendTaskHappyPath` — mock binary 输出 stdout
  - `TestCodex_RunOnce` — runOnce 直接调用
  - `TestCodex_IsTransientBranches` — 表驱动
- 验证：`go test ./internal/host/codex/...` 覆盖率 ≥ 90%

### Step 7: CI golangci-lint-action v9 升级

- 文件：`.github/workflows/ci.yml`
- 改动：`golangci/golangci-lint-action@v8` → `@v9`
- 验证：CI 触发后无 Node 20 deprecation warning

### Step 8: ADR-0013 记录决策

- 文件：`docs/adr/ADR-0013-rag-memory-director-wiring.md`
- 内容：背景/决策/后果

## 关键路径

```
Step 1 (struct) → Step 2 (Handle) → Step 5 (test) → Step 3 (main wiring) → Step 4 (CLI)
Step 6 (codex tests) ∥ Step 7 (CI) ∥ Step 8 (ADR)
```

Step 6/7/8 与 1-5 独立，可并行。

## 验收清单

- [ ] `go build ./...` 通过
- [ ] `go test ./... -race` 通过
- [ ] golangci-lint 0 issue
- [ ] codex 覆盖率 ≥ 90%
- [ ] coordinator 覆盖不降
- [ ] `knowledge demo` 端到端跑通
- [ ] CI 三 job 全绿
- [ ] ADR-0013 已写