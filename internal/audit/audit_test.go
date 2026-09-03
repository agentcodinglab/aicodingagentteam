package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLog_WritesJSONL(t *testing.T) {
	dir := t.TempDir()
	l := New(dir)
	err := l.Log(Entry{Type: "test", Detail: "unit test"})
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "events.jsonl"))
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("audit log should not be empty")
	}
}
