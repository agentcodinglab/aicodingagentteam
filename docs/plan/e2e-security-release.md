# 实现计划：端到端验证(J) + 安全加固(K) + 发布打磨(L)

> 文件名：`docs/plan/e2e-security-release.md`
> 方向：J（端到端验证）、K（CI 安全加固）、L（cmd 覆盖与发布）

---

## J — 端到端验证与集成测试

### J1: Go 端绿场 E2E 测试
- 文件：`internal/coordinator/e2e_test.go`
- 内容：init → run(quick) → verify → report 全链路，验证 Delivery 闭环
- 验证：`go test -run TestE2E ./internal/coordinator/...`

### J2: fail-open 治理故障注入测试
- 文件：`internal/governance/governance_test.go`（追加）
- 内容：注入 panic/nil rule，验证 fail-open 不阻断
- 验证：`go test -run TestFailOpen ./internal/governance/...`

### J3: 单写者并发审计测试
- 文件：`internal/scheduler/scheduler_test.go`（追加）
- 内容：并发 Execute 同一 plan，验证 write.lock 互斥
- 验证：`go test -run TestConcurrentWriteLock ./internal/scheduler/...`

### J4: proof-pack 交付打包
- 文件：`internal/qualitygate/proof.go`、`internal/qualitygate/proof_test.go`
- 内容：生成 proof-pack-*.zip（plan.json + verify.jsonl + scorecard.md）+ Scorecard
- 验证：`go test ./internal/qualitygate/...`

### J5: RAG spec 验收项闭合
- 文件：`docs/spec/rag-knowledge-memory-integration.md`
- 内容：勾选已实现的 13 个 checkbox
- 验证：文档 checkbox 全勾

## K — CI 安全加固

### K1: gitleaks 密钥泄露扫描
- 文件：`.github/workflows/ci.yml`
- 改动：go job 增 gitleaks step
- 验证：CI 通过

### K2: Semgrep SAST
- 文件：`.github/workflows/ci.yml`
- 改动：go job 增 semgrep scan step
- 验证：CI 通过

### K3: trivy 容器镜像扫描
- 文件：`.github/workflows/ci.yml`
- 改动：新增 security job，构建 Docker 镜像后 trivy scan
- 验证：CI 通过

### K4: A2A 契约 JSON Schema 校验
- 文件：`internal/a2a/schema_test.go`
- 内容：AgentCard / Task / Result JSON 序列化后 schema 校验
- 验证：`go test ./internal/a2a/...`

## L — cmd 覆盖与发布打磨

### L1: cmd/aicodingagentteam 测试
- 文件：`cmd/aicodingagentteam/main_test.go`
- 内容：exec.Command 测试 init/version/memory show/knowledge demo
- 目标：0% → 40%+
- 验证：`go test ./cmd/aicodingagentteam/...`

### L2: TUI 组件渲染测试
- 文件：`tui/src/__tests__/components.test.tsx`
- 依赖：ink-testing-library（已安装）
- 目标：PlanView/StatusBar/ResultPanel 渲染快照
- 验证：`npm run test:unit`

### L3: v0.2.0 tag 发布
- 文件：`CHANGELOG.md`（补 0.2.0 段）
- 动作：git tag v0.2.0 → push → 触发 release workflow
- 验证：GitHub Release 创建成功

## 关键路径
```
J1→J2→J3 (集成测试) ∥ J4 (proof) ∥ J5 (spec) → K1→K2→K3→K4 (安全) → L1→L2→L3 (发布)
```