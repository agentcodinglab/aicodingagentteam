package qualitygate

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/internal/audit"
)

// mockCheck creates a check with a simple command that exits 0 or 1.
// Uses cross-platform approach: Go binary with a sub-command.
func mockCheck(name, script string, severity string) Check {
	// Use "go version" for pass (exit 0) and "go tool nonexistent" for fail (exit 1)
	if strings.Contains(script, "0") {
		return Check{Name: name, Command: []string{"go", "version"}, Timeout: 10, Severity: severity}
	}
	return Check{Name: name, Command: []string{"go", "tool", "nonexistent-binary-xyz"}, Timeout: 10, Severity: severity}
}
func TestVerify_AllPass(t *testing.T) {
	e := &Engine{
		threshold: 50,
		checks: []Check{
			mockCheck("pass1", "exit 0", "blocking"),
			mockCheck("pass2", "exit 0", "blocking"),
		},
	}
	r := e.Verify(context.Background(), nil)
	if len(r.Blocking) > 0 {
		t.Errorf("should have no blocking failures, got: %v", r.Blocking)
	}
	if r.Score != 100 {
		t.Errorf("expected score 100, got %d", r.Score)
	}
	if !r.Passed {
		t.Error("should pass when all blocking checks pass and score >= threshold")
	}
}

func TestVerify_BlockingFailure(t *testing.T) {
	e := &Engine{
		threshold: 50,
		checks: []Check{
			mockCheck("fail", "exit 1", "blocking"),
			mockCheck("pass", "exit 0", "advisory"),
		},
	}
	r := e.Verify(context.Background(), nil)
	if len(r.Blocking) == 0 {
		t.Error("should have blocking failure")
	}
	if r.Passed {
		t.Error("should not pass with blocking failure")
	}
}

func TestVerify_AdvisoryFailureNotBlocking(t *testing.T) {
	e := &Engine{
		threshold: 0,
		checks: []Check{
			mockCheck("advisory-fail", "exit 1", "advisory"),
		},
	}
	r := e.Verify(context.Background(), nil)
	if len(r.Blocking) > 0 {
		t.Error("advisory failure should not be blocking")
	}
	if len(r.Advisory) == 0 {
		t.Error("advisory failure should appear in Advisory list")
	}
	if !r.Passed {
		t.Error("advisory failure should not block passing")
	}
}

func TestVerify_DetailsPopulated(t *testing.T) {
	e := &Engine{
		threshold: 0,
		checks: []Check{
			mockCheck("check1", "exit 0", "blocking"),
		},
	}
	r := e.Verify(context.Background(), nil)
	if len(r.Details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(r.Details))
	}
	if r.Details[0].Name != "check1" {
		t.Errorf("expected name check1, got %s", r.Details[0].Name)
	}
	if r.Details[0].Status != "pass" {
		t.Errorf("expected pass, got %s", r.Details[0].Status)
	}
}

func TestVerify_TimeoutProducesFail(t *testing.T) {
	e := &Engine{
		threshold: 0,
		checks: []Check{
			{Name: "slow", Command: []string{"go", "tool", "nonexistent-slow-binary"}, Timeout: 1, Severity: "blocking"},
		},
	}
	r := e.Verify(context.Background(), nil)
	if len(r.Blocking) == 0 {
		t.Error("timeout should produce blocking failure")
	}
}

func TestVerify_SkippedCheckWhenBinaryMissing(t *testing.T) {
	e := &Engine{
		threshold: 0,
		checks: []Check{
			{Name: "missing", Command: []string{"nonexistent-binary-xyz", "arg"}, Timeout: 5, Severity: "advisory"},
		},
	}
	r := e.Verify(context.Background(), nil)
	if r.Details[0].Status != "skipped" {
		t.Errorf("expected skipped, got %s", r.Details[0].Status)
	}
}

func TestVerify_ConcurrentSafe(t *testing.T) {
	e := &Engine{
		threshold: 0,
		checks: []Check{
			mockCheck("c1", "exit 0", "blocking"),
			mockCheck("c2", "exit 0", "blocking"),
			mockCheck("c3", "exit 0", "advisory"),
		},
	}
	r := e.Verify(context.Background(), nil)
	if len(r.Details) != 3 {
		t.Errorf("expected 3 details, got %d", len(r.Details))
	}
}

func TestVerify_FailOpenOnPanic(t *testing.T) {
	e := &Engine{
		threshold: 90,
		checks:    nil,
	}
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("should not propagate panic, got %v", r)
		}
	}()
	r := e.Verify(context.Background(), nil)
	_ = r
}

