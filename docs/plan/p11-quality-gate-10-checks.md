# 计划：P11 质量门禁 10 项校验补齐

> 关联 spec：`docs/spec/p11-quality-gate-10-checks.md`
> 关联 ADR：ADR-0024
> 状态：✅ 已完成

## 1. 目标

将 `internal/qualitygate/engine.go` 从 4 项校验扩展到 10 项，补齐 6 项缺失校验，并将 3 项占位实现升级为真实逻辑。

## 2. 步骤

### 阶段一：骨架扩展（CheckFunc 机制）
- [x] 新增 `CheckFunc` 类型定义
- [x] `Check` 结构新增 `Func CheckFunc` 字段
- [x] `runCheck()` 优先执行 `Func`（如有），否则执行 `Command`
- [x] `defaultChecks()` 从 4 项扩展到 10 项

### 阶段二：6 项新校验函数
- [x] `prdCompleteness` — 文件检查 `output/*-prd.md` + 验收关键词
- [x] `architectureDesign` — 文件检查 `output/*-architecture.md` + `output/openapi.*`
- [x] `apiContractCrossCheck` — 扫描 OpenAPI spec + 前端 fetch/axios 调用，用 `contracts.CrossCheck` 交叉校验
- [x] `uiSmells` — 用 `governance.Engine.Check` 扫描 `output/**/*.tsx` 和 `*.ts`
- [x] `secretLeak` — 用 `governance.Engine.Check` + `checkSecretLeak` 扫描 `output/` 全部文件
- [x] `auditLog` — 文件检查 `.aicodingagentteam/audit/*.jsonl`

### 阶段三：占位升级为真实实现
- [x] `apiContractCrossCheck` — 接入 `pkg/contracts.CrossCheck`
- [x] `uiSmells` — 接入 `governance.Engine.Check`
- [x] `secretLeak` — 接入 `governance.checkSecretLeak`

### 阶段四：OpenAPI 解析增强
- [x] governance `loadOpenAPIPaths` 改为 `encoding/json` 解析 + fallback 行扫描
- [x] qualitygate `apiContractCrossCheck` 改为 `encoding/json` 解析
- [x] `Continue` 方法 fallback 到 `GetPlan`（pkg/api/server.go）

### 阶段五：测试覆盖
- [x] `internal/qualitygate/checks_test.go` — 20 个新测试
- [x] `internal/governance/apicontract_test.go` — 17 个新测试
- [x] `TestNew_CreatesEngineWithDefaultChecks` 更新为期望 10 项
- [x] qualitygate 覆盖率 52.9% → 86.3%
- [x] governance 覆盖率维持 96.4%+

## 3. 验收标准

- [x] `defaultChecks()` 返回 10 项校验
- [x] 全部 10 项校验实现真实逻辑（无占位）
- [x] `go build` + `go vet` + `go test` 通过
- [x] `golangci-lint` 0 issues
- [x] qualitygate 覆盖率 ≥ 80%
- [x] ADR-0024 记录决策
- [x] CHANGELOG 更新
- [x] quality-gate spec 更新

## 4. 产出物

| 文件 | 描述 |
|------|------|
| `internal/qualitygate/engine.go` | 核心实现 |
| `internal/qualitygate/checks_test.go` | 6 个 CheckFunc 测试 + runCheck 分支测试 |
| `internal/qualitygate/engine_test.go` | 更新 defaultChecks 测试 |
| `internal/governance/rules.go` | checkAPIContract 真实实现 + loadOpenAPIPaths JSON 解析 |
| `internal/governance/apicontract_test.go` | 17 个 governance 契约测试 |
| `pkg/api/server.go` | Continue 方法 fallback 实现 |
| `docs/spec/p11-quality-gate-10-checks.md` | 功能规格 |
| `docs/spec/quality-gate.md` | 更新校验表状态 |
| `docs/adr/ADR-0024-quality-gate-10-checks.md` | 架构决策记录 |
| `CHANGELOG.md` | 变更日志 |
