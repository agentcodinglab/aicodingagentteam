package qualitygate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- prdCompleteness tests ---

func TestPrdCompleteness_SkippedWhenNoFiles(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := prdCompleteness(context.Background(), nil)
	if detail.Status != "skipped" {
		t.Errorf("expected skipped, got %s", detail.Status)
	}
}

func TestPrdCompleteness_PassWithAcceptanceCriteria(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "output")
	_ = os.MkdirAll(outDir, 0755)
	_ = os.WriteFile(filepath.Join(outDir, "test-prd.md"), []byte("# PRD\n验收标准: all features pass"), 0644)

	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := prdCompleteness(context.Background(), nil)
	if detail.Status != "pass" {
		t.Errorf("expected pass, got %s (output: %s)", detail.Status, detail.Output)
	}
}

func TestPrdCompleteness_FailWithoutAcceptanceCriteria(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "output")
	_ = os.MkdirAll(outDir, 0755)
	_ = os.WriteFile(filepath.Join(outDir, "test-prd.md"), []byte("# PRD\njust a doc without criteria"), 0644)

	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := prdCompleteness(context.Background(), nil)
	if detail.Status != "fail" {
		t.Errorf("expected fail, got %s", detail.Status)
	}
}

// --- architectureDesign tests ---

func TestArchitectureDesign_SkippedWhenNoArchDocs(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "output")
	_ = os.MkdirAll(outDir, 0755)

	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := architectureDesign(context.Background(), nil)
	if detail.Status != "skipped" {
		t.Errorf("expected skipped, got %s", detail.Status)
	}
}

func TestArchitectureDesign_SkippedWhenNoOpenAPI(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "output")
	_ = os.MkdirAll(outDir, 0755)
	_ = os.WriteFile(filepath.Join(outDir, "test-architecture.md"), []byte("# Architecture"), 0644)

	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := architectureDesign(context.Background(), nil)
	if detail.Status != "skipped" {
		t.Errorf("expected skipped, got %s", detail.Status)
	}
}

func TestArchitectureDesign_SkippedWhenNoArtifacts(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := architectureDesign(context.Background(), nil)
	if detail.Status != "skipped" {
		t.Errorf("expected skipped, got %s", detail.Status)
	}
}

func TestArchitectureDesign_PassWithBothFiles(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "output")
	_ = os.MkdirAll(outDir, 0755)
	_ = os.WriteFile(filepath.Join(outDir, "test-architecture.md"), []byte("# Architecture"), 0644)
	_ = os.WriteFile(filepath.Join(outDir, "openapi.json"), []byte("{\"openapi\":\"3.0\"}"), 0644)

	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := architectureDesign(context.Background(), nil)
	if detail.Status != "pass" {
		t.Errorf("expected pass, got %s", detail.Status)
	}
}

// --- apiContractCrossCheck tests ---

func TestApiContractCrossCheck_SkippedWhenNoOpenAPI(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := apiContractCrossCheck(context.Background(), nil)
	if detail.Status != "skipped" {
		t.Errorf("expected skipped, got %s", detail.Status)
	}
}

func TestApiContractCrossCheck_PassWhenNoFrontendCalls(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "output")
	_ = os.MkdirAll(outDir, 0755)
	_ = os.WriteFile(filepath.Join(outDir, "openapi.json"), []byte(`{"paths":{"/api/users":{}}}`), 0644)

	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := apiContractCrossCheck(context.Background(), nil)
	if detail.Status != "pass" {
		t.Errorf("expected pass, got %s (output: %s)", detail.Status, detail.Output)
	}
}

// --- uiSmells tests ---

func TestUiSmells_SkippedWhenNoTSFiles(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := uiSmells(context.Background(), nil)
	if detail.Status != "skipped" {
		t.Errorf("expected skipped, got %s", detail.Status)
	}
}

