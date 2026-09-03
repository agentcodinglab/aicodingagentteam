// Package dsh implements the DeepSeek-DSH host driver.
package dsh

import (
	"context"
	"fmt"

	"github.com/yourorg/aicodingagentteam/pkg/runtime"
)

// Driver is the DeepSeek-DSH CLI driver (ACP v1 / private protocol).
type Driver struct{}

func New() *Driver { return &Driver{} }

func (d *Driver) StartSession(ctx context.Context, opts runtime.SessionOpts) (runtime.SessionID, error) {
	return runtime.SessionID("dsh-session"), nil
}
func (d *Driver) DestroySession(ctx context.Context, id runtime.SessionID) error { return nil }
func (d *Driver) SendTask(ctx context.Context, id runtime.SessionID, task runtime.TaskPayload) (<-chan runtime.Event, error) {
	ch := make(chan runtime.Event, 1)
	go func() {
		defer close(ch)
		ch <- runtime.Event{Type: runtime.EventDone, Content: fmt.Sprintf("dsh processed: %s", task.Instruction)}
	}()
	return ch, nil
}
func (d *Driver) Capabilities() runtime.HostCapabilities {
	return runtime.HostCapabilities{SessionResume: true, ToolCalls: true, WebSearch: true, WriteHook: false}
}
func (d *Driver) ModelInfo() runtime.ModelInfo {
	return runtime.ModelInfo{ID: "deepseek-dsh", Provider: "deepseek", Context: 128000}
}
func (d *Driver) Pause(ctx context.Context, id runtime.SessionID) error  { return nil }
func (d *Driver) Resume(ctx context.Context, id runtime.SessionID) error { return nil }
func (d *Driver) AuthStatus(ctx context.Context, id runtime.SessionID) (runtime.AuthStatus, error) {
	return runtime.AuthStatus{Ready: true, Detail: "logged in"}, nil
}
