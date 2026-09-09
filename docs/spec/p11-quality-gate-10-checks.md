# SPEC: P11 质量门禁 10 项校验补齐

> 关联 ADR：ADR-0006（fail-open 治理）、ADR-0009（质量门禁执行）、ADR-0023（插件与 init）
> 关联计划：`docs/plan/p11-quality-gate-10-checks.md`
> 关联 spec：`docs/spec/quality-gate.md`

## 1. 背景

`docs/spec/quality-gate.md` 定义质量门禁 10 项校验，当前 `internal/qualitygate/engine.go` 仅实现 4 项（build/vet/test/lint）+ runtime probe。缺 6 项，其中大部分功能已在 governance / contracts 包中实现，只是未接入 qualitygate 统一执行。

### 当前 4 项已实现

| # | 校验项 | 类型 | 实现方式 | 状态 |
|---|---|---|---|---|
| 5 | 编译 | blocking | `go build ./...` | ✅ |
| 6 | 单元测试 | blocking | `go test ./... -count=1` | ✅ |
| 7 | Lint | advisory | `golangci-lint run ./...` | ✅ |
| 10b | Runtime 探针 | advisory | backend `--version` | ✅ |

> 注：vet 算第 5 项的子项。

### 缺失 6 项

| # | 校验项 | 类型 | 现有实现 | 缺口 |
|---|---|---|---|---|
| 1 | PRD 完整性 | advisory | 无 | 需新建：检查 `output/*-prd.md` 存在 + 含验收标准关键词 |
| 2 | 架构/API 设计 | advisory | 无 | 需新建：检查 `output/*-architecture.md` + `output/openapi.*` 存在 |
| 3 | 前后端契约交叉校验 | blocking | `pkg/contracts.CrossCheck` | 需接入 qualitygate |
| 4 | UI 坏味道 | advisory | `governance.checkEmoji` / `checkHardcodedColor` | 需接入 qualitygate |
| 8 | 密钥泄露扫描 | blocking | `governance.checkSecretLeak` | 需接入 qualitygate |
| 10a | 审计日志完整性 | advisory | 无 | 需新建：检查 `.aicodingagentteam/audit/` 目录存在 + 非空 |

## 2. 设计

### 核心问题

当前 `Check` 结构仅支持 CLI 命令（`Command []string`）。缺失的 6 项需要不同类型的校验逻辑（文件检查、governance 扫描、contract 交叉校验）。

### 方案：CheckFunc 扩展

新增 `CheckFunc` 类型，允许注册自定义校验函数，与现有 `Command` 并行：

```go
type CheckFunc func(ctx context.Context, artifacts []string) CheckDetail
```

`Check` 结构新增 `Func CheckFunc` 字段。Verify 时优先执行 `Func`（如有），否则执行 `Command`。

### 10 项校验清单

| # | 名称 | Severity | 实现方式 |
|---|---|---|---|
| 1 | prd-completeness | advisory | 文件检查：`output/*-prd.md` 存在 + 含 "验收" 关键词 |
| 2 | architecture-design | advisory | 文件检查：`output/*-architecture.md` + `output/openapi.*` 存在 |
| 3 | api-contract-crosscheck | blocking | `contracts.CrossCheck` 对 output/ 下 .go 文件扫描 fetch 调用 |
| 4 | ui-smells | advisory | `governance.Engine.Check` 扫描 output/ 下 .tsx/.ts 文件 |
| 5 | build | blocking | `go build ./...` |
| 6 | test | blocking | `go test ./... -count=1` |
| 7 | lint | advisory | `golangci-lint run ./...` |
| 8 | secret-leak | blocking | `governance.checkSecretLeak` 扫描全部产物文件 |
| 9 | audit-log | advisory | 文件检查：`.aicodingagentteam/audit/` 存在 + 非空 |
| 10 | runtime-probe | advisory | backend `--version` |

## 3. 验收标准

- [x] `defaultChecks()` 返回 10 项校验（含 vet 子项）
- [x] PRD/architecture/audit-log 校验用文件检查实现，不依赖外部二进制
- [x] secret-leak 校验复用 governance 的 `checkSecretLeak` 规则
- [x] api-contract-crosscheck 校验复用 `pkg/contracts.CrossCheck`
- [x] ui-smells 校验复用 governance 引擎
- [x] Verify() 执行全部 10 项，评分按 100/10 项计算
- [x] 缺少产物目录时，advisory 项返回 skipped 不 panic
- [x] secret-leak 命中时返回 blocking，detail 含匹配行
- [x] 覆盖率 86.3%（qualitygate 包）
- [x] `go build ./...` + `go vet ./...` + `go test ./internal/qualitygate/... -cover` 通过

## 4. 非功能约束

- fail-open：panic 不阻断交付
- 不引入新外部依赖
- CheckFunc 执行超时控制（继承 Check.Timeout）
- advisory 项缺失文件时返回 skipped，不报 fail

## 5. 完成定义

- [x] 全部 10 项校验实现并通过测试
- [x] ADR-0024 记录决策
- [x] CHANGELOG 更新
- [x] quality-gate spec 更新标注已实现项
