// Package gemdriver is an example community-contributed host driver for
// the Gemini CLI. It demonstrates the self-registration pattern with a
// realistic backend name, capabilities, and model info.
//
// Usage:
//
//	import _ "github.com/agentcodinglab/aicodingagentteam/pkg/plugin/examples/gemini-driver"
//
// After this import, host.NewRegistry() will include the gemini driver.
package gemdriver

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/agentcodinglab/aicodingagentteam/pkg/plugin"
	"github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

const Backend runtime.Backend = "gemini"

type driver struct {
	binary string
}

func init() {
	plugin.Register(Backend, &driver{binary: "gemini"})
}

func (d *driver) StartSession(ctx context.Context, opts runtime.SessionOpts) (runtime.SessionID, error) {
	return runtime.SessionID("gemini-session"), nil
}

func (d *driver) DestroySession(ctx context.Context, id runtime.SessionID) error { return nil }

func (d *driver) SendTask(ctx context.Context, id runtime.SessionID, task runtime.TaskPayload) (<-chan runtime.Event, error) {
	ch := make(chan runtime.Event, 4)
	go func() {
		defer close(ch)
		ch <- runtime.Event{Type: runtime.EventStart, Content: "gemini task started"}

		binary := d.binary
		if _, err := exec.LookPath(binary); err != nil {
			ch <- runtime.Event{Type: runtime.EventError, Content: fmt.Sprintf("%s not installed", binary), Err: err}
			return
		}

		ch <- runtime.Event{Type: runtime.EventDone, Content: fmt.Sprintf("gemini processed: %s", task.Instruction)}
	}()
	return ch, nil
}

func (d *driver) Capabilities() runtime.HostCapabilities {
	return runtime.HostCapabilities{
		SessionResume: true,
		ToolCalls:     true,
		WebSearch:     true,
		WriteHook:     false,
	}
}

func (d *driver) ModelInfo() runtime.ModelInfo {
	return runtime.ModelInfo{
		ID:       "gemini-2.0-flash",
		Provider: "google",
		Context:  1000000,
	}
}

func (d *driver) Pause(ctx context.Context, id runtime.SessionID) error  { return nil }
func (d *driver) Resume(ctx context.Context, id runtime.SessionID) error { return nil }

func (d *driver) AuthStatus(ctx context.Context, id runtime.SessionID) (runtime.AuthStatus, error) {
	binary := d.binary
	if _, err := exec.LookPath(binary); err != nil {
		return runtime.AuthStatus{Ready: false, Detail: fmt.Sprintf("%s not installed", binary)}, nil
	}
	return runtime.AuthStatus{Ready: true, Detail: "gemini CLI detected"}, nil
}
