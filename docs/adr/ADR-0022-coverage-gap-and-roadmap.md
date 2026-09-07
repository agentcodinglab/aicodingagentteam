# ADR-0022：覆盖率缺口补齐与下一步路线图

> 状态：Accepted
> 日期：2026-09-06
> 关联：ADR-0018（方向 C e2e）、ADR-0019（P4 流式）、ADR-0020（P6 ACP）、ADR-0021（P5 OpenCode ACP）
> 计划：`docs/plan/next-steps-coverage-and-release.md`

## 背景

P4–P6 + 方向 C/D 全部落地发版后（v0.8.1），项目进入稳定期。当前 `go test ./... -cover` 实测显示：

- `cmd/aicodingagentteam` 和 `cmd/godocgen` 覆盖率 **0.0%**，是最大盲区。
- `internal/godocgen`（73.5%）、`internal/planner`（82.4%）、`internal/host/opencode`（84.1%）、`internal/agent`（84.6%）四个包未达 ≥85% 红线。
- 其余 20+ 包全部 ≥85%，核心编排逻辑（coordinator/scheduler/governance/host/codex/knowledge）均 ≥87%。

用户明确指示：**不搞 Claude/DSH 真实 exec**。这两个包当前为 stub mock 实现，覆盖率 100%，维持现状。

## 决策

### 决策 1：补齐覆盖率缺口（P7）

优先级：**最高**。覆盖率红线是 CONSTRAINTS.md 的硬约束，`cmd/*` 零覆盖是质量风险最大盲区。

范围：
- `cmd/aicodingagentteam`：写 smoke 测试（`version`/`init`/`knowledge demo`/`verify` 子命令），目标 ≥60%。
- `cmd/godocgen`：写 smoke 测试，目标 ≥60%。
- `internal/godocgen`→85%、`internal/planner`→85%、`internal/host/opencode`→85%、`internal/agent`→85%。
- CI 加 CLI smoke job（soft-fail），先 warn 不 block，稳定后可升级。

### 决策 2：文档站测试入口 + SEO 深化（P8）

优先级：**中**。文档站已有 sitemap/robots/a11y，但 `package.json` 缺 `test` 脚本入口，tests/ 目录的 Playwright a11y 未纳入 npm scripts。

范围：
- 加 `test` 脚本统一跑 perf-budget + a11y。
- OG image 动态生成（每语言首页一图）。
- JSON-LD 结构化数据（SoftwareApplication + BreadcrumbList）。
- Lighthouse 配置更新。

### 决策 3：发布 v0.9.0（P9）

优先级：**最低**。P7+P8 完成后发版，整理 CHANGELOG + 完善 ISSUE_TEMPLATE + tag 触发 goreleaser。

### 不做的

- **不搞 Claude/DSH 真实 exec**：用户明确拒绝。维持 stub + 100% 覆盖现状。
- 不动 ACP/MCP 协议层（P5/P6 已稳定发版）。
- 不重构编排引擎核心（Director/Scheduler 全部达标）。

## 后果

- 覆盖率红线达标后，`cmd/*` 从 0% 提升到 ≥60%，消除最大质量盲区。
- 文档站 SEO 深化后 Lighthouse SEO 评分提升，多语言首页有 OG image 社交分享。
- v0.9.0 发版后项目进入社区可贡献状态（good-first-issue + 完善 CONTRIBUTING）。
