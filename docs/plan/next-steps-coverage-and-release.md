# 下一步计划 — 覆盖率补齐、CLI smoke、文档站深化、发布准备

> 决策依据：`docs/adr/ADR-0022-coverage-gap-and-roadmap.md`
> 基线版本：v0.8.1（已发版）
> 前置依赖：P4–P6 全部落地 + 方向 C/D 完成（已满足）
> 用户约束：**不搞 Claude/DSH 真实 exec**，仅维护现有 stub + 零覆盖状态

## 0. 盘点快照（2026-09-06 实测）

### 覆盖率缺口

| 包 | 当前覆盖率 | 目标 | 状态 |
|---|---|---|---|
| `cmd/aicodingagentteam` | 0.0% | ≥60% | 最大盲区 |
| `cmd/godocgen` | 0.0% | ≥60% | 零覆盖 |
| `internal/godocgen` | 73.5% | ≥85% | 偏低 |
| `internal/planner` | 82.4% | ≥85% | 接近阈值 |
| `internal/host/opencode` | 84.1% | ≥85% | 差一点 |
| `internal/agent` | 84.6% | ≥85% | 差一点 |
| `pkg/api` | 85.5% | ≥85% | 达标 |
| `pkg/api/gen` | 0.0% | 豁免 | 生成代码 |
| 其余包 | ≥87% | ≥85% | 全部达标 |

### CI 完整度

- `ci.yml`：build/vet/test/codecov/golangci-lint@v9/govulncheck/gitleaks/semgrep/trivy — 完整
- `governance.yml`：perf-budget(hard)/lighthouse(soft)/a11y(soft)/rag-demo(soft)/host-e2e(soft)/godocgen(soft)/summary — 完整
- `docs-site.yml`：GitHub Pages 静态部署 — 正常
- `release.yml`：tag-triggered goreleaser — 正常

### 文档站状态

- Next.js 14 + next-intl 9 语言，`website/` 下
- 有 `scripts/gen-sitemap.mjs`、`scripts/perf-budget.mjs`、`.lighthouserc.cjs`
- `package.json` 无 `test` 脚本入口（tests/ 目录存在但 Playwright a11y 未纳入 scripts）
- ISSUE_TEMPLATE 有 bug/feature 两模板，CONTRIBUTING.md 存在

## 1. 范围

### 做

1. **P7 — CLI smoke 测试 + 覆盖率补齐**：为 `cmd/aicodingagentteam` 和 `cmd/godocgen` 写 smoke 测试；补 `internal/godocgen`/`planner`/`opencode`/`agent` 到 ≥85%。
2. **P8 — 文档站测试入口 + SEO 深化**：给 `website/package.json` 加 `test` 脚本跑 a11y + perf-budget；加 OG image 生成 + structured data。
3. **P9 — 发布 v0.9.0 准备**：整理 CHANGELOG `[Unreleased]` → `[0.9.0]`；完善 ISSUE_TEMPLATE good-first-issue 标签体系；tag 触发 goreleaser。

### 不做

- **不搞 Claude/DSH 真实 exec**：用户明确拒绝。现有 `claude`/`dsh` 包 100% 覆盖率（stub mock 实现），维持现状即可。
- 不改 ACP/MCP 协议层（P5/P6 已稳定）。
- 不动编排引擎核心逻辑（Director/Scheduler 已达标）。

## 2. 实施步骤

### P7 — CLI smoke + 覆盖率补齐

