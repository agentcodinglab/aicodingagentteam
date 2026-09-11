// Package qwendriver is an example community-contributed host driver for
// the Qwen CLI. It demonstrates the self-registration pattern with a
// realistic backend name, capabilities, and model info.
//
// Usage:
//
//	import _ "github.com/agentcodinglab/aicodingagentteam/pkg/plugin/examples/qwen-driver"
//
// After this import, host.NewRegistry() will include the qwen driver.
package qwendriver

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/agentcodinglab/aicodingagentteam/pkg/plugin"
	"github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

const Backend runtime.Backend = "qwen"

type driver struct {
	binary string
}

func init() {
	plugin.Register(Backend, &driver{binary: "qwen"})
}

func (d *driver) StartSession(ctx context.Context, opts runtime.SessionOpts) (runtime.SessionID, error) {
	return runtime.SessionID("qwen-session"), nil
}

func (d *driver) DestroySession(ctx context.Context, id runtime.SessionID) error { return nil }

func (d *driver) SendTask(ctx context.Context, id runtime.SessionID, task runtime.TaskPayload) (<-chan runtime.Event, error) {
	ch := make(chan runtime.Event, 4)
	go func() {
		defer close(ch)
		ch <- runtime.Event{Type: runtime.EventStart, Content: "qwen task started"}

		binary := d.binary
		if _, err := exec.LookPath(binary); err != nil {
			ch <- runtime.Event{Type: runtime.EventError, Content: fmt.Sprintf("%s not installed", binary), Err: err}
			return
		}

		ch <- runtime.Event{Type: runtime.EventDone, Content: fmt.Sprintf("qwen processed: %s", task.Instruction)}
	}()
	return ch, nil
}

func (d *driver) Capabilities() runtime.HostCapabilities {
	return runtime.HostCapabilities{
		SessionResume: false,
		ToolCalls:     true,
		WebSearch:     false,
		WriteHook:     true,
	}
}

func (d *driver) ModelInfo() runtime.ModelInfo {
	return runtime.ModelInfo{
		ID:       "qwen-2.5-coder",
		Provider: "alibaba",
		Context:  128000,
	}
}

func (d *driver) Pause(ctx context.Context, id runtime.SessionID) error  { return nil }
func (d *driver) Resume(ctx context.Context, id runtime.SessionID) error { return nil }

func (d *driver) AuthStatus(ctx context.Context, id runtime.SessionID) (runtime.AuthStatus, error) {
	binary := d.binary
	if _, err := exec.LookPath(binary); err != nil {
		return runtime.AuthStatus{Ready: false, Detail: fmt.Sprintf("%s not installed", binary)}, nil
	}
	return runtime.AuthStatus{Ready: true, Detail: "qwen CLI detected"}, nil
}