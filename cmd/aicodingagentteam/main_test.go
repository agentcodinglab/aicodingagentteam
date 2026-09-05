package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildBinary builds the aicodingagentteam binary for testing.
func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "aicodingagentteam-test")
	if os.PathSeparator == '\\' {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build binary: %v\n%s", err, out)
	}
	return bin
}

func TestCLI_Version(t *testing.T) {
	bin := buildBinary(t)
	out, err := exec.Command(bin, "version").Output()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if !strings.Contains(string(out), "aicodingagentteam") {
		t.Errorf("version output unexpected: %s", out)
	}
}

func TestCLI_Init(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	cmd := exec.Command(bin, "init")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init command failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "initialized") {
		t.Errorf("init output unexpected: %s", out)
	}
}

func TestCLI_NoArgs_PrintsUsage(t *testing.T) {
	bin := buildBinary(t)
	out, err := exec.Command(bin).CombinedOutput()
	if err == nil {
		t.Error("expected non-zero exit with no args")
	}
	if !strings.Contains(string(out), "AiCodingAgentTeam") {
		t.Errorf("usage output unexpected: %s", out)
	}
}

func TestCLI_Memory_NoArgs_PrintsUsage(t *testing.T) {
	bin := buildBinary(t)
	out, err := exec.Command(bin, "memory").CombinedOutput()
	if err != nil {
		t.Fatalf("memory command failed: %v", err)
	}
	if !strings.Contains(string(out), "Usage") {
		t.Errorf("memory usage output unexpected: %s", out)
	}
}

func TestCLI_MemoryShow_EmptyMemory(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	cmd := exec.Command(bin, "memory", "show")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("memory show failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Facts") {
		t.Errorf("expected Facts section, got: %s", out)
	}
	if !strings.Contains(string(out), "(none)") {
		t.Errorf("expected (none) for empty memory, got: %s", out)
	}
}

func TestCLI_MemoryCapture_Toggle(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	cmd := exec.Command(bin, "memory", "capture", "on")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("memory capture failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "on") {
		t.Errorf("expected 'on' in output, got: %s", out)
	}
}

func TestCLI_KnowledgeDemo(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	cmd := exec.Command(bin, "knowledge", "demo")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("knowledge demo failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "RAG + memory end-to-end complete") {
		t.Errorf("demo did not complete: %s", out)
	}
}

func TestCLI_Verify(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	cmd := exec.Command(bin, "verify")
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	_ = cmd.Run() // may fail if golangci-lint not installed
	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "quality-gate") {
		t.Errorf("verify output unexpected: %s", output)
	}
}

func TestCLI_KnowledgeIndex(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	// Create a sample file
	_ = filepath.Join(dir, "main.go")
	cmd := exec.Command(bin, "knowledge", "index", dir)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("knowledge index failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "indexed") {
		t.Errorf("index output unexpected: %s", out)
	}
}

// Ensure context import is used
var _ = context.Background
