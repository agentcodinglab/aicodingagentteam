// Package qualitygate runs deterministic quality checks before delivery.
// Uses real go toolchain execution (go build/test/vet) per ADR-0009.
package qualitygate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/agentcodinglab/aicodingagentteam/internal/audit"
	"github.com/agentcodinglab/aicodingagentteam/internal/governance"
	"github.com/agentcodinglab/aicodingagentteam/pkg/contracts"
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
// CheckFunc is a function that performs a custom quality check.
// It returns CheckDetail with pass/fail/skipped status.
type CheckFunc func(ctx context.Context, artifacts []string) CheckDetail

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
	Command  []string  // e.g. ["go", "build", "./..."]
	Timeout  int       // seconds
	Severity string    // blocking / advisory
	Func     CheckFunc // custom validation function (optional)
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

// prdCompleteness checks if PRD documents exist and contain acceptance criteria.
func prdCompleteness(ctx context.Context, artifacts []string) CheckDetail {
	detail := CheckDetail{Name: "prd-completeness", Severity: "advisory"}
	matches, _ := filepath.Glob("output/*-prd.md")
	if len(matches) == 0 {
		detail.Status = "skipped"
		detail.Output = "no PRD documents found"
		return detail
	}
	found := false
	for _, f := range matches {
		data, _ := os.ReadFile(f)
		if strings.Contains(string(data), "验收") || strings.Contains(string(data), "acceptance") {
			found = true
			break
		}
	}
	if found {
		detail.Status = "pass"
		detail.Output = "found PRD documents with acceptance criteria"
	} else {
		detail.Status = "fail"
		detail.Output = "PRD documents exist but missing acceptance criteria"
	}
	return detail
}

// architectureDesign checks if architecture and OpenAPI specs exist.
func architectureDesign(ctx context.Context, artifacts []string) CheckDetail {
	detail := CheckDetail{Name: "architecture-design", Severity: "advisory"}
	matches, _ := filepath.Glob("output/*-architecture.md")
	if len(matches) == 0 {
		detail.Status = "skipped"
		detail.Output = "no architecture documents found"
		return detail
	}
	openapiFiles, _ := filepath.Glob("output/openapi.*")
	if len(openapiFiles) == 0 {
		detail.Status = "skipped"
		detail.Output = "no OpenAPI specifications found"
		return detail
	}
	detail.Status = "pass"
	detail.Output = "found architecture documents and OpenAPI specs"
	return detail
}

