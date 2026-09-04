# ADR-0011: config.Load 合并默认值 + 环境变量覆盖

> 状态：已接受
> 日期：2026-09-04
> 关联规范：`docs/CONSTRAINTS.md`、`aicoding_docs/docs/standards/general.md` 根因优先原则

## 背景

`config.Load()` 原实现用空结构体 unmarshal 文件内容，导致部分配置文件（如只设了 `quality.threshold`）会丢失其他字段的默认值（ports 变为 0）。

同时 TUI 客户端文档了 `AICODINGAGENTTEAM_PORT` 等环境变量，但 server 端从未读取这些变量，TUI 文档与实现不一致。

## 决策

1. **合并而非替换**：`Load` 以 `Default()` 为基底 unmarshal 覆盖，缺失字段保留默认值。
2. **环境变量覆盖**：新增 `applyEnvOverrides`，支持 `AICODINGAGENTTEAM_PORT`、`AICODINGAGENTTEAM_BACKEND`、`AICODINGAGENTTEAM_QUALITY_THRESHOLD`，在所有返回路径执行（含无文件、坏 JSON 的 fallback）。
3. **优先级**：env 覆盖 default 值；文件值覆盖 default 值；env 与文件并存时 env 胜出（文档化行为，便于 CI/运维动态调端口）。

## 后果

- 正面：部分配置文件不再丢失默认值；运维可通过 env 动态调端口而无需改文件。
- 负面：测试须清理 env 变量（`isolateHome` 清空 `AICODINGAGENTTEAM_*`），否则 server 设的 env 会污染 `go test` 的 config 测试。
- 已在 `internal/config/config_test.go` 的 `isolateHome` 中修复此副作用。