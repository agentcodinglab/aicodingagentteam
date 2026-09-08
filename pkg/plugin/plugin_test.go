package plugin

import (
	"context"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

// stubDriver is a minimal Runtime implementation for testing.
type stubDriver struct {
	id string
}

func (d *stubDriver) StartSession(ctx context.Context, opts runtime.SessionOpts) (runtime.SessionID, error) {
	return runtime.SessionID(d.id), nil
}
func (d *stubDriver) DestroySession(ctx context.Context, id runtime.SessionID) error  { return nil }
func (d *stubDriver) SendTask(ctx context.Context, id runtime.SessionID, task runtime.TaskPayload) (<-chan runtime.Event, error) {
	ch := make(chan runtime.Event)
	close(ch)
	return ch, nil
}
func (d *stubDriver) Capabilities() runtime.HostCapabilities { return runtime.HostCapabilities{} }
func (d *stubDriver) ModelInfo() runtime.ModelInfo          { return runtime.ModelInfo{ID: d.id} }
func (d *stubDriver) Pause(ctx context.Context, id runtime.SessionID) error  { return nil }
func (d *stubDriver) Resume(ctx context.Context, id runtime.SessionID) error { return nil }
func (d *stubDriver) AuthStatus(ctx context.Context, id runtime.SessionID) (runtime.AuthStatus, error) {
	return runtime.AuthStatus{Ready: true, Detail: "test-ready"}, nil
}

func TestRegisterAndGet(t *testing.T) {
	Clear()
	defer Clear()

	backend := runtime.Backend("test-backend")
	drv := &stubDriver{id: "test"}
	Register(backend, drv)

	got, ok := Get(backend)
	if !ok {
		t.Fatal("expected driver to be registered")
	}
	if got.ModelInfo().ID != "test" {
		t.Errorf("unexpected driver: %v", got.ModelInfo())
	}
}

func TestGet_NotRegistered(t *testing.T) {
	Clear()
	defer Clear()

	_, ok := Get(runtime.Backend("nonexistent"))
	if ok {
		t.Error("expected not found for unregistered backend")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	Clear()
	defer Clear()

	Register(runtime.Backend("b1"), &stubDriver{id: "1"})
	Register(runtime.Backend("b2"), &stubDriver{id: "2"})

	all := All()
	if len(all) != 2 {
		t.Fatalf("expected 2 drivers, got %d", len(all))
	}

	// Mutating the returned map should not affect the global registry
	all[runtime.Backend("b3")] = &stubDriver{id: "3"}
	if Count() != 2 {
		t.Error("mutating returned map should not affect global registry")
	}
}

func TestBackends_Sorted(t *testing.T) {
	Clear()
	defer Clear()

	Register(runtime.Backend("zeta"), &stubDriver{})
	Register(runtime.Backend("alpha"), &stubDriver{})
	Register(runtime.Backend("mid"), &stubDriver{})

	bs := Backends()
	if len(bs) != 3 {
		t.Fatalf("expected 3 backends, got %d", len(bs))
	}
	if string(bs[0]) != "alpha" || string(bs[1]) != "mid" || string(bs[2]) != "zeta" {
		t.Errorf("backends not sorted: %v", bs)
	}
}

func TestCount(t *testing.T) {
	Clear()
	defer Clear()

	if Count() != 0 {
		t.Error("expected 0 after clear")
	}
	Register(runtime.Backend("a"), &stubDriver{})
	if Count() != 1 {
		t.Error("expected 1 after register")
	}
}

func TestRegister_Overwrites(t *testing.T) {
	Clear()
	defer Clear()

	backend := runtime.Backend("replace-me")
	Register(backend, &stubDriver{id: "v1"})
	Register(backend, &stubDriver{id: "v2"})

	got, _ := Get(backend)
	if got.ModelInfo().ID != "v2" {
		t.Error("Register should overwrite existing driver")
	}
}

func TestClear(t *testing.T) {
	Register(runtime.Backend("a"), &stubDriver{})
	Register(runtime.Backend("b"), &stubDriver{})
	Clear()

	if Count() != 0 {
		t.Error("expected 0 after clear")
	}
}
