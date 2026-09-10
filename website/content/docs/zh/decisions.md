# 架构决策记录

本页面按时间顺序列出 AiCodingAgentTeam 项目的所有架构决策。

## 决策索引

| ADR | 标题 | 状态 | 日期 |
|-----|------|------|------|
| ADR-0001 | 选择 Go 而非 Rust | Accepted | 2026-09-02 |
| ADR-0002 | Redis 作为 A2A 消息总线 | Accepted | 2026-09-02 |
| ADR-0003 | 容器即角色隔离 | Accepted | 2026-09-02 |
| ADR-0004 | 单写者模型 | Accepted | 2026-09-02 |
| ADR-0005 | 不持有密钥 | Accepted | 2026-09-02 |
| ADR-0006 | Fail-open 治理策略 | Accepted | 2026-09-02 |
| ADR-0007 | 宿主 CLI 可行性（已被 ADR-0021 取代） | Superseded | 2026-09-04 |
| ADR-0008 | A2A 总线基础设施 | Accepted | 2026-09-02 |
| ADR-0009 | 质量门禁执行模型 | Accepted | 2026-09-02 |
| ADR-0010 | 并发与单写者 | Accepted | 2026-09-02 |
| ADR-0011 | 配置合并与环境变量覆盖 | Accepted | 2026-09-04 |
| ADR-0012 | 质量门禁详情传播 | Accepted | 2026-09-04 |
| ADR-0013 | RAG 记忆 Director 接入 | Accepted | 2026-09-05 |
| ADR-0014 | ACP/MCP 真实实现 | Accepted | 2026-09-05 |
| ADR-0015 | 端到端验证与安全 | Accepted | 2026-09-05 |
| ADR-0016 | 方向 A：体系守护 | Accepted | 2026-09-05 |
| ADR-0017 | 方向 D：RAG Demo | Accepted | 2026-09-05 |
| ADR-0018 | 方向 C：真实宿主 E2E | Accepted | 2026-09-06 |
| ADR-0019 | P4 流式 stdout | Accepted | 2026-09-06 |
| ADR-0020 | P6 ACP session/newTask | Accepted | 2026-09-06 |
| ADR-0021 | P5 OpenCode serve HTTP | Accepted | 2026-09-06 |
| ADR-0022 | 覆盖率缺口与路线图 | Accepted | 2026-09-06 |
| ADR-0023 | 插件发现与 init 向导 | Accepted | 2026-09-08 |
| ADR-0024 | 质量门禁 10 项校验 | Accepted | 2026-09-09 |

## 关键决策

### 不持有密钥（ADR-0005）
Coordinator 从不持有 API Key，鉴权完全交由底层 CLI 处理。这确保编排层在设计上就是安全的。

### Fail-open 治理（ADR-0006）
当治理引擎 panic 时，系统选择 fail-open（通过）而非阻断交付。这防止质量门禁基础设施成为单点故障。

### 容器即角色（ADR-0003）
每个角色 Agent 运行在独立容器中，可独立伸缩和替换。角色逻辑不跨容器直接调用。

### CheckFunc 扩展（ADR-0024）
质量门禁支持通过 `CheckFunc` 注册自定义校验函数，实现文件检查、治理扫描和契约交叉校验与 CLI 命令并行。