func TestUiSmells_PassWhenCleanFiles(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "output", "src")
	_ = os.MkdirAll(outDir, 0755)
	_ = os.WriteFile(filepath.Join(outDir, "App.tsx"), []byte("export function App() { return null }"), 0644)

	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := uiSmells(context.Background(), nil)
	if detail.Status != "pass" {
		t.Errorf("expected pass, got %s (output: %s)", detail.Status, detail.Output)
	}
}

func TestUiSmells_ScansTSAndTSXFiles(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "output", "src")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "App.tsx"), []byte("export function App() { return null }"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "bad.tsx"), []byte(`const color = "#00ff00"`), 0644); err != nil {
		t.Fatal(err)
	}

	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := uiSmells(context.Background(), nil)
	if detail.Status != "fail" {
		t.Fatalf("expected fail, got %s: %s", detail.Status, detail.Output)
	}
	if !strings.Contains(detail.Output, "scanned 2 files") && !strings.Contains(detail.Output, "1 violation") {
		t.Errorf("expected both TypeScript extensions to be scanned, got %s", detail.Output)
	}
}

func TestUiSmells_FailWhenHardcodedColorPresent(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "output", "src")
	_ = os.MkdirAll(outDir, 0755)
	_ = os.WriteFile(filepath.Join(outDir, "Bad.tsx"), []byte("export function Bad() { const color = \"#ff0000\"; return color }"), 0644)

	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := uiSmells(context.Background(), nil)
	if detail.Status != "fail" {
		t.Errorf("expected fail, got %s (output: %s)", detail.Status, detail.Output)
	}
}

// --- secretLeak tests ---

func TestSecretLeak_SkippedWhenNoFiles(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := secretLeak(context.Background(), nil)
	if detail.Status != "skipped" {
		t.Errorf("expected skipped, got %s", detail.Status)
	}
}

func TestSecretLeak_PassWhenCleanFiles(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "output", "src")
	_ = os.MkdirAll(outDir, 0755)
	_ = os.WriteFile(filepath.Join(outDir, "main.go"), []byte("package main\nfunc main() {}"), 0644)

	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := secretLeak(context.Background(), nil)
	if detail.Status != "pass" {
		t.Errorf("expected pass, got %s (output: %s)", detail.Status, detail.Output)
	}
}

func TestSecretLeak_FailWhenSecretPresent(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "output", "src")
	_ = os.MkdirAll(outDir, 0755)
	_ = os.WriteFile(filepath.Join(outDir, "config.go"), []byte(`package main
var token = "ghp_1234567890abcdefghijklmnopqrstuvwxyz"`), 0644)

	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := secretLeak(context.Background(), nil)
	if detail.Status != "fail" {
		t.Errorf("expected fail, got %s (output: %s)", detail.Status, detail.Output)
	}
}

// --- auditLog tests ---

func TestAuditLog_SkippedWhenNoAuditLogs(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := auditLog(context.Background(), nil)
	if detail.Status != "skipped" {
		t.Errorf("expected skipped, got %s", detail.Status)
	}
}

func TestAuditLog_PassWhenAuditLogsExist(t *testing.T) {
	dir := t.TempDir()
	auditDir := filepath.Join(dir, ".aicodingagentteam", "audit")
	_ = os.MkdirAll(auditDir, 0755)
	_ = os.WriteFile(filepath.Join(auditDir, "events.jsonl"), []byte(`{"type":"test"}`), 0644)

	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	detail := auditLog(context.Background(), nil)
	if detail.Status != "pass" {
		t.Errorf("expected pass, got %s", detail.Status)
	}
}

// --- runCheck CheckFunc branch test ---

func TestRunCheck_UsesCheckFuncWhenProvided(t *testing.T) {
	e := &Engine{threshold: 0}
	called := false
	check := Check{
		Name:     "custom-func-check",
		Severity: "advisory",
		Func: func(ctx context.Context, artifacts []string) CheckDetail {
			called = true
			return CheckDetail{
				Name:     "custom-func-check",
				Status:   "pass",
				Output:   "custom function executed",
				Severity: "advisory",
			}
		},
	}
	detail := e.runCheck(context.Background(), check, nil)
	if !called {
		t.Error("CheckFunc was not called")
	}
	if detail.Status != "pass" {
		t.Errorf("expected pass, got %s", detail.Status)
	}
	if detail.Output != "custom function executed" {
		t.Errorf("expected custom output, got %s", detail.Output)
	}
}