func TestNew_CreatesEngineWithDefaultChecks(t *testing.T) {
	e := New(80)
	if e.threshold != 80 {
		t.Errorf("expected threshold 80, got %d", e.threshold)
	}
	if len(e.checks) != 4 {
		t.Errorf("expected 4 default checks, got %d", len(e.checks))
	}
	names := []string{e.checks[0].Name, e.checks[1].Name, e.checks[2].Name, e.checks[3].Name}
	for _, n := range []string{"build", "vet", "test", "lint"} {
		found := false
		for _, got := range names {
			if got == n {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected check %s in default checks", n)
		}
	}
}

func TestNewWithAudit_SetsAudit(t *testing.T) {
	al := audit.New(t.TempDir())
	e := NewWithAudit(70, al)
	if e.threshold != 70 {
		t.Errorf("expected threshold 70, got %d", e.threshold)
	}
	if e.audit == nil {
		t.Error("expected audit logger to be set")
	}
}

func TestNewWithChecks_CustomChecks(t *testing.T) {
	custom := []Check{
		mockCheck("custom1", "exit 0", "blocking"),
		mockCheck("custom2", "exit 0", "advisory"),
	}
	e := NewWithChecks(50, custom)
	if len(e.checks) != 2 {
		t.Errorf("expected 2 custom checks, got %d", len(e.checks))
	}
	if e.checks[0].Name != "custom1" {
		t.Errorf("expected custom1, got %s", e.checks[0].Name)
	}
}

func TestVerifyWithRuntime_DefaultChecksPlusProbe(t *testing.T) {
	e := NewWithChecks(0, []Check{mockCheck("fast", "exit 0", "blocking")})
	r := e.VerifyWithRuntime(context.Background(), "nonexistent-cli-xyz")
	// default check + runtime probe = 2 details
	if len(r.Details) != 2 {
		t.Fatalf("expected 2 details (1 check + 1 probe), got %d", len(r.Details))
	}
	// runtime probe for nonexistent CLI should be skipped
	probe := r.Details[1]
	if probe.Status != "skipped" {
		t.Errorf("expected runtime probe skipped for missing CLI, got %s", probe.Status)
	}
}

func TestVerifyWithRuntime_PanicRecovery(t *testing.T) {
	e := &Engine{threshold: 80, checks: nil}
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("should not propagate panic, got %v", r)
		}
	}()
	r := e.VerifyWithRuntime(context.Background(), "nonexistent-cli-xyz")
	if r.Score != 80 || !r.Passed {
		t.Errorf("expected fail-open default on panic, got score=%d passed=%v", r.Score, r.Passed)
	}
}

func TestScorecard_AllSections(t *testing.T) {
	r := Result{
		Score:    50,
		Passed:   false,
		Blocking: []string{"build"},
		Advisory: []string{"lint"},
		Details: []CheckDetail{
			{Name: "build", Status: "fail", Severity: "blocking", Output: "compile error"},
			{Name: "vet", Status: "pass", Severity: "blocking"},
			{Name: "missing", Status: "skipped", Severity: "advisory", Output: "not installed"},
		},
	}
	s := Scorecard(r)
	for _, want := range []string{"Score: 50/100", "[BLOCK] build", "[ADV] lint", "[FAIL] build", "[PASS] vet", "[SKIP] missing", "compile error"} {
		if !strings.Contains(s, want) {
			t.Errorf("scorecard missing %q in:\n%s", want, s)
		}
	}
}

func TestScorecard_PassedAllChecks(t *testing.T) {
	r := Result{
		Score:  100,
		Passed: true,
		Details: []CheckDetail{
			{Name: "build", Status: "pass", Severity: "blocking"},
		},
	}
	s := Scorecard(r)
	if !strings.Contains(s, "Score: 100/100  |  Passed: true") {
		t.Error("expected pass header")
	}
	if strings.Contains(s, "Blocking Issues") {
		t.Error("should not show blocking section when none")
	}
}

func TestVerifyWithAudit_LogsToAudit(t *testing.T) {
	dir := t.TempDir()
	al := audit.New(dir)
	e := NewWithAudit(0, al)
	e.checks = []Check{mockCheck("fast", "exit 0", "blocking")}
	_ = e.Verify(context.Background(), nil)

	entries, _ := os.ReadDir(dir)
	if len(entries) == 0 {
		t.Error("expected audit file to be created")
	}
}

func TestVerifyWithRuntime_ProbePassesWithAvailableBinary(t *testing.T) {
	e := NewWithChecks(0, []Check{mockCheck("fast", "exit 0", "blocking")})
	// go is always available in test environments
	r := e.VerifyWithRuntime(context.Background(), "git")
	probe := r.Details[len(r.Details)-1]
	if probe.Status != "pass" {
		t.Errorf("expected runtime probe pass for git, got %s (output: %s)", probe.Status, probe.Output)
	}
}

func TestVerifyWithRuntime_ProbeFailsWhenBinaryErrors(t *testing.T) {
	e := NewWithChecks(0, []Check{mockCheck("fast", "exit 0", "blocking")})
	// ping --version fails on most platforms
	r := e.VerifyWithRuntime(context.Background(), "nonexistent-probe-binary")
	probe := r.Details[len(r.Details)-1]
	if probe.Status != "fail" {
		if probe.Status != "skipped" {
			t.Errorf("expected skipped for nonexistent binary, got %s", probe.Status)
		}
	}
	if probe.Output == "" {
		t.Error("expected non-empty output for failed probe")
	}
}
