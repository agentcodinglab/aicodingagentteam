package qualitygate

import (
	"context"
	"testing"
)

// mockCheck creates a check with a simple command.
// Uses cmd /c for Windows compatibility: "exit 0" passes, "exit 1" fails.
func mockCheck(name, script string, severity string) Check {
	return Check{Name: name, Command: []string{"cmd", "/c", script}, Timeout: 10, Severity: severity}
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
			{Name: "slow", Command: []string{"cmd", "/c", "ping -n 10 127.0.0.1"}, Timeout: 1, Severity: "blocking"},
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
