package qualitygate

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestProofPack_CreatesZip(t *testing.T) {
	ws := t.TempDir()
	// Create prerequisite files
	auditDir := filepath.Join(ws, ".aicodingagentteam", "audit")
	_ = os.MkdirAll(auditDir, 0o755)
	_ = os.WriteFile(filepath.Join(auditDir, "verify.jsonl"), []byte(`{"type":"verify","result":"pass"}`+"\n"), 0o644)

	planDir := filepath.Join(ws, ".aicodingagentteam")
	_ = os.MkdirAll(planDir, 0o755)
	_ = os.WriteFile(filepath.Join(planDir, "plan.json"), []byte(`{"id":"plan-123","nodes":[]}`), 0o644)

	result := Result{Score: 95, Passed: true, Blocking: nil, Advisory: nil}

	zipPath, err := ProofPack(ws, "plan-123", result)
	if err != nil {
		t.Fatalf("ProofPack failed: %v", err)
	}

	// Verify zip exists
	if _, err := os.Stat(zipPath); err != nil {
		t.Fatalf("proof pack not created: %v", err)
	}

	// Open and verify contents
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer func() { _ = r.Close() }()

	expected := map[string]bool{
		"scorecard.md":        false,
		"verify.jsonl":        false,
		"plan.json":           false,
		"delivery-summary.md": false,
	}
	for _, f := range r.File {
		if _, ok := expected[f.Name]; ok {
			expected[f.Name] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("expected %s in proof pack, not found", name)
		}
	}
}

func TestProofPack_NoAuditFiles(t *testing.T) {
	ws := t.TempDir()
	// No audit files or plan.json — should still create zip with scorecard + summary
	result := Result{Score: 0, Passed: false, Blocking: []string{"build"}}

	zipPath, err := ProofPack(ws, "plan-failed", result)
	if err != nil {
		t.Fatalf("ProofPack failed: %v", err)
	}
	if _, err := os.Stat(zipPath); err != nil {
		t.Fatalf("proof pack not created: %v", err)
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer func() { _ = r.Close() }()

	// Should have at least scorecard.md and delivery-summary.md
	var hasScorecard, hasSummary bool
	for _, f := range r.File {
		if f.Name == "scorecard.md" {
			hasScorecard = true
		}
		if f.Name == "delivery-summary.md" {
			hasSummary = true
		}
	}
	if !hasScorecard {
		t.Error("scorecard.md missing from proof pack")
	}
	if !hasSummary {
		t.Error("delivery-summary.md missing from proof pack")
	}
}
