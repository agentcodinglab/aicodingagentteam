# 质量约束（质量红线）— AiCodingAgentTeam

> 本文件定义项目质量阈值，是不可静默降级的红线。
> 范本：`E:\javaproject\my\2026\aicoding_docs\docs\CONSTRAINTS.md`
> 具体项目可在此覆盖 `AGENTS.md` 的默认值；未覆盖项沿用默认。

---

## 1. 阈值表

| 维度 | 默认阈值 | 可覆盖 | 说明 |
|---|---|---|---|
| Go 单元测试覆盖率 | ≥ 80% | 是 | 核心编排逻辑（coordinator/scheduler/router）≥ 90% |
| Go 集成测试覆盖率 | ≥ 60% | 是 | A2A 协议、宿主驱动边界 |
| TS TUI 单元覆盖率 | ≥ 70% | 是 | Slash 命令路由、状态管理 |
| 端到端关键路径 | 100% 通过 | 否 | A2A 协作闭环、完整绿场流水线、质量门禁阻塞 |
| Lint | 0 新增警告 | 否 | golangci-lint + eslint，不允许新增 disable 无原因 |
| 构建时长 | ≤ 5 分钟 | 是 | CI 全流程含测试 |
| A2A 消息处理 p95 | ≤ 500ms | 是 | Coordinator→Agent 委派 + 结果回传 |
| 容器启动到就绪 | ≤ 30s | 是 | 各角色 Agent 容器 |
| 质量门禁得分 | ≥ 90 | 否 | AiCodingAgentTeam Quality-Gate |
| 依赖漏洞 | 0 高危 | 否 | govulncheck + npm audit |
| 安全静态扫描 | 0 高危 | 否 | SAST（Semgrep / CodeQL） |
| 密钥泄露扫描 | 0 命中 | 否 | 宿主 API Key 不许出现在仓库 |

---

## 2. 红线规则

1. **不得为通过检查而调低阈值**。确需调整须经用户确认并记 ADR。
2. **不得静默跳过或删除测试**。确需 skip 须标原因与跟进条件。
3. **不得新增 `//nolint` / `eslint-disable` / `@ts-ignore` 抑制无原因**；须带原因且限定范围。
4. **不得为提覆盖率而改断言期望值或删用例**（除非原断言确实错误，且记原因）。
5. **不得提交「红」状态**（构建失败/测试失败）到主干。
6. **不得在 Coordinator 或 Agent 代码中硬编码任何 API Key / 凭证**。

---

## 3. 覆盖流程

- 想调整某阈值 → 先在 `docs/adr/` 记录决策（背景/决策/后果）。
- 用户确认后，在本文件修改对应行，并标注变更日期与 ADR 编号。
- 不经此流程的降级视为违规。

---

## 4. 检查工具基线

| 类型 | 工具示例 |
|---|---|
| Go 覆盖率 | `go test -cover` / `go test -coverprofile` |
| TS 覆盖率 | Istanbul (nyc) / vitest coverage |
| Go Lint | golangci-lint + go vet |
| TS Lint | eslint + tsc --noEmit |
| Go 漏洞 | govulncheck |
| TS 漏洞 | npm audit |
| SAST | Semgrep / CodeQL |
| 容器安全 | trivy image 扫描 |
| 性能 | k6 / wrk（A2A 消息压测） |

- CI 中以上工具全量执行，阈值不达标即失败。

---

## 5. 项目专属校验项

| 校验 | 工具 | 说明 |
|---|---|---|
| A2A 协议契约 | JSON Schema 校验 | Agent Card / Task / Verdict 结构合规 |
| 工件完整性 | 文件存在性校验 | 各阶段产出文件齐全 |
| 单写者约束 | 并发审计 | 同一时刻无两个写角色修改源码 |
| fail-open 治理 | 故障注入 | 治理引擎异常不阻断开发 |