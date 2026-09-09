# 质量门禁使用示例

> 关联：`docs/spec/quality-gate.md`、`docs/adr/ADR-0024-quality-gate-10-checks.md`

## 1. 基本用法

```go
package main

import (
	"context"
	"fmt"

	"github.com/agentcodinglab/aicodingagentteam/internal/qualitygate"
)

func main() {
	// 创建引擎，阈值 80
	e := qualitygate.New(80)

	// 执行全部门禁校验
	result := e.Verify(context.Background(), nil)

	// 输出结果
	fmt.Printf("Score: %d/100 | Passed: %v\n", result.Score, result.Passed)
	if len(result.Blocking) > 0 {
		fmt.Printf("Blocking: %v\n", result.Blocking)
	}
	if len(result.Advisory) > 0 {
		fmt.Printf("Advisory: %v\n", result.Advisory)
	}

	// 输出 scorecard
	fmt.Println(qualitygate.Scorecard(result))
}
```

## 2. 使用 Runtime 探针

```go
// 执行全部门禁 + runtime 探针
e := qualitygate.New(80)
result := e.VerifyWithRuntime(context.Background(), "codex")
// result.Details 包含 10 项校验 + 1 个 runtime 探针
```

## 3. 自定义校验

```go
e := qualitygate.NewWithChecks(50, []qualitygate.Check{
	{
		Name:     "my-check",
		Severity: "advisory",
		Func: func(ctx context.Context, artifacts []string) qualitygate.CheckDetail {
			detail := qualitygate.CheckDetail{
				Name:     "my-check",
				Severity: "advisory",
			}
			// 自定义校验逻辑
			if /* 检查通过 */ true {
				detail.Status = "pass"
				detail.Output = "all good"
			} else {
				detail.Status = "fail"
				detail.Output = "check failed"
			}
			return detail
		},
	},
})

result := e.Verify(context.Background(), nil)
```

## 4. 10 项校验清单

| # | 名称 | 类型 | 实现方式 |
|---|------|------|---------|
| 1 | prd-completeness | advisory | 文件检查：`output/*-prd.md` |
| 2 | architecture-design | advisory | 文件检查：`output/*-architecture.md` + `output/openapi.*` |
| 3 | api-contract-crosscheck | blocking | 占位（stub） |
| 4 | ui-smells | advisory | 文件检查：`output/**/*.tsx` |
| 5 | build | blocking | `go build ./...` |
| 6 | vet | blocking | `go vet ./...` |
| 7 | test | blocking | `go test ./... -count=1` |
| 8 | lint | advisory | `golangci-lint run ./...` |
| 9 | secret-leak | blocking | 占位（stub） |
| 10 | audit-log | advisory | 文件检查：`.aicodingagentteam/audit/*.jsonl` |

## 5. CLI 命令

```bash
# 执行全部门禁校验
aicodingagentteam verify

# 执行全部门禁 + runtime 探针
aicodingagentteam verify --runtime

# 输出 scorecard
aicodingagentteam report
```
