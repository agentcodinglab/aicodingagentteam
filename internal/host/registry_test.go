package host

import (
	"context"
	"testing"

	"github.com/yourorg/aicodingagentteam/pkg/runtime"
)

func TestRegistry_GetAllBackends(t *testing.T) {
	r := NewRegistry()
	backends := r.List()
	if len(backends) != 4 {
		t.Errorf("expected 4 backends, got %d", len(backends))
	}
}

func TestRegistry_GetCodex(t *testing.T) {
	r := NewRegistry()
	d, err := r.Get(runtime.BackendCodex)
	if err != nil {
		t.Fatalf("get codex failed: %v", err)
	}
	caps := d.Capabilities()
	if !caps.SessionResume {
		t.Error("codex should support session resume")
	}
}

func TestRegistry_OpenCodeNoResume(t *testing.T) {
	r := NewRegistry()
	d, _ := r.Get(runtime.BackendOpenCode)
	if d.Capabilities().SessionResume {
		t.Error("opencode should NOT support session resume (capability honesty)")
	}
	err := d.Resume(context.Background(), "test")
	if err == nil {
		t.Error("opencode Resume should return error (not fake support)")
	}
}
