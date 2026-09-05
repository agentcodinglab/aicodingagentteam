# 实现计划：发布与安全就绪(F) + 质量缺口闭合(G)

> 文件名：`docs/plan/release-and-quality-gap.md`
> 关联：方向 F（发布与安全就绪）、方向 G（质量缺口闭合）

---

## F — 发布与安全就绪

### F1: 根 README.md
- 文件：`README.md`
- 内容：项目定位、架构总览、快速上手、命令速查、CI badge、架构图链接
- 验证：GitHub 首页渲染正常

### F2: CI 增 govulncheck
- 文件：`.github/workflows/ci.yml`
- 改动：go job 增 `govulncheck ./...` 步骤
- 验证：CI 通过且无高危漏洞

### F3: CI 增覆盖率上报
- 文件：`.github/workflows/ci.yml`
- 改动：test 步骤加 `-coverprofile`，增 codecov upload 步骤
- 验证：codecov badge 可用

### F4: goreleaser 跨平台构建
- 文件：`.goreleaser.yml`
- 内容：build targets (linux/darwin/windows, amd64/arm64)，archives，checksums，snapshot
- 验证：`goreleaser check` 通过

### F5: GitHub Release workflow
- 文件：`.github/workflows/release.yml`
- 改动：tag v* 触发 goreleaser + GitHub Release
- 验证：workflow 语法正确

### F6: LICENSE + CHANGELOG
- 文件：`LICENSE`、`CHANGELOG.md`
- 内容：MIT 许可证、Keep a Changelog 格式
- 验证：文件存在且格式正确

### F7: CONTRIBUTING + issue/PR 模板
- 文件：`CONTRIBUTING.md`、`.github/ISSUE_TEMPLATE/`、`.github/PULL_REQUEST_TEMPLATE.md`
- 验证：GitHub 自动识别模板

## G — 质量缺口闭合

### G1: RedisBus 单元测试
- 文件：`internal/a2a/redis_bus_test.go`
- 依赖：go-redis/miniredis 或 RedisBus 可测性重构
- 目标：a2a 32.2% → 80%+
- 验证：`go test ./internal/a2a/...` 通过

### G2: host/registry 测试
- 文件：`internal/host/registry_test.go`
- 目标：58.3% → 80%+
- 验证：`go test ./internal/host/...` 通过

### G3: MCP Server 真实 JSON-RPC
- 文件：`internal/mcp/mcp.go`
- 改动：替换 stub Serve()，实现 stdio JSON-RPC（govern_file/govern_dir）
- 验证：单元测试覆盖 MCP 工具调用

### G4: TUI 组件单元测试
- 文件：`tui/src/__tests__/`、`tui/package.json`
- 依赖：ink-testing-library + vitest
- 目标：TUI 覆盖 ≥70%
- 验证：`npm test -- --run` 通过

### G5: A2A 消息 p95 benchmark
- 文件：`internal/a2a/a2a_test.go`
- 改动：BenchmarkDelegateParallel
- 验证：benchmark 输出 p95 ≤500ms

## 关键路径
```
F1→F2→F3→F4→F5→F6→F7 (顺序，独立文件)
G1∥G2 (并行测试) → G3 (MCP) → G4 (TUI) → G5 (bench)
```