package scheduler

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/internal/host/opencode"
	"github.com/agentcodinglab/aicodingagentteam/internal/types"
)

// opencodeStubBinary returns the absolute path to the opencode-acp stub binary,
// skipping the test if it cannot be found.
func opencodeStubBinary(t *testing.T) string {
	t.Helper()
	name := "opencode-acp"
	if runtime.GOOS == "windows" {
		name = "opencode-acp.cmd"
	}
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
	p := filepath.Join(root, "testdata", "stubbin", name)
	if _, err := os.Stat(p); err != nil {
		t.Skipf("opencode stub binary not found: %s", p)
	}
	return p
}

// TestScheduler_HostE2E_OpenCode_StubBinary drives the real opencode driver
// against the ACP stub binary, exercising the writer-node dispatch path
// (Director -> scheduler -> opencode driver -> ACP stub -> stdout artifact)
// end-to-end without an API key.
func TestScheduler_HostE2E_OpenCode_StubBinary(t *testing.T) {
	ws := t.TempDir()
	drv := opencode.New(
		opencode.WithBinary(opencodeStubBinary(t)),
		opencode.WithTimeout(30),
	)
	s := NewWithDriver(ws, drv)

	plan := &types.Plan{
		ID: "opencode-stub-e2e",
		Nodes: []types.TaskNode{{
			ID:           "ow1",
			Phase:        types.PhaseFrontend,
			Role:         types.RoleFrontend,
			Writer:       true,
			ArtifactsOut: []string{"src/generated/opencode_output.go"},
		}},
	}

	res, err := s.Execute(context.Background(), plan)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if res.Parked {
		t.Error("expected not parked")
	}
	// The host stdout artifact must be persisted under the workspace.
	artPath := filepath.Join(ws, ".aicodingagentteam", "host", "ow1.txt")
	b, rerr := os.ReadFile(artPath)
	if rerr != nil {
		t.Fatalf("read host artifact: %v", rerr)
	}
	if len(b) == 0 {
		t.Error("host artifact empty")
	}
	// Planned artifacts should still be recorded.
	found := false
	for _, a := range res.Artifacts {
		if a == "src/generated/opencode_output.go" {
			found = true
		}
	}
	if !found {
		t.Error("planned artifact missing from result")
	}
}