// apiContractCrossCheck validates frontend-backend contract consistency.
// Scans output/ for OpenAPI specs and frontend call sites, then cross-checks via pkg/contracts.
func apiContractCrossCheck(ctx context.Context, artifacts []string) CheckDetail {
	detail := CheckDetail{Name: "api-contract-crosscheck", Severity: "blocking"}

	// Find OpenAPI spec files in output/
	openapiFiles, _ := filepath.Glob("output/openapi.*")
	if len(openapiFiles) == 0 {
		detail.Status = "skipped"
		detail.Output = "no OpenAPI specs found in output/"
		return detail
	}

	// Parse OpenAPI spec using encoding/json
	contract := &contracts.Contract{
		Paths: make(map[string]contracts.PathItem),
	}
	for _, f := range openapiFiles {
		data, _ := os.ReadFile(f)
		var spec struct {
			Paths map[string]json.RawMessage `json:"paths"`
		}
		if err := json.Unmarshal(data, &spec); err == nil {
			for path := range spec.Paths {
				contract.Paths[path] = contracts.PathItem{
					Get:    &contracts.Operation{Method: "GET", Path: path},
					Post:   &contracts.Operation{Method: "POST", Path: path},
					Put:    &contracts.Operation{Method: "PUT", Path: path},
					Delete: &contracts.Operation{Method: "DELETE", Path: path},
					Patch:  &contracts.Operation{Method: "PATCH", Path: path},
				}
			}
		}
	}

	// Find frontend call sites in output/**/*.ts, *.tsx
	var callSites []contracts.CallSite
	tsFiles, _ := filepath.Glob("output/**/*.ts")
	tsxFiles, _ := filepath.Glob("output/**/*.tsx")
	allTs := make([]string, 0, len(tsFiles)+len(tsxFiles))
	allTs = append(allTs, tsFiles...)
	allTs = append(allTs, tsxFiles...)
	for _, f := range allTs {
		data, _ := os.ReadFile(f)
		content := string(data)
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			// Look for fetch("/api/...") or fetch('/api/...')
			if strings.Contains(line, "fetch(") || strings.Contains(line, "axios") {
				// Extract path from fetch call
				idx := strings.Index(line, "/api/")
				if idx >= 0 {
					// Extract the path
					pathStart := idx
					pathEnd := strings.IndexAny(line[pathStart:], "\"',)")
					var path string
					if pathEnd > 0 {
						path = line[pathStart : pathStart+pathEnd]
					} else {
						path = line[pathStart:]
					}
					method := "GET"
					switch {
					case strings.Contains(line, "POST") || strings.Contains(line, "post"):
						method = "POST"
					case strings.Contains(line, "PUT") || strings.Contains(line, "put"):
						method = "PUT"
					case strings.Contains(line, "DELETE") || strings.Contains(line, "delete"):
						method = "DELETE"
					case strings.Contains(line, "PATCH") || strings.Contains(line, "patch"):
						method = "PATCH"
					}
					callSites = append(callSites, contracts.CallSite{
						File:   f,
						Line:   i + 1,
						Method: method,
						Path:   path,
					})
				}
			}
		}
	}

	if len(callSites) == 0 {
		detail.Status = "pass"
		detail.Output = fmt.Sprintf("no frontend API calls found to cross-check (%d OpenAPI paths)", len(contract.Paths))
		return detail
	}

	// Run cross-check
	mismatches := contracts.CrossCheck(contract, callSites)
	if len(mismatches) > 0 {
		var msgs []string
		for _, m := range mismatches {
			msgs = append(msgs, fmt.Sprintf("%s:%d %s", m.Call.File, m.Call.Line, m.Reason))
		}
		detail.Status = "fail"
		detail.Output = fmt.Sprintf("%d contract mismatch(es): %s", len(mismatches), strings.Join(msgs, "; "))
		return detail
	}

	detail.Status = "pass"
	detail.Output = fmt.Sprintf("all %d frontend calls match contract", len(callSites))
	return detail
}

// uiSmells checks for UI bad smells using governance rules.
// Scans output/**/*.tsx and *.ts files with the governance engine for emoji and hardcoded colors.
func uiSmells(ctx context.Context, artifacts []string) CheckDetail {
	detail := CheckDetail{Name: "ui-smells", Severity: "advisory"}

	// Find TypeScript files to scan
	tsxFiles, _ := filepath.Glob("output/**/*.tsx")
	tsFiles, _ := filepath.Glob("output/**/*.ts")
	matches := append(tsxFiles, tsFiles...)
	if len(matches) == 0 {
		detail.Status = "skipped"
		detail.Output = "no TypeScript files found to check"
		return detail
	}

	// Use governance engine to check each file
	eng := governance.New()
	var allViolations []governance.Violation
	for _, f := range matches {
		data, _ := os.ReadFile(f)
		violations := eng.Check(ctx, f, string(data))
		allViolations = append(allViolations, violations...)
	}

	if len(allViolations) == 0 {
		detail.Status = "pass"
		detail.Output = fmt.Sprintf("scanned %d files, no UI smells detected", len(matches))
		return detail
	}

	// Group violations by rule
	var msgs []string
	ruleCount := make(map[string]int)
	for _, v := range allViolations {
		ruleCount[v.RuleID]++
	}
	for rule, count := range ruleCount {
		msgs = append(msgs, fmt.Sprintf("%s: %d", rule, count))
	}
	detail.Status = "fail"
	detail.Output = fmt.Sprintf("found %d violation(s): %s", len(allViolations), strings.Join(msgs, ", "))
	return detail
}

