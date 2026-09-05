All notable changes to AiCodingAgentTeam are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] — 待发布

### Added — 新增
- Go 集成测试（quick edit + build + park/continue + API surface）
- proof-pack 压缩包生成（plan.json + verify.jsonl + scorecard.md + delivery-summary.md）
- 容错型治理注入测试（关闭引擎 / 排除路径 / 禁用规则）
- 单写者锁并发加锁测试
- A2A 协议 JSON Schema 校验（AgentCard / Task / Result / ProgressEvent 互转）
- CI：gitleaks 秘密信息扫描
- CI：Semgrep SAST 静态分析
- CI：Trivy 镜像扫描
- CLI 烟雾测试（version / init / memory show / knowledge demo / verify / knowledge index）
- TUI 组件渲染测试（StatusBar / ResultPanel / HelpPanel，ink-testing-library）
- RAG 规范 13 项验收点全部关闭
- 根 README.md（架构概览 + 快速上手）
- MIT LICENSE
- CONTRIBUTING.md
- GitHub Issue / PR 模板
- goreleaser.yml 跨平台二进制构建
- GitHub Release 工作流（tag 触发）
- CI govulncheck 漏洞扫描
- CI codecov 覆盖率报告
- RedisBus 单元测试（miniredis）
- host/registry 单元测试
- MCP Server 真实 JSON-RPC stdio 实现
- TUI 组件单元测试（ink-testing-library + vitest）
- A2A DelegateParallel p95 基准
- ACP Server 真实 stdio JSON-RPC 2.0（initialize / session/start / session/stop / session/list）
- CLI 子命令：\continue\ / \mcp\ / \2a serve\ / \cp\ / \memory show|capture|recall\ / \ci- claude / dsh 驱动棒测试（覆盖率 0% → 100%）
- mcp 覆盖率 77.9% → 89.6%（子目录 + 工具定义）
- goreleaser ldflags 打入 version / commit / date 元信息

### Changed — 调整
- golangci-lint-action v8 → v9（兼容 Node 20 弃用）
- GitHub Actions checkout v4 → v5 / setup-go v5 → v6 / setup-node v4 → v5

## [0.2.1] - 2026-09-05

### Fixed — 修复
- 网站：README 多语言导航收敛为单行（｜ 分隔）

## [0.2.0] - 2026-09-05

### Added — 新增
- 市场网站 \website/\uff08Next.js 14 App Router + Tailwind + next-intl）
- 9 语言静态文档站（en / zh / ja / ko / fr / de / ru / es / it）
- GitHub Pages 静态部署工作流（\.github/workflows/docs-site.yml\uff09
- 本地化 README 集 \website/README.{lang}.md\uff0c互相导航
- Stage 1 首页重设计（Hero / LogoCloud / Features / Stats / Code / Architecture / Quickstart / FinalCTA）
- Stage 2 暗色雷虹双色调设计（cyan + magenta），三套可变字体（Manrope / Space Grotesk / JetBrains Mono）
- Hero 3 页自动打字机终端演示（光标闪烁 + RESTART 按钮）
- IntersectionObserver Reveal 滚动入场动画
- FeatureGrid 与 ArchitectureDiagram 的 3D 鼠标随动 Tilt 包装
- Markdown 渲染文档页（react-markdown + remark-gfm），多语言前缀导航
- 贴近智能的根路径重定向（Accept-Language 检测 + localStorage 记乏）
- OG 图（\og.png\ 1200×630）与渐变 \avicon.svg
### Changed — 调整
- 文档组织：项目内部文档迁入 \docs/\uff0c公开文档放入 \website/content/docs/
## [0.1.0] - 2026-09-02

### Added — 新增
- Go 集成测试（quick edit + build + park/continue + API surface）
- proof-pack 压缩包生成（plan.json + verify.jsonl + scorecard.md + delivery-summary.md）
- 容错型治理注入测试（关闭引擎 / 排除路径 / 禁用规则）
- 单写者锁并发加锁测试
- A2A 协议 JSON Schema 校验（AgentCard / Task / Result / ProgressEvent 互转）
- CI：gitleaks 秘密信息扫描
- CI：Semgrep SAST 静态分析
- CI：Trivy 镜像扫描
- CLI 烟雾测试（version / init / memory show / knowledge demo / verify / knowledge index）
- TUI 组件渲染测试（StatusBar / ResultPanel / HelpPanel，ink-testing-library）
- RAG 规范 13 项验收点全部关闭
- Go 编排引擎：router → planner → scheduler → coordinator 五层流程
- 宿主驱动：Codex（真运行） / OpenCode（真运行） / Claude（棒） / DSH（棒）
- A2A 协议：InProcBus + RedisBus Pub/Sub
- 9 个团队角色，以 A2A 评审者注册
- 质量门禁引擎（golangci-lint + go vet + go test）
- 治理引擎：113+ 条规则，容错型设计
- RAG 知识引擎（BM25，纯 Go，零依赖）
- 项目记忆库（facts / pitfalls / lessons / recipes）
- gRPC API 与 proto 定义
- A2A HTTP 服务器，以 Agent Card 发现
- MCP / ACP 服务器棒
- TypeScript TUI 客户端（Ink + React + Zustand + gRPC）
- Docker Compose 部署（coordinator + redis + agents + hosts）
- ADR-0001 至 ADR-0012 架构决策记录
- 完整文档套餐（需求 / 架构 / 设计 / 部署）
