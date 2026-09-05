# ADR-0015: 端到端验证 + 安全加固 + 发布打磨

> 状态：已接受
> 日期：2026-09-05
> 关联规范：`docs/CONSTRAINTS.md`（安全红线）、`docs/spec/coordinator.md`

## 背景

项目所有核心包覆盖率达标（≥80%），CLI 命令补齐，MCP/ACP 真实实现。但存在三类缺口：
1. 缺少 Go 端端到端集成测试，无法证明完整流水线可跑通
2. CI 缺少 gitleaks/Semgrep/trivy 安全扫描，不满足 CONSTRAINTS 安全红线
3. proof-pack 交付打包仅在 spec 中提及但未实现

## 决策

1. **E2E 集成测试**：在 `internal/coordinator/e2e_test.go` 中验证 quick edit 全链路、build 9 节点 DAG、park/continue 恢复、gRPC API 表面 4 个测试覆盖完整生命周期。
2. **proof-pack**：在 `internal/qualitygate/proof.go` 中实现 `ProofPack(workdir, planID, result)` 生成 zip（scorecard.md + verify.jsonl + plan.json + delivery-summary.md）。
3. **fail-open 故障注入**：测试关闭引擎、排除路径、禁用规则三种 fail-open 路径。
4. **并发写锁审计**：验证 write.lock 在执行后释放、parked 时保持。
5. **CI 安全加固**：gitleaks（密钥扫描）、Semgrep（SAST）、Trivy（容器镜像扫描），新增 container-scan job。
6. **CLI 冒烟测试**：exec.Command 驱动子进程测试 8 个 CLI 命令。
7. **TUI 组件渲染测试**：ink-testing-library 渲染 StatusBar/ResultPanel/HelpPanel。
8. **A2A 契约校验**：AgentCard/Task/Result/ProgressEvent JSON 序列化字段完整性 + round-trip。

## 后果

- 正面：E2E 测试证明 Route→Plan→Schedule→Verify→Finalize 闭环可跑通
- 正面：proof-pack 让交付物可审计可追溯
- 正面：CI 安全红线全部满足（govulncheck + gitleaks + Semgrep + Trivy）
- 正面：TUI 测试从 22 → 30（增 8 个组件渲染测试）
- 负面：CLI 子进程测试覆盖率显示 0%（Go 覆盖率工具不追踪子进程），但 8 个测试实际运行并验证 CLI 行为