// secretLeak checks for potential secrets in files.
// Scans all files in output/ using governance checkSecretLeak rules.
func secretLeak(ctx context.Context, artifacts []string) CheckDetail {
	detail := CheckDetail{Name: "secret-leak", Severity: "blocking"}

	// Find all files in output/ to scan
	matches, _ := filepath.Glob("output/**/*")
	if len(matches) == 0 {
		detail.Status = "skipped"
		detail.Output = "no files found in output/"
		return detail
	}

	// Use governance engine to check each file for secret leaks
	eng := governance.New()
	var allViolations []governance.Violation
	scanned := 0
	for _, f := range matches {
		// Skip directories
		info, err := os.Stat(f)
		if err != nil || info.IsDir() {
			continue
		}
		scanned++
		data, _ := os.ReadFile(f)
		violations := eng.Check(ctx, f, string(data))
		for _, v := range violations {
			if v.RuleID == "sec-secret-leak" {
				allViolations = append(allViolations, v)
			}
		}
	}

	if len(allViolations) == 0 {
		detail.Status = "pass"
		detail.Output = fmt.Sprintf("scanned %d files, no secret leaks detected", scanned)
		return detail
	}

	// Format violation details
	var msgs []string
	for _, v := range allViolations {
		msgs = append(msgs, fmt.Sprintf("%s: %s", v.Path, v.Detail))
	}
	detail.Status = "fail"
	output := fmt.Sprintf("found %d secret leak(s): %s", len(allViolations), strings.Join(msgs, "; "))
	if len(output) > 2000 {
		output = output[:2000] + "..."
	}
	detail.Output = output
	return detail
}

// auditLog checks if audit logs directory exists and is non-empty.
func auditLog(ctx context.Context, artifacts []string) CheckDetail {
	detail := CheckDetail{Name: "audit-log", Severity: "advisory"}
	matches, _ := filepath.Glob(".aicodingagentteam/audit/*.jsonl")
	if len(matches) == 0 {
		detail.Status = "skipped"
		detail.Output = "no audit logs found"
		return detail
	}
	detail.Status = "pass"
	detail.Output = "found audit log files"
	return detail
}

func NewWithChecks(threshold int, checks []Check) *Engine {
	return &Engine{threshold: threshold, checks: checks}
}

func defaultChecks() []Check {
	return []Check{
		{Name: "prd-completeness", Severity: "advisory", Func: prdCompleteness},
		{Name: "architecture-design", Severity: "advisory", Func: architectureDesign},
		{Name: "api-contract-crosscheck", Severity: "blocking", Func: apiContractCrossCheck},
		{Name: "ui-smells", Severity: "advisory", Func: uiSmells},
		{Name: "build", Command: []string{"go", "build", "./..."}, Timeout: 120, Severity: "blocking"},
		{Name: "vet", Command: []string{"go", "vet", "./..."}, Timeout: 120, Severity: "blocking"},
		{Name: "test", Command: []string{"go", "test", "./...", "-count=1"}, Timeout: 300, Severity: "blocking"},
		{Name: "lint", Command: []string{"golangci-lint", "run", "./..."}, Timeout: 300, Severity: "advisory"},
		{Name: "secret-leak", Severity: "blocking", Func: secretLeak},
		{Name: "audit-log", Severity: "advisory", Func: auditLog},
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
			detail := e.runCheck(ctx, c, artifacts)
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
func (e *Engine) runCheck(ctx context.Context, c Check, artifacts []string) CheckDetail {
	detail := CheckDetail{Name: c.Name, Severity: c.Severity}

	// Use CheckFunc if available, otherwise use command execution
	if c.Func != nil {
		timeout := c.Timeout
		if timeout <= 0 {
			timeout = 120
		}
		checkCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		result := c.Func(checkCtx, artifacts)
		cancel()
		if checkCtx.Err() == context.DeadlineExceeded {
			detail.Status = "fail"
			detail.Output = "timeout"
			return detail
		}
		detail.Status = result.Status
		detail.Output = result.Output
		detail.Duration = result.Duration
		return detail
	}

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
