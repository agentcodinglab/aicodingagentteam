# ADR-0009: 质量门禁真实执行 Spike 结论

> 状态：已接受
> 日期：2026-09-02
> 关联计划：`docs/plan/mvp-preparation.md` A3

## 背景

质量门禁需在容器内执行真实构建/测试/lint，需验证执行可行性与耗时。

## 决策

### 实测数据

| 检查项 | 工具 | 退出码 | 耗时 |
|---|---|---|---|
| 构建 | `go build ./...` | 0 | 8.1s |
| 静态检查 | `go vet ./...` | 0 | 2.5s |
| 测试 | `go test ./...` | 0 | 9.5s |
| Lint | golangci-lint | 未安装 | — |
| 漏洞 | govulncheck | 未安装 | — |

### 结论

1. **go build/vet/test 均可程序化执行**，退出码语义正确，耗时可接受
2. **golangci-lint 和 govulncheck 未安装**，需在 CI 镜像中预装
3. **总耗时约 20s**（build+vet+test），远低于 CONSTRAINTS.md 的 5 分钟阈值
4. **执行方式**：`exec.Command("go", "build", "./...")` 捕获 stdout/stderr + 退出码，与 A1 Spike 的 CLI 调用模式一致

### 质量门禁实现策略

```go
type Check struct {
    Name    string
    Command []string  // ["go", "build", "./..."]
    Timeout int       // seconds
}
// 退出码 0 = pass, 非 0 = fail/blocking
```

## 后果

- 正面：质量门禁可直接用 exec.Command 调 go 工具链，无需额外框架
- 负面：容器内需预装 go + golangci-lint + govulncheck，镜像体积增大
- 后续：runtime-probe（启动应用访问路由）需单独实现，比 go test 复杂