# 功能：治理引擎（Governance Engine）

> 文件名：`docs/spec/governance.md`
> 遵循模板：`aicoding_docs/docs/templates/SPEC.md`
> 关联 ADR：ADR-0006

---

## 背景与目标

治理引擎是 Pre-write 钩子和 CI 扫描的执行器，包含 113+ 条规则覆盖 UI、安全、前后端工程风险。fail-open 设计确保治理自身故障不阻断开发。已在 `internal/governance/governance.go` 中定义 Rule/Violation 结构和 4 条默认规则。

### 为谁解决什么问题

- 代码写入文件前拦截不安全/不规范代码（Pre-write 钩子）；
- CI/Pre-commit 阶段扫描全量代码违规；
- MCP 客户端可调用 `govern_file` 工具对单文件治理扫描。

## 用户故事

- 作为 Coordinator，我希望在写角色产出代码后执行治理扫描，以便在交付前拦截违规。
- 作为开发者，我希望治理引擎故障时不阻断我的开发流程（fail-open）。
- 作为 MCP 客户端，我希望调用 `govern_file` 工具对指定文件执行治理扫描。

## 功能描述

### 做什么

1. **规则结构**
   - 每条规则：ID、Severity（advisory/blocking）、Enabled（可配置开关）
   - 规则覆盖维度：
     - UI：emoji 图标、硬编码颜色、AI 模板代码、组件状态缺失
     - 安全：密钥泄露、SQL 注入、XSS、CSRF、不安全依赖
     - 后端：API 契约不匹配、错误处理缺失、鉴权缺失
     - 前端：fetch/axios 调用与 OpenAPI 不一致

2. **fail-open 设计**（ADR-0006）
   - 治理引擎内部异常（panic/解析失败/规则加载失败）→ 默认返回通过，记录 WARN 日志
   - 规则自身的违规检测不是 fail-open（违规如实报告）
   - 关键路径可配置 fail-close（如 CI 模式）

3. **触发入口**
   | 入口 | 时机 | 说明 |
   |---|---|---|
   | Pre-write 钩子 | 代码写入文件前 | 对接宿主 CLI 的文件写入回调，违规则阻止写入 |
   | CI / Pre-commit | 提交前扫描 | 全量文件扫描，违规则拒绝提交 |
   | MCP `govern_file` | 外部调用 | MCP 客户端对单文件执行扫描 |
   | Quality-Gate | 交付前扫描 | 作为质量门禁的子检查项 |

4. **配置**
   ```toml
   # .aicodingagentteam/rules.toml
   [disabled]
   clauses = ["ui-emoji-icon"]

   [exclusions]
   paths = ["src/legacy/**", "**/*.test.ts"]
   ```
   - 规则可独立关闭（disabled.clauses）
   - 路径可排除（exclusions.paths，支持 glob）
   - 每条规则可覆盖 Severity

5. **审计输出**
   - 违规记录写入 `.aicodingagentteam/audit/governance.jsonl`
   - 每条：ts、rule_id、severity、path、detail

6. **规则分组**（MVP 先实现核心规则，后续迭代补全 113+）
   | 分组 | 规则数 | MVP 先实现 |
   |---|---|---|
   | UI 坏味道 | ~20 | emoji、硬编码颜色 |
   | 安全 | ~30 | 密钥泄露、SQL 注入 |
   | 前后端契约 | ~15 | API 路径不匹配 |
   | 工程规范 | ~48 | TODO 占位、假数据 |

## 验收标准

- [x] `Engine.Check(ctx, path, content)` 返回 `[]Violation` 无 panic
- [x] 规则关闭后（disabled.clauses）不产生该规则的 Violation
- [x] 路径匹配 exclusions.paths 的文件不被扫描
- [x] fail-open：引擎内部异常时返回空 Violation 列表 + WARN 日志，不 panic
- [x] 密钥泄露规则能检测 `AKIA...`、`sk-...` 等常见模式
- [x] emoji 图标规则能检测代码中的 emoji 字符
- [x] 治理结果写入 `governance.jsonl` 审计日志
- [x] MCP `govern_file` 工具能对单文件执行扫描
- [x] CI 模式（fail-close）下 blocking 违规导致非零退出码

## 非目标

- 不实现全部 113 条规则（MVP 先实现 4 条核心规则，后续迭代补全）。
- 不实现 Pre-write 钩子的宿主对接（MVP 阶段仅 Quality-Gate 和 CI 触发）。
- 不实现规则的热加载（MVP 启动时加载配置，后续迭代加热更新）。

## 依赖与约束

- 前置依赖：`internal/audit`（审计日志）、`internal/config`（规则配置加载）
- 技术约束：规则匹配用正则，路径排除用 filepath.Match/glob
- 风险与缓解：
  - 正则误报 → 规则可配置关闭 + advisory 级别不阻断
  - 规则加载失败 → fail-open

## 关联

- 实现计划：`docs/plan/mvp-preparation.md`
- 相关 ADR：ADR-0006（fail-open）
- 质量约束：`docs/CONSTRAINTS.md`
- 源码：`internal/governance/governance.go`