| # | 任务 | 产出物 | 估时 | 关键路径 |
|---|---|---|---|---|
| P7.1 | `cmd/aicodingagentteam` smoke：`version`/`init`/`knowledge demo`/`verify` 子命令跑通 | `cmd/aicodingagentteam/main_test.go` | 40 min | ✓ |
| P7.2 | `cmd/godocgen` smoke：生成文档快照比对 | `cmd/godocgen/main_test.go` | 25 min | ✓ |
| P7.3 | `internal/godocgen` 补测 73.5%→85%：覆盖未测分支 | `internal/godocgen/*_test.go` | 40 min | |
| P7.4 | `internal/planner` 补测 82.4%→85%：补边界用例 | `internal/planner/*_test.go` | 20 min | |
| P7.5 | `internal/host/opencode` 补测 84.1%→85%：补 ACP error path | `internal/host/opencode/*_test.go` | 20 min | |
| P7.6 | `internal/agent` 补测 84.6%→85%：补 edge case | `internal/agent/*_test.go` | 15 min | |
| P7.7 | CI 加 CLI smoke job（soft-fail）到 `governance.yml` | `.github/workflows/governance.yml` | 15 min | |
| P7.8 | ADR-0022 + 计划 + CHANGELOG | `docs/adr/` + `CHANGELOG.md` | 20 min | 末 |

总估时 ~3 小时。

### P8 — 文档站测试入口 + SEO 深化

| # | 任务 | 产出物 | 估时 | 关键路径 |
|---|---|---|---|---|
| P8.1 | `website/package.json` 加 `test` 脚本：跑 perf-budget + a11y | `website/package.json` | 15 min | ✓ |
| P8.2 | OG image 生成脚本（每语言一页动态图） | `website/scripts/gen-og-images.mjs` | 40 min | |
| P8.3 | 结构化数据 JSON-LD（SoftwareApplication + BreadcrumbList） | `website/app/structured-data.tsx` | 30 min | |
| P8.4 | Lighthouse 配置更新：加 OG image audit | `website/.lighthouserc.cjs` | 10 min | |
| P8.5 | CHANGELOG 更新 | `CHANGELOG.md` | 5 min | 末 |

总估时 ~1.5 小时。

### P9 — 发布 v0.9.0 准备

| # | 任务 | 产出物 | 估时 | 关键路径 |
|---|---|---|---|---|
| P9.1 | 整理 CHANGELOG `[Unreleased]` → `[0.9.0]` | `CHANGELOG.md` | 15 min | |
| P9.2 | ISSUE_TEMPLATE 加 `good-first-issue` 配置 + 细化 bug report | `.github/ISSUE_TEMPLATE/` | 20 min | |
| P9.3 | CONTRIBUTING.md 补充：本地开发快速开始 + 测试命令 | `CONTRIBUTING.md` | 15 min | |
| P9.4 | tag `v0.9.0` 触发 goreleaser release | git tag | 5 min | 末 |

总估时 ~1 小时。

## 3. 验收标准

- [x] `go test ./... -cover` 全包 ≥85%（`cmd/*` ≥60%，`pkg/api/gen` 豁免）
- [x] `website/package.json` 有 `test` 脚本，`npm test` 退出码 0
- [x] OG image + JSON-LD 在 Lighthouse SEO 评分 ≥90
- [x] CHANGELOG `[0.9.0]` + `[0.9.1]` section 完整
- [x] `v0.9.0` tag 推送后 release.yml 成功触发（dry-run 验证通过）
- [x] 所有 lint 通过（golangci-lint + eslint）

## 4. 并行与依赖

```
P7.1 ─┐
P7.2 ─┤─ P7.3-P7.6（可并行）─ P7.7 ─ P7.8
P7.3-P7.6 可在 P7.1/P7.2 完成后并行启动

P8.1 ─ P8.2 ─ P8.3 ─ P8.4 ─ P8.5（串行，有依赖）

P9 全部在 P7+P8 完成后执行
```

## 5. 风险

| 风险 | 概率 | 缓解 |
|---|---|---|
| CLI smoke 在 CI 环境行为不一致 | 中 | soft-fail 降级，先 warn 不 block |
| OG image 生成需额外依赖 | 低 | 用纯 Canvas API 无外部依赖 |
| v0.9.0 tag 时 goreleaser 配置过时 | 低 | 发版前先 dry-run `goreleaser --snapshot` |
