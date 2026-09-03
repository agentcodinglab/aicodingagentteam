package knowledge

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestBM25RetrieveReturnsResults(t *testing.T) {
	e := New(false)
	e.IndexFile("docs/spec.md", "The coordinator orchestrates the pipeline scheduling.")
	e.IndexFile("docs/arch.md", "The architect designs API contracts and data models.")
	e.IndexFile("docs/test.md", "The QA agent generates test cases and coverage reports.")

	results := e.Retrieve(context.Background(), "coordinator pipeline scheduling", 3)
	if len(results) == 0 {
		t.Fatal("expected non-empty results for BM25 search")
	}
	if results[0].Path != "docs/spec.md" {
		t.Errorf("expected top result docs/spec.md, got %s", results[0].Path)
	}
	if results[0].Score <= 0 {
		t.Errorf("expected positive score, got %f", results[0].Score)
	}
}

func TestBM25NoDocsReturnsEmpty(t *testing.T) {
	e := New(false)
	results := e.Retrieve(context.Background(), "anything", 5)
	// Should return non-nil empty slice, not panic
	if results == nil {
		t.Error("expected non-nil slice, got nil")
	}
}

func TestBM25NoPanicOnEmptyQuery(t *testing.T) {
	e := New(false)
	e.IndexFile("test.go", "some content")
	results := e.Retrieve(context.Background(), "", 5)
	if results == nil {
		t.Error("expected non-nil slice for empty query")
	}
}

func TestBM25CJKTokenization(t *testing.T) {
	e := New(false)
	e.IndexFile("docs/cn.md", "协调器编排流水线调度")
	results := e.Retrieve(context.Background(), "协调器", 1)
	if len(results) == 0 {
		t.Error("expected results for CJK query")
	}
}

func TestBM25CloudEmbedDegradesWithoutEnvVars(t *testing.T) {
	_ = os.Unsetenv("AICODINGAGENTTEAM_ALLOW_CLOUD_EMBED")
	_ = os.Unsetenv("OPENAI_EMBED_KEY")
	e := New(true) // cloudEmbed=true but no env vars
	if e.IsCloudEmbed() {
		t.Error("cloud embed should be disabled without double env vars")
	}
	// Should still work via BM25
	e.IndexFile("test.go", "func main() {}")
	results := e.Retrieve(context.Background(), "main", 1)
	if len(results) == 0 {
		t.Error("BM25 should still work even if cloud embed degrades")
	}
}

func TestIndexDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test project coordinator pipeline"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, "ignore.txt"), []byte("not indexed"), 0o644)

	e := New(false)
	if err := e.IndexDirectory(context.Background(), tmpDir); err != nil {
		t.Fatalf("IndexDirectory error: %v", err)
	}
	if e.DocCount() != 2 {
		t.Errorf("expected 2 indexed docs, got %d", e.DocCount())
	}
}

func TestDocCount(t *testing.T) {
	e := New(false)
	if e.DocCount() != 0 {
		t.Errorf("expected 0 docs, got %d", e.DocCount())
	}
	e.IndexFile("a.go", "content a")
	e.IndexFile("b.go", "content b")
	if e.DocCount() != 2 {
		t.Errorf("expected 2 docs, got %d", e.DocCount())
	}
}

func TestRepomapScanDirectory(t *testing.T) {
	idx := NewIndex()
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main

func main() {}

type Config struct {
	Name string
}

func (c *Config) GetName() string {
	return c.Name
}
`), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte(`package main

func TestMain(t *testing.T) {}
`), 0o644)

	if err := idx.ScanDirectory(tmpDir); err != nil {
		t.Fatalf("ScanDirectory error: %v", err)
	}
	// main.go has: func main, type Config, method GetName (3 symbols)
	// main_test.go should be skipped
	if idx.Count() != 3 {
		t.Errorf("expected 3 symbols (func+type+method), got %d", idx.Count())
	}
}

func TestRepomapSkipsTestFiles(t *testing.T) {
	idx := NewIndex()
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "handler_test.go"), []byte(`package main
func TestHandler() {}
`), 0o644)

	_ = idx.ScanDirectory(tmpDir)
	if idx.Count() != 0 {
		t.Error("test files should not be indexed")
	}
}

func TestRepomapSummary(t *testing.T) {
	idx := NewIndex()
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644)

	_ = idx.ScanDirectory(tmpDir)
	summary := idx.Summary()
	if summary == "" {
		t.Error("summary should not be empty")
	}
	if !contains(summary, "func") {
		t.Error("summary should contain 'func'")
	}
}

func TestRepomapSymbolsByPath(t *testing.T) {
	idx := NewIndex()
	tmpDir := t.TempDir()
	p := filepath.Join(tmpDir, "main.go")
	_ = os.WriteFile(p, []byte("package main\nfunc foo() {}\ntype Bar struct{}\n"), 0o644)

	_ = idx.ScanDirectory(tmpDir)
	syms := idx.SymbolsByPath(p)
	if len(syms) != 2 {
		t.Errorf("expected 2 symbols for file, got %d", len(syms))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && indexString(s, substr) >= 0))
}

func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