func TestRunCheck_PassesArtifactsToCheckFunc(t *testing.T) {
	e := &Engine{threshold: 0}
	want := []string{"src/main.ts", "output/openapi.json"}
	var got []string
	check := Check{
		Name:     "artifacts-check",
		Severity: "advisory",
		Func: func(ctx context.Context, artifacts []string) CheckDetail {
			got = append([]string(nil), artifacts...)
			return CheckDetail{Status: "pass", Severity: "advisory"}
		},
	}

	e.runCheck(context.Background(), check, want)
	if len(got) != len(want) {
		t.Fatalf("expected %d artifacts, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("artifact[%d]: expected %s, got %s", i, want[i], got[i])
		}
	}
}

func TestRunCheck_CheckFuncTimeout(t *testing.T) {
	e := &Engine{threshold: 0}
	check := Check{
		Name:     "slow-func-check",
		Timeout:  1,
		Severity: "blocking",
		Func: func(ctx context.Context, artifacts []string) CheckDetail {
			select {
			case <-ctx.Done():
				return CheckDetail{Status: "pass", Severity: "blocking"}
			case <-time.After(10 * time.Second):
				return CheckDetail{Status: "pass", Severity: "blocking"}
			}
		},
	}

	start := time.Now()
	detail := e.runCheck(context.Background(), check, nil)
	if detail.Status != "fail" || detail.Output != "timeout" {
		t.Fatalf("expected timeout failure, got status=%s output=%s", detail.Status, detail.Output)
	}
	if elapsed := time.Since(start); elapsed >= 2*time.Second {
		t.Fatalf("timeout was not enforced: %s", elapsed)
	}
}

func TestRunCheck_FallsBackToCommandWhenNoFunc(t *testing.T) {
	e := &Engine{threshold: 0}
	check := Check{
		Name:     "cmd-check",
		Command:  []string{"nonexistent-binary-xyz"},
		Timeout:  5,
		Severity: "advisory",
	}
	detail := e.runCheck(context.Background(), check, nil)
	if detail.Status != "skipped" {
		t.Errorf("expected skipped for missing binary, got %s", detail.Status)
	}
}

// --- defaultChecks tests ---

func TestDefaultChecks_ReturnsTenChecks(t *testing.T) {
	checks := defaultChecks()
	if len(checks) != 10 {
		t.Fatalf("expected 10 checks, got %d", len(checks))
	}

	expected := []string{
		"prd-completeness",
		"architecture-design",
		"api-contract-crosscheck",
		"ui-smells",
		"build",
		"vet",
		"test",
		"lint",
		"secret-leak",
		"audit-log",
	}

	for i, exp := range expected {
		if checks[i].Name != exp {
			t.Errorf("check[%d]: expected %s, got %s", i, exp, checks[i].Name)
		}
	}
}

func TestDefaultChecks_HasFuncForCustomChecks(t *testing.T) {
	checks := defaultChecks()

	funcChecks := []int{0, 1, 2, 3, 8, 9} // indices of CheckFunc-based checks
	for _, idx := range funcChecks {
		if checks[idx].Func == nil {
			t.Errorf("check[%d] (%s): expected non-nil Func", idx, checks[idx].Name)
		}
	}

	cmdChecks := []int{4, 5, 6, 7} // indices of Command-based checks
	for _, idx := range cmdChecks {
		if checks[idx].Func != nil {
			t.Errorf("check[%d] (%s): expected nil Func for command check", idx, checks[idx].Name)
		}
		if len(checks[idx].Command) == 0 {
			t.Errorf("check[%d] (%s): expected non-empty Command", idx, checks[idx].Name)
		}
	}
}
