# 功能：质量门禁（Quality-Gate）

> 文件名：`docs/spec/quality-gate.md`
> 遵循模板：`aicoding_docs/docs/templates/SPEC.md`
> 关联 ADR：ADR-0006、ADR-0009

---

## 背景与目标

质量门禁是交付前的最后一道确定性校验，不依赖模型主观判断，机器硬执行。已在 `internal/qualitygate/engine.go` 中搭建骨架，当前返回硬编码 score=95。ADR-0009 验证 go build/test/vet 均可程序化执行（build 8s + vet 2.5s + test 9.5s）。

### 为谁解决什么问题

- Coordinator 需要在交付前自动校验交付物质量，不信任模型自评；
- 用户需要看到可审计的质量报告（scorecard），而非模糊的"已完成"。

## 用户故事

- 作为 Coordinator，我希望在交付前执行确定性校验，以便阻塞不合格交付。
- 作为用户，我希望看到质量得分和具体阻塞项，以便知道哪些问题必须修。
- 作为 QA Agent，我希望我的 Verdict 能接入质量门禁的校验结果。

## 功能描述

### 做什么

1. **校验清单（10 项）**
   | # | 校验项 | 类型 | 工具 | 阻塞级别 |
   |---|---|---|---|---|
   | 1 | PRD 完整性 | 文档 | 文件存在 + 关键词 | advisory |
   | 2 | 架构/API/数据模型 | 文档 | OpenAPI 文件存在 | advisory |
   | 3 | 前后端契约交叉校验 | 契约 | `pkg/contracts` CrossCheck | blocking |
   | 4 | UI 坏味道（emoji/硬编码颜色） | 治理 | 正则扫描 | advisory |
   | 5 | 编译 | 构建 | `go build ./...` | blocking |
   | 6 | 单元测试 | 测试 | `go test ./...` | blocking |
   | 7 | Lint | 静态 | `golangci-lint` + `eslint` | blocking |
   | 8 | 密钥泄露扫描 | 安全 | 正则 + git-secrets | blocking |
   | 9 | Runtime 探针 | 运行时 | 启动应用 + 访问路由 | advisory |
   | 10 | 审计日志完整性 | 审计 | `.aicodingagentteam/audit/` 存在 | advisory |

2. **评分规则**
   - 每项校验权重相等（10 分/项），总分 100
   - blocking 项失败 → 该项 0 分 + 总分直接标记不通过
   - advisory 项失败 → 该项 0 分，不影响通过判定
   - 阈值：`CONSTRAINTS.md` 默认 ≥ 90

3. **执行方式**
   - 每项校验是一个 `Check` 结构：Name、Command（exec.Command 调用）、Timeout、阻塞级别
   - 并行执行无依赖的校验项（如文档检查 + 安全扫描 + 构建）
   - 串行执行有依赖的项（先构建 → 再测试 → 再 lint）
   - 超时项标记失败，不阻塞其他项

4. **阻塞处理**
   - blocking 项失败 → 生成修复方案（含失败详情 + evidence 文件路径）
   - 指纹快照防止无限循环修复（记录已尝试的修复，重复则 park）
   - `Result{Score, Passed, Blocking[], Advisory[]}`

5. **输出产物**
   - `output/*-quality-gate.json`（结构化结果）
   - `output/*-quality-gate.md`（人类可读报告）
   - `.aicodingagentteam/audit/verify.jsonl`（审计日志）

6. **fail-open 策略**（ADR-0006）
   - 质量门禁引擎自身异常（非校验项失败）→ 记录 WARN + 返回默认通过
   - 校验项自身失败不是 fail-open（是真实 blocking）

## 验收标准

- [x] `Engine.Verify(ctx, artifacts)` 返回 `Result{Score, Passed, Blocking, Advisory}`
- [x] Score ≥ 阈值时 `Passed=true`，低于阈值时 `Passed=false`
- [x] go build 失败时 blocking 列表包含 "build" 项
- [x] go test 失败时 blocking 列表包含 "test" 项
- [x] 前后端契约不匹配时 blocking 列表包含 "api-contract" 项
- [x] advisory 项失败不影响 Passed 判定
- [x] 校验超时标记失败，不 panic
- [x] 引擎自身异常时返回默认通过（fail-open），不阻断开发
- [x] 结果写入 `verify.jsonl` 审计日志
- [x] `aicodingagentteam verify --runtime` 命令执行全部门禁 + Runtime 探针
- [x] `aicodingagentteam report` 命令输出 scorecard

## 非目标

- 不依赖模型自评校验（全部机器硬执行，ADR-0006 治理层 fail-open，质量门禁层是硬校验）。
- 不实现 Runtime 探针的完整逻辑（MVP 阶段 stub，后续迭代实现应用启动 + 路由访问）。
- 不实现合规映射（SOC-2/ISO27001/EU-AI-Act），后续迭代加。
- 不实现指纹快照防循环（MVP 单次执行，后续迭代加）。

## 依赖与约束

- 前置依赖：`pkg/contracts`（契约交叉校验）、`internal/governance`（UI 坏味道规则）、`internal/audit`（审计日志）
- 技术约束：go 工具链在 PATH 中可用（ADR-0009 验证）；golangci-lint 需安装
- 风险与缓解：
  - 校验耗时长 → 并行执行 + 超时控制（单项默认 120s）
  - 容器内无 go 工具链 → CI 镜像预装

## 关联

- 实现计划：`docs/plan/mvp-preparation.md`
- 相关 ADR：ADR-0006（fail-open）、ADR-0009（执行可行性）
- 质量约束：`docs/CONSTRAINTS.md`（阈值表）
- 源码：`internal/qualitygate/engine.go`、`pkg/contracts/contracts.go`、`internal/governance/governance.go`