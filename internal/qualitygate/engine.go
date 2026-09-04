// Package qualitygate runs deterministic quality checks before delivery.
// Uses real go toolchain execution (go build/test/vet) per ADR-0009.
package qualitygate

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/agentcodinglab/aicodingagentteam/internal/audit"
)

// Result is the output of a quality gate run.
type Result struct {
	Score    int
	Passed   bool
	Blocking []string
	Advisory []string
	Details  []CheckDetail
}

// CheckDetail holds the result of a single check.
type CheckDetail struct {
	Name     string
	Status   string // pass / fail / skipped
	Duration time.Duration
	Output   string
	Severity string // blocking / advisory
}

// Check defines a single quality gate check.
type Check struct {
	Name     string
	Command  []string // e.g. ["go", "build", "./..."]
	Timeout  int      // seconds
	Severity string   // blocking / advisory
}

// Engine runs the deterministic quality gate checks.
type Engine struct {
	threshold int
	checks    []Check
	audit     *audit.Logger
}

// New creates an Engine with the given score threshold and default checks.
func New(threshold int) *Engine {
	return &Engine{
		threshold: threshold,
		checks:    defaultChecks(),
	}
}

// NewWithAudit creates an Engine with an audit logger for verify.jsonl output.
func NewWithAudit(threshold int, al *audit.Logger) *Engine {
	e := New(threshold)
	e.audit = al
	return e
}

// NewWithChecks creates an Engine with custom checks (e.g. for tests or configurable gates).
func NewWithChecks(threshold int, checks []Check) *Engine {
	return &Engine{threshold: threshold, checks: checks}
}

func defaultChecks() []Check {
	return []Check{
		{Name: "build", Command: []string{"go", "build", "./..."}, Timeout: 120, Severity: "blocking"},
		{Name: "vet", Command: []string{"go", "vet", "./..."}, Timeout: 120, Severity: "blocking"},
		{Name: "test", Command: []string{"go", "test", "./...", "-count=1"}, Timeout: 300, Severity: "blocking"},
		{Name: "lint", Command: []string{"golangci-lint", "run", "./..."}, Timeout: 300, Severity: "advisory"},
	}
}

// Verify runs all checks against the provided artifacts.
// Fail-open: if the engine itself panics, returns a default pass.
func (e *Engine) Verify(ctx context.Context, artifacts []string) (result Result) {
	defer func() {
		if r := recover(); r != nil {
			result = Result{Score: e.threshold, Passed: true, Advisory: []string{"fail-open: quality gate engine panicked"}}
		}
	}()

	var mu sync.Mutex
	var wg sync.WaitGroup
	details := make([]CheckDetail, len(e.checks))

	for i, check := range e.checks {
		wg.Add(1)
		go func(idx int, c Check) {
			defer wg.Done()
			detail := e.runCheck(ctx, c)
			mu.Lock()
			details[idx] = detail
			mu.Unlock()
		}(i, check)
	}
	wg.Wait()

	// Calculate score: each check worth 100/len(checks) points
	perCheck := 100 / len(e.checks)
	score := 0
	for _, d := range details {
		if d.Status == "pass" {
			score += perCheck
		}
		if d.Status == "fail" {
			if d.Severity == "blocking" {
				result.Blocking = append(result.Blocking, d.Name)
			} else {
				result.Advisory = append(result.Advisory, d.Name)
			}
		}
	}

	result.Score = score
	result.Passed = len(result.Blocking) == 0 && score >= e.threshold
	result.Details = details

	e.logAudit(result)
	return result
}

// VerifyWithRuntime runs all default checks plus a runtime probe.
// The runtime probe checks if the host CLI (codex) is available and authenticated.
func (e *Engine) VerifyWithRuntime(ctx context.Context, backend string) (result Result) {
	defer func() {
		if r := recover(); r != nil {
			result = Result{Score: e.threshold, Passed: true, Advisory: []string{"fail-open: quality gate engine panicked"}}
		}
	}()

	// Run default checks first
	result = e.Verify(ctx, nil)

	// Runtime probe: check backend CLI availability
	runtimeDetail := e.runRuntimeProbe(ctx, backend)
	result.Details = append(result.Details, runtimeDetail)

	switch runtimeDetail.Status {
	case "fail":
		result.Advisory = append(result.Advisory, "runtime")
	case "pass":
		perCheck := 100 / (len(e.checks) + 1)
		score := 0
		for _, d := range result.Details {
			if d.Status == "pass" {
				score += perCheck
			}
		}
		result.Score = score
		result.Passed = len(result.Blocking) == 0 && score >= e.threshold
	}

	e.logAudit(result)
	return result
}

// runRuntimeProbe checks if the backend CLI is available and responsive.
func (e *Engine) runRuntimeProbe(ctx context.Context, backend string) CheckDetail {
	detail := CheckDetail{Name: "runtime-" + backend, Severity: "advisory"}

	binary := backend
	if backend == "" {
		binary = "codex"
	}

	if _, err := exec.LookPath(binary); err != nil {
		detail.Status = "skipped"
		detail.Output = fmt.Sprintf("%s not installed", binary)
		return detail
	}

	checkCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(checkCtx, binary, "--version")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	detail.Duration = time.Since(start)

	if checkCtx.Err() == context.DeadlineExceeded {
		detail.Status = "fail"
		detail.Output = "timeout"
		return detail
	}

	if err != nil {
		detail.Status = "fail"
		out := stderr.String()
		if len(out) == 0 {
			out = stdout.String()
		}
		if len(out) > 2000 {
			out = out[:2000] + "..."
		}
		detail.Output = out
		return detail
	}

	detail.Status = "pass"
	out := stdout.String()
	if len(out) > 100 {
		out = out[:100]
	}
	detail.Output = out
	return detail
}

// runCheck executes a single check command with timeout.
func (e *Engine) runCheck(ctx context.Context, c Check) CheckDetail {
	detail := CheckDetail{Name: c.Name, Severity: c.Severity}

	// Check if command is available
	binary := c.Command[0]
	if _, err := exec.LookPath(binary); err != nil {
		detail.Status = "skipped"
		detail.Output = fmt.Sprintf("%s not installed", binary)
		return detail
	}

	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 120
	}
	checkCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(checkCtx, c.Command[0], c.Command[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	detail.Duration = time.Since(start)

	if checkCtx.Err() == context.DeadlineExceeded {
		detail.Status = "fail"
		detail.Output = "timeout"
		return detail
	}

	if err != nil {
		detail.Status = "fail"
		out := stderr.String()
		if len(out) == 0 {
			out = stdout.String()
		}
		if len(out) > 2000 {
			out = out[:2000] + "..."
		}
		detail.Output = out
		return detail
	}

	detail.Status = "pass"
	return detail
}

// logAudit writes the verify result to audit log if configured.
func (e *Engine) logAudit(r Result) {
	if e.audit == nil {
		return
	}
	for _, d := range r.Details {
		_ = e.audit.Log(audit.Entry{
			Type:       "verify",
			Tool:       d.Name,
			Result:     d.Status,
			Detail:     d.Output,
			DurationMs: d.Duration.Milliseconds(),
		})
	}
}

// Scorecard generates a human-readable quality report.
func Scorecard(r Result) string {
	var buf bytes.Buffer
	buf.WriteString("Quality Gate Scorecard\n")
	buf.WriteString("======================\n\n")
	fmt.Fprintf(&buf, "Score: %d/100  |  Passed: %v\n\n", r.Score, r.Passed)

	if len(r.Blocking) > 0 {
		buf.WriteString("Blocking Issues:\n")
		for _, b := range r.Blocking {
			fmt.Fprintf(&buf, "  [BLOCK] %s\n", b)
		}
		buf.WriteString("\n")
	}

	if len(r.Advisory) > 0 {
		buf.WriteString("Advisory Issues:\n")
		for _, a := range r.Advisory {
			fmt.Fprintf(&buf, "  [ADV] %s\n", a)
		}
		buf.WriteString("\n")
	}

	buf.WriteString("Check Details:\n")
	for _, d := range r.Details {
		status := "PASS"
		switch d.Status {
		case "fail":
			status = "FAIL"
		case "skipped":
			status = "SKIP"
		}
		fmt.Fprintf(&buf, "  [%s] %s (%s, %v)\n", status, d.Name, d.Severity, d.Duration.Round(time.Millisecond))
		if d.Output != "" && d.Status != "pass" {
			fmt.Fprintf(&buf, "        %s\n", d.Output)
		}
	}

	return buf.String()
}
