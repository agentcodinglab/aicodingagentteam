package qualitygate

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ProofPack bundles audit evidence into a zip archive for delivery verification.
// Contains: plan.json, verify.jsonl, scorecard.md.
func ProofPack(workdir, planID string, result Result) (string, error) {
	dir := filepath.Join(workdir, ".aicodingagentteam", "proof")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create proof dir: %w", err)
	}

	ts := time.Now().Format("20060102-150405")
	zipPath := filepath.Join(dir, fmt.Sprintf("proof-pack-%s-%s.zip", planID, ts))

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("create zip: %w", err)
	}
	defer func() { _ = zipFile.Close() }()

	w := zip.NewWriter(zipFile)
	defer func() { _ = w.Close() }()

	// Add scorecard.md
	scorecard := Scorecard(result)
	if err := addToZip(w, "scorecard.md", []byte(scorecard)); err != nil {
		return "", err
	}

	// Add verify.jsonl if exists
	auditDir := filepath.Join(workdir, ".aicodingagentteam", "audit")
	verifyPath := filepath.Join(auditDir, "verify.jsonl")
	if data, err := os.ReadFile(verifyPath); err == nil {
		if err := addToZip(w, "verify.jsonl", data); err != nil {
			return "", err
		}
	}

	// Add plan.json if exists
	planPath := filepath.Join(workdir, ".aicodingagentteam", "plan.json")
	if data, err := os.ReadFile(planPath); err == nil {
		if err := addToZip(w, "plan.json", data); err != nil {
			return "", err
		}
	}

	// Add delivery-summary.md
	summary := fmt.Sprintf("# Delivery Summary\n\n- Plan ID: %s\n- Score: %d/100\n- Passed: %v\n- Generated: %s\n",
		planID, result.Score, result.Passed, time.Now().Format(time.RFC3339))
	if err := addToZip(w, "delivery-summary.md", []byte(summary)); err != nil {
		return "", err
	}

	return zipPath, nil
}

func addToZip(w *zip.Writer, name string, data []byte) error {
	f, err := w.Create(name)
	if err != nil {
		return fmt.Errorf("create zip entry %s: %w", name, err)
	}
	_, err = f.Write(data)
	return err
}
