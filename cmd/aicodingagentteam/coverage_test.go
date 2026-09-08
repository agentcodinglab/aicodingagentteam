package main

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/agentcodinglab/aicodingagentteam/internal/config"
	"github.com/agentcodinglab/aicodingagentteam/internal/knowledge"
	"github.com/agentcodinglab/aicodingagentteam/internal/memory"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrintUsage(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	printUsage()
	w.Close()
	os.Stdout = old
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	out := string(buf[:n])
	for _, want := range []string{"init", "run", "quick", "verify", "govern", "version", "knowledge"} {
		if !strings.Contains(out, want) {
			t.Errorf("usage missing %q", want)
		}
	}
}

func TestCmdInit(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	cmdInit(nil)
	for _, d := range []string{".aicodingagentteam", ".aicodingagentteam/audit", "output"} {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", d)
		}
	}
}

func TestPrintCheckDetails(t *testing.T) {
	details := []types.CheckSummary{
		{Name: "lint", Status: "fail", Output: "has issues"},
		{Name: "test", Status: "pass", Output: "ok"},
		{Name: "build", Status: "warn", Output: strings.Repeat("x", 300)},
	}
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	printCheckDetails(details)
	w.Close()
	os.Stdout = old
	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	out := string(buf[:n])
	if !strings.Contains(out, "FAIL") {
		t.Errorf("expected FAIL in output, got: %s", out)
	}
	if !strings.Contains(out, "WARN") {
		t.Errorf("expected WARN in output, got: %s", out)
	}
}

func TestCmdMemory_Show(t *testing.T) {
	ctx := context.Background()
	cmdMemory(ctx, []string{"show"})
}

func TestCmdMemory_Capture(t *testing.T) {
	ctx := context.Background()
	cmdMemory(ctx, []string{"capture", "on"})
	cmdMemory(ctx, []string{"recall", "on"})
}

func TestCmdKnowledge_NoArgs(t *testing.T) {
	ctx := context.Background()
	cmdKnowledge(ctx, nil)
}

func TestWriteDemoReport(t *testing.T) {
	dir := t.TempDir()
	chunks := []knowledge.Chunk{
		{Path: "a.go", Score: 0.9, Content: "hello"},
	}
	facts := []memory.Fact{
		{Key: "k1", Value: "v1", Source: "test"},
	}
	delivery := &types.Delivery{
		PlanID:    "test-plan",
		Score:     90,
		Passed:    true,
		Artifacts: []string{"a.go"},
	}
	writeDemoReport(dir, 1, "workspace", chunks, facts, delivery)
	if _, err := os.Stat(filepath.Join(dir, "demo-report.md")); os.IsNotExist(err) {
		t.Error("expected demo-report.md to exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "demo-report.json")); os.IsNotExist(err) {
		t.Error("expected demo-report.json to exist")
	}
}

func TestCmdKnowledge_Search(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	ctx := context.Background()
	cmdKnowledge(ctx, []string{"search", "main"})
}

func TestCmdKnowledge_Index(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	ctx := context.Background()
	cmdKnowledge(ctx, []string{"index", dir})
}

func TestCmdKnowledge_Demo(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	ctx := context.Background()
	knowledgeDemo(ctx)
}

func TestCmdGovern_InProcess(t *testing.T) {
	dir := t.TempDir()
	srcFile := filepath.Join(dir, "sample.go")
	if err := os.WriteFile(srcFile, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	ctx := context.Background()
	cmdGovern(ctx, []string{dir})
}

func TestCmdReport(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	ctx := context.Background()
	cfg, _ := config.Load()
	cmdReport(ctx, cfg)
}

func TestCmdMemory_RecallOff(t *testing.T) {
	ctx := context.Background()
	cmdMemory(ctx, []string{"recall", "off"})
}

func TestCmdMemory_CaptureOff(t *testing.T) {
	ctx := context.Background()
	cmdMemory(ctx, []string{"capture", "off"})
}

func TestCmdMemory_NoArgs(t *testing.T) {
	ctx := context.Background()
	cmdMemory(ctx, nil)
}

func TestToGateDetails_Nil(t *testing.T) {
	if got := toGateDetails(nil); got != nil {
		t.Errorf("expected nil for empty input, got %v", got)
	}
}

func TestTempDir(t *testing.T) {
	d := tempDir()
	if d == "" {
		t.Error("tempDir returned empty")
	}
}


func TestCmdInit_NonInteractive(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	cmdInit([]string{"-non-interactive"})
	cfgPath := filepath.Join(".aicodingagentteam", "config.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("config.json not created: %v", err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("config.json not valid JSON: %v", err)
	}
	def := cfg["default"].(map[string]interface{})
	if def["backend"] != "codex" {
		t.Errorf("expected codex, got %v", def["backend"])
	}
	q := cfg["quality"].(map[string]interface{})
	if int(q["threshold"].(float64)) != 90 {
		t.Errorf("expected 90, got %v", q["threshold"])
	}
}

func TestCmdInit_CreatesDirs(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	cmdInit([]string{"-non-interactive"})
	for _, d := range []string{".aicodingagentteam", ".aicodingagentteam/audit", "output"} {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", d)
		}
	}
}

func TestCmdBackends(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	cmdBackends(context.Background())
	w.Close()
	os.Stdout = old
	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	out := string(buf[:n])
	if !strings.Contains(out, "codex") {
		t.Errorf("expected codex in backends output: %s", out)
	}
	if !strings.Contains(out, "BACKEND") {
		t.Errorf("expected BACKEND header: %s", out)
	}
}

func TestPromptSelect_Default(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("\n"))
	got := promptSelect(reader, "test", []string{"a", "b"}, "a")
	if got != "a" {
		t.Errorf("expected default a, got %s", got)
	}
}

func TestPromptSelect_ValidChoice(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("b\n"))
	got := promptSelect(reader, "test", []string{"a", "b"}, "a")
	if got != "b" {
		t.Errorf("expected b, got %s", got)
	}
}

func TestPromptInt_Default(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("\n"))
	got := promptInt(reader, "test", 80)
	if got != 80 {
		t.Errorf("expected 80, got %d", got)
	}
}

func TestPromptInt_ValidInput(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("75\n"))
	got := promptInt(reader, "test", 90)
	if got != 75 {
		t.Errorf("expected 75, got %d", got)
	}
}
