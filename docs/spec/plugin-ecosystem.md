# Spec: 插件生态 PoC

> 关联 ADR：ADR-0023（插件发现机制与 init 向导）
> 状态：✅ 已完成

## 1. 背景

ADR-0023 建立了 `pkg/plugin` 全局注册表和 `stub-driver` 示例。本 spec 扩展为完整 PoC：脚手架 CLI、第二个真实示例 driver、本地市场索引。

## 2. 功能清单

| # | 功能 | 实现 | 验收 |
|---|---|---|---|
| 1 | `plugin new <name> [backend]` 脚手架 | `cmd/aicodingagentteam/plugin.go` | 生成可编译 driver.go + 激活说明 |
| 2 | `plugin search <query>` 搜索本地索引 | `pkg/plugin/marketplace.go` | 无索引文件时返回空不报错 |
| 3 | `plugin list` 列出本地索引 | `pkg/plugin/marketplace.go` | 无索引文件时返回空不报错 |
| 4 | gemini-cli 示例 driver | `pkg/plugin/examples/gemini-driver/` | 自注册 + Capabilities + AuthStatus |
| 5 | 本地市场索引（CRUD） | `pkg/plugin/marketplace.go` | Add/Load/Search/Remove |

## 3. 验收标准

- [x] `aicodingagentteam plugin new my-driver my-backend` 生成可编译 driver.go
- [x] `aicodingagentteam plugin search query` 在无索引时返回空不 panic
- [x] `aicodingagentteam plugin list` 在无索引时返回空不 panic
- [x] gemini-driver 自注册到 plugin registry
- [x] Marketplace.Add/Load/Search/Remove 全部测试通过
- [x] `go build ./...` + `go vet ./...` 通过
- [x] `go test ./pkg/plugin/... -cover` 通过
