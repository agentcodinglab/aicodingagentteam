# ADR-0024：质量门禁 10 项校验补齐

> 日期：2026-09-08
> 状态：Accepted
> 关联：ADR-0006（fail-open 治理）、ADR-0009（质量门禁执行）、ADR-0023（插件与 init）
> 关联 spec：`docs/spec/quality-gate.md`、`docs/spec/p11-quality-gate-10-checks.md`

## 背景

`internal/qualitygate/engine.go` 原仅实现 4 项校验（build/vet/test/lint）+ runtime probe。缺 6 项，其中大部分功能已在 governance / contracts 包中实现，但未接入 qualitygate 统一执行。

### 缺失的 6 项

| # | 校验项 | 类型 | 现有实现 | 缺口 |
|---|---|---|---|---|
| 1 | PRD 完整性 | advisory | 无 | 需新建：文件检查 |
| 2 | 架构/API 设计 | advisory | 无 | 需新建：文件检查 |
| 3 | 前后端契约交叉校验 | blocking | `pkg/contracts.CrossCheck` | 需接入 qualitygate |
| 4 | UI 坏味道 | advisory | `governance.checkEmoji` / `checkHardcodedColor` | 需接入 qualitygate |
| 8 | 密钥泄露扫描 | blocking | `governance.checkSecretLeak` | 需接入 qualitygate |
| 10a | 审计日志完整性 | advisory | 无 | 需新建：文件检查 |

## 决策

### 方案：CheckFunc 扩展

新增 `CheckFunc` 类型，允许注册自定义校验函数，与现有 `Command` 并行：

```go
type CheckFunc func(ctx context.Context, artifacts []string) CheckDetail
```

`Check` 结构新增 `Func CheckFunc` 字段。`runCheck()` 优先执行 `Func`（如有），否则执行 `Command`。

### 理由

1. **最小侵入**：不改变现有 `Command` 执行逻辑，向后兼容
2. **可扩展**：未来新增校验只需注册 CheckFunc，无需修改引擎
3. **复用现有**：governance / contracts 包的逻辑可直接调用
4. **可测试**：CheckFunc 是纯函数，易于单元测试

## 后果

### 正面

- 质量门禁从 4 项扩展到 10 项，覆盖文档/架构/安全/审计
- `CheckFunc` 机制为未来扩展提供标准接口
- advisory 项缺失文件时返回 skipped，不 panic

### 负面

- #3 api-contract-crosscheck 和 #8 secret-leak 已接入 governance/contracts 真实逻辑。#3 的 OpenAPI 解析已从行扫描升级为 `encoding/json` 完整解析。
- 评分计算从 `100/4` 变为 `100/10`，单项权重降低

## 实现清单

- [x] `CheckFunc` 类型定义
- [x] `Check.Func` 字段
- [x] `runCheck()` 支持 CheckFunc
- [x] `prdCompleteness()` — 文件检查
- [x] `architectureDesign()` — 文件检查
- [x] `apiContractCrossCheck()` — 接入 `pkg/contracts.CrossCheck`，OpenAPI 解析升级为 `encoding/json`
- [x] `uiSmells()` — 文件检查
- [x] `secretLeak()` — 接入 `governance.checkSecretLeak`，扫描 output/ 全部文件
- [x] `auditLog()` — 文件检查
- [x] `defaultChecks()` 返回 10 项
- [x] 单元测试更新（`TestNew_CreatesEngineWithDefaultChecks`）
- [x] `go build` + `go vet` + `go test` 通过
- [x] qualitygate 覆盖率 52.9% → 86.3%（≥80% 红线达标）
- [x] governance 覆盖率维持 96.4%+
- [x] `golangci-lint` 0 issues（P11 变更包无新增告警）
- [x] `runCheck()` 对 CheckFunc 应用 Timeout 超时控制
- [x] `uiSmells()` 同时扫描 .ts 和 .tsx 文件
