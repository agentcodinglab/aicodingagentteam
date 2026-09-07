package main

import (
	"github.com/agentcodinglab/aicodingagentteam/internal/godocgen"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGodocGen_GenerateSingleLocale(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "pkg")
	outDir := filepath.Join(dir, "out")
	os.MkdirAll(pkgDir, 0o755)
	os.MkdirAll(filepath.Join(pkgDir, "api"), 0o755)
	os.WriteFile(filepath.Join(pkgDir, "api", "server.go"), []byte("package api\n\ntype Server struct{}\n"), 0o644)
	opts := godocgen.Options{
		PkgDir:  pkgDir,
		OutDir:  outDir,
		Locale:  "en",
		Version: "test",
	}
	if err := godocgen.Generate(opts); err != nil {
		t.Fatalf("generate failed: %v", err)
	}
}

func TestGodocGen_GenerateMultipleLocales(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "pkg")
	outDir := filepath.Join(dir, "out")
	os.MkdirAll(pkgDir, 0o755)
	os.WriteFile(filepath.Join(pkgDir, "types.go"), []byte("package pkg\n\ntype Item struct{\n	Name string\n}\n"), 0o644)
	for _, locale := range []string{"en", "zh"} {
		opts := godocgen.Options{
			PkgDir:  pkgDir,
			OutDir:  outDir,
			Locale:  locale,
			Version: "test",
		}
		if err := godocgen.Generate(opts); err != nil {
			t.Fatalf("generate %s failed: %v", locale, err)
		}
	}
}

func TestGodocGen_BuildBinary(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "godocgen-test")
	if os.PathSeparator == '\\' {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build binary: %v %s", err, string(out))
	}
}

func TestGodocGen_CLI_Help(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "godocgen-test")
	if os.PathSeparator == '\\' {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build binary: %v %s", err, string(out))
	}
	out, err := exec.Command(bin, "--help").CombinedOutput()
	_ = err
	s := string(out)
	if !strings.Contains(s, "godocgen") {
		t.Errorf("expected godocgen in output, got: %s", s)
	}
}

func TestRunGodocGen(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "pkg")
	outDir := filepath.Join(dir, "out")
	os.MkdirAll(pkgDir, 0o755)
	os.WriteFile(filepath.Join(pkgDir, "types.go"), []byte("package pkg\n\ntype Item struct{ Name string }\n"), 0o644)
	if err := runGodocGen(pkgDir, outDir, "en,zh", "test"); err != nil {
		t.Fatalf("runGodocGen failed: %v", err)
	}
}
