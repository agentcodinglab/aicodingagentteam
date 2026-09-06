package scheduler

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/internal/host/codex"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
)

// stubBinary returns the absolute path to the codex stub binary for the
// current OS, skipping the test if it cannot be made executable (non-CI).
func stubBinary(t *testing.T) string {
	t.Helper()
	name := "codex"
	if runtime.GOOS == "windows" {
		name = "codex.cmd"
	}
	// Resolve relative to the testdata dir using the caller file location.
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
	p := filepath.Join(root, "testdata","stubbin",name)
	if _, err := os.Stat(p); err != nil {
		t.Skipf("stub binary not found: %s",p)
	}
	return p
}

// TestScheduler_HostE2E_StubBinary drives the real codex driver against the
// stub binary so the writer-node dispatch path (Director -> scheduler -> driver
// -> subprocess -> stdout artifact) is exercised end-to-end without an API key.
func TestScheduler_HostE2E_StubBinary(t *testing.T) {
	ws := t.TempDir()
	drv := codex.New(codex.WithBinary(stubBinary(t)), codex.WithTimeout(30))
	s := NewWithDriver(ws, drv)

	plan := &types.Plan{
		ID: "stub-e2e",
		Nodes: []types.TaskNode{{
			ID: "w1",Phase: types.PhaseFrontend, Role: types.RoleFrontend, Writer: true,
			ArtifactsOut: []string{"src/generated/stub_output.go"},
		}},
	}

	res, err := s.Execute(context.Background(), plan)
	if err != nil {
		t.Fatalf("execute: %v",err)
	}
	if res.Parked {
		t.Error("expected not parked")
	}
	// The host stdout artifact must be persisted under the workspace.
	artPath := filepath.Join(ws, ".aicodingagentteam","host","w1.txt")
	b, rerr := os.ReadFile(artPath)
	if rerr != nil {
		t.Fatalf("read host artifact: %v",rerr)
	}
	if len(b) == 0 {
		t.Error("host artifact empty")
	}
	// Planned artifacts should still be recorded.
	found := false
	for _, a := range res.Artifacts {
		if a == "src/generated/stub_output.go" {
			found = true
		}
	}
	if !found {
		t.Error("planned artifact missing from result")
	}
}

// TestScheduler_HostE2E_NilDriver_NoPanic verifies that without a driver the
// legacy stub path still works and does not panic (backward compatibility).
func TestScheduler_HostE2E_NilDriver_NoPanic(t *testing.T) {
	ws := t.TempDir()
	s := New(ws)
	plan := &types.Plan{
		ID: "nil-drv",
		Nodes: []types.TaskNode{{
			ID: "w2",Phase: types.PhaseBackend, Role: types.RoleBackend, Writer: true,
			ArtifactsOut: []string{"src/app.go"},
		}},
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v",r)
		}
	}()
	res, err := s.Execute(context.Background(), plan)
	if err != nil {
		t.Fatalf("execute: %v",err)
	}
	if len(res.Artifacts) == 0 {
		t.Error("expected legacy artifacts")
	}
}
