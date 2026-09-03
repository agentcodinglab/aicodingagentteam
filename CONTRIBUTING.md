# 贡献指南 — AiCodingAgentTeam

> 本文件补充 `aicoding_docs/docs/standards/git/git-workflow.md` 的项目专属约定。

## 开发流程

1. **Spec-first**：新功能先写 `docs/spec/`，遵循 `aicoding_docs/docs/templates/SPEC.md`
2. **Plan**：再写 `docs/plan/`，遵循 `aicoding_docs/docs/templates/PLAN.md`
3. **Implement**：按计划逐步实现，测试先行
4. **Verify**：`make all`（lint + vet + test + build）全绿
5. **ADR**：非平凡决策记 `docs/adr/`

## 提交规范

遵循 Conventional Commits（见 git-workflow.md）：

```
feat(coordinator): add DAG persistence to planner
fix(host): handle codex stderr noise filtering
docs(adr): record ADR-0007 host CLI spike
```

## 分支模型

- `main`：稳定主干，PR 合入
- `feat/*`：功能分支
- `fix/*`：修复分支

## 本地开发

```bash
make all       # lint + vet + test + build
make test      # 仅测试
make run       # 启动服务
make clean     # 清理产物
```

## 代码规范

- Go：遵循 `aicoding_docs/docs/standards/languages/go.md`
- TypeScript（TUI）：遵循 `aicoding_docs/docs/standards/languages/typescript.md`
- 通用：遵循 `aicoding_docs/docs/standards/general.md`
- 质量红线：`docs/CONSTRAINTS.md`，不得静默降级