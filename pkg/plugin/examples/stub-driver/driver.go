// Package stubdriver is an example community-contributed host driver.
// It demonstrates the self-registration pattern: importing this package
// (blank import) automatically registers a stub backend in the plugin registry.
//
// Usage:
//
//	import _ "github.com/agentcodinglab/aicodingagentteam/pkg/plugin/examples/stub-driver"
//
// After this import, host.NewRegistry() will include the stub driver.
package stubdriver

import (
	"context"
	"fmt"

	"github.com/agentcodinglab/aicodingagentteam/pkg/plugin"
	"github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

const Backend runtime.Backend = "stub-example"

type driver struct{}

func (d *driver) StartSession(ctx context.Context, opts runtime.SessionOpts) (runtime.SessionID, error) {
	return runtime.SessionID("stub-session"), nil
}
func (d *driver) DestroySession(ctx context.Context, id runtime.SessionID) error { return nil }
func (d *driver) SendTask(ctx context.Context, id runtime.SessionID, task runtime.TaskPayload) (<-chan runtime.Event, error) {
	ch := make(chan runtime.Event, 1)
	ch <- runtime.Event{Type: runtime.EventDone, Content: fmt.Sprintf("stub executed: %s", task.Instruction)}
	close(ch)
	return ch, nil
}
func (d *driver) Capabilities() runtime.HostCapabilities {
	return runtime.HostCapabilities{SessionResume: false, ToolCalls: false, WebSearch: false, WriteHook: false}
}
func (d *driver) ModelInfo() runtime.ModelInfo {
	return runtime.ModelInfo{ID: "stub-model", Provider: "example", Context: 4096}
}
func (d *driver) Pause(ctx context.Context, id runtime.SessionID) error  { return nil }
func (d *driver) Resume(ctx context.Context, id runtime.SessionID) error { return nil }
func (d *driver) AuthStatus(ctx context.Context, id runtime.SessionID) (runtime.AuthStatus, error) {
	return runtime.AuthStatus{Ready: true, Detail: "stub driver needs no auth"}, nil
}

func init() {
	plugin.Register(Backend, &driver{})
}
