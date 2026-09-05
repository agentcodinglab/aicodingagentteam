# Maintenance Standard

> 本规范确立 AiCodingAgentTeam 项目的长期治理责任、季度回归清单与运维节奏。
> 与 `AGENTS.md` §4（质量约束）配套。
> ADR-0016 派生：`docs/adr/ADR-0016-direction-a-governance.md`

## 1. 适用范围

本规范覆盖以下三类持续治理活动：

1. **门禁守恒**：保证 ADR-0016 引入的四类门禁（perf budget / a11y / Lighthouse / Go API 文档）长期有效。
2. **依赖与漏洞**：周期性刷新依赖、运行漏洞扫描、清理技术债。
3. **文档同步**：保证代码、ADR、文档站三者不脱节。

## 2. 季度回归清单

每季度（首周一）由维护负责人执行以下清单，结果写到 `docs/maintenance/YYYY-QN.md`。

### 2.1 性能与 a11y 门禁（ADR-0016 §1、§2、§3）

| 项 | 命令 | 通过条件 |
|---|---|---|
| Perf budget | `node website/scripts/perf-budget.mjs` | total gz <= 200 KB |
| Lighthouse CI | `npx lhci autorun --config=website/.lighthouserc.cjs` | perf >= 0.90 / 其余 >= 0.95 |
| axe-core a11y | `cd tests && npm run test:a11y` | serious/critical = 0 |

低于阈值时 **不静默调低**；先回归修复，必要时新增 ADR 解释阈值调整原因。

### 2.2 依赖与漏洞

| 项 | 命令 | 通过条件 |
|---|---|---|
| Go 漏洞扫描 | `govulncheck ./...` | 0 高危 |
| TUI/网站漏洞 | `cd tui && npm audit --omit=dev` / `cd website && npm audit --omit=dev` | 0 high / 0 critical |
| Go 模块更新 | `go get -u ./...` 然后 `go mod tidy` | 编译通过 + 测试通过 |
| Node 模块更新 | `cd website && npm outdated` / `cd tui && npm outdated` | 评估 major 升级影响 |

升级 Major 版本（next / next-intl / ink / React）必须先起 ADR 评估。

### 2.3 文档同步

| 项 | 命令 | 通过条件 |
|---|---|---|
| Go API 文档 | `go run ./cmd/godocgen --locale=en,zh,ja,ko,fr,de,ru,es,it --version=vX.Y.Z` | 9 语言产物齐 |
| 站内链接 | `node website/scripts/check-links.mjs`（如已存在） | 0 broken internal link |
| sitemap 更新 | `node website/scripts/gen-sitemap.mjs` | 91 URLs（root + 9 × 10） |

`website/content/docs/api/` 已 gitignore，godocgen 产物走 artifact 通道而非 commit。

### 2.4 治理与流程审计

| 项 | 通过条件 |
|---|---|
| 守护性 ADR（no-third-party-branding / RAG / quality-gate 等） | 内容与现状一致 |
| PR / Issue 模板 | 与当前 workflow 匹配 |
| Issue 积压 | > 30 天的 Issue 必须关闭或重定优先级 |

## 3. 治理责任分工

| 角色 | 责任 |
|---|---|
| Maintainer | 每季度执行清单，结果写到 `docs/maintenance/` |
| Reviewer | PR 阶段确保门禁通过，不接受"暂时关掉检查" |
| 临时贡献者 | 阅读本规范与 `AGENTS.md` 后再提交 PR |

## 4. 紧急维护（按需触发）

发现以下情况之一即触发紧急维护，跳过季度节奏：

- **安全告警**：govulncheck / npm audit 出现 critical
- **CI 红**：`.github/workflows/governance.yml` 任一核心 job 持续失败 3 次
- **契约破坏**：godocgen 产物与发布版本不一致
- **合规问题**：发现 `umadev` / `umacloud` / `goder.ai` 等违禁字符串残留（见 ADR-0013、CI 守卫）

紧急维护流程：

1. 起 issue `chore: emergency maintenance YYYY-MM-DD`
2. 在 24h 内修复并发布 patch 版本
3. 在 release notes 顶部说明修复内容
4. 必要时追加 ADR 记录经验教训

## 5. 工具与脚本索引

| 用途 | 路径 |
|---|---|
| Perf budget | `website/scripts/perf-budget.mjs` |
| Lighthouse 配置 | `website/.lighthouserc.cjs` |
| A11y 测试 | `tests/a11y/*.spec.ts` + `tests/playwright.config.ts` |
| Go API 文档 | `internal/godocgen/main.go` + `cmd/godocgen/main.go` |
| Sitemap | `website/scripts/gen-sitemap.mjs` |
| 治理工作流 | `.github/workflows/governance.yml`（待 A7 落地） |

## 6. 与现有规范的关系

- **AGENTS.md §4（质量约束）**：阈值定义
- **ADR-0016**：本规范的来源决策
- **CONSTRAINTS.md**：补充具体阈值数值
- **aicoding_docs/docs/standards/general.md**：通用原则（根因优先、最小改动等）

## 7. 修订记录

| 日期 | 版本 | 变更 |
|---|---|---|
| 2026-09-05 | 0.1 | 初始版本（ADR-0016 落地） |
