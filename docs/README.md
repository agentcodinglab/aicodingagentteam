# AiCodingAgentTeam — 文档导读

> 基于 Golang 的 AI 编码编排 Agent Team 项目文档集
> 编写日期：2026-09-02

---

## 文档清单

| 序号 | 文档 | 内容 |
|---|---|---|
| 00 | [AGENTS.md](../AGENTS.md) | 项目规范总纲，引用 Vibe Coding 全部参考规范 |
| 01 | [需求分析文档.md](01-需求分析文档.md) | 产品定位、行业痛点、核心功能需求、业务流水线、设计原则 |
| 02 | [技术架构设计.md](02-技术架构设计.md) | 分层架构、Go 模块划分、容器化部署、A2A 协议、宿主驱动、角色编排、质量门禁、对外协议、TUI 架构 |
| 03 | [系统设计与实施规划.md](03-系统设计与实施规划.md) | 配置体系、命令系统、编译构建、技术选型、7 阶段实施规划、23 周里程碑、风险对策、源码路径 |
| 04 | [系统架构图.md](04-系统架构图.md) | PlantUML 源码：高层总览、五层流转、A2A 时序、容器拓扑、九角色编排 |
| 05 | [快速上手部署.md](05-快速上手部署.md) | 环境准备、容器化部署、单二进制部署、TUI 安装、首次运行、配置自定义、命令速查、故障排查 |
| — | [CONSTRAINTS.md](CONSTRAINTS.md) | 项目质量红线（覆盖率/性能/安全阈值） |
| — | [domain.md](domain.md) | 领域模型与统一语言（术语表、业务边界、流程、聚合） |
| — | [adr/](adr/) | 架构决策记录（ADR） |
| — | [spec/](spec/) | 功能规格与验收标准 |
| — | [plan/](plan/) | 实现计划 |

## 参考规范（外部引用）

本项目开发规范继承自 Vibe Coding 规范仓库，位于 `E:\javaproject\my\2026\aicoding_docs\`：

| 维度 | 规范文件 |
|---|---|
| 总纲 | `aicoding_docs\AGENTS.md` |
| 质量约束 | `aicoding_docs\docs\CONSTRAINTS.md` |
| 文档编写 | `aicoding_docs\docs\writing-guide.md` |
| 通用开发 | `aicoding_docs\docs\standards\general.md` |
| Go | `aicoding_docs\docs\standards\languages\go.md` |
| TypeScript | `aicoding_docs\docs\standards\languages\typescript.md` |
| 测试 | `aicoding_docs\docs\standards\testing\testing.md` |
| 安全 | `aicoding_docs\docs\standards\security\security.md` |
| Git | `aicoding_docs\docs\standards\git\git-workflow.md` |
| ADR 模板 | `aicoding_docs\docs\templates\ADR.md` |
| SPEC 模板 | `aicoding_docs\docs\templates\SPEC.md` |
| PLAN 模板 | `aicoding_docs\docs\templates\PLAN.md` |

## 项目一句话定位

AiCodingAgentTeam 是基于 **Golang** 开发的 AI 编码编排平台，**本身不拥有大模型**，调度 4 款 AI 编码 CLI（Claude-Code、Codex、OpenCode、DeepSeek-DSH），模拟软件开发团队角色，通过 **A2A 协议**协作，提供可审计、带质量门禁的软件交付流水线，**容器化部署**，对外暴露 **MCP/ACP/A2A** 三协议，提供 **TypeScript TUI 客户端**。

## 阅读建议

1. **所有人** → 先读 `AGENTS.md`（项目根目录）；
2. **产品/需求人员** → 读 `01-需求分析文档.md` + `domain.md`；
3. **架构师/技术负责人** → 读 `02-技术架构设计.md` + `04-系统架构图.md` + `CONSTRAINTS.md`；
4. **开发工程师** → 读 `03-系统设计与实施规划.md` 第 5 节 + 对应语言规范；
5. **DevOps** → 读 `05-快速上手部署.md`。
