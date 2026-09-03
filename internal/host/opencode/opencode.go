// Package opencode implements the OpenCode CLI host driver.
// It calls `opencode run --format json` via exec.Command in non-interactive mode.
// See ADR-0007 for feasibility verification.
package opencode

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

// Driver is the OpenCode CLI driver (opencode run --format json non-interactive mode).
type Driver struct {
	binary     string
	timeout    int
	maxRetries int
}

// Option configures a Driver.
type Option func(*Driver)

// WithBinary overrides the opencode binary path.
func WithBinary(path string) Option { return func(d *Driver) { d.binary = path } }

// WithTimeout sets the per-task timeout in seconds.
func WithTimeout(secs int) Option { return func(d *Driver) { d.timeout = secs } }

// WithMaxRetries sets the max retry count for transient failures.
func WithMaxRetries(n int) Option { return func(d *Driver) { d.maxRetries = n } }

// New creates an OpenCode driver. Defaults: binary="opencode", timeout=300, retries=3.
func New(opts ...Option) *Driver {
	d := &Driver{binary: "opencode", timeout: 300, maxRetries: 3}
	for _, o := range opts {
		o(d)
	}
	return d
}

// opencodeEvent is a single JSON event from opencode run --format json output.
type opencodeEvent struct {
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
	Part      struct {
		Type      string `json:"type"`
		Text      string `json:"text"`
		Reason    string `json:"reason"`
		MessageID string `json:"messageID"`
	} `json:"part"`
}

func (d *Driver) StartSession(ctx context.Context, opts runtime.SessionOpts) (runtime.SessionID, error) {
	return runtime.SessionID("opencode"), nil
}

func (d *Driver) DestroySession(ctx context.Context, id runtime.SessionID) error { return nil }

// SendTask runs `opencode run --format json <prompt>` and streams events.
// On transient errors it retries up to maxRetries times.
func (d *Driver) SendTask(ctx context.Context, id runtime.SessionID, task runtime.TaskPayload) (<-chan runtime.Event, error) {
	ch := make(chan runtime.Event, 8)

	go func() {
		defer close(ch)

		timeout := d.timeout
		if task.Timeout > 0 {
			timeout = task.Timeout
		}

		prompt := task.Instruction
		var lastErr error

		for attempt := 0; attempt <= d.maxRetries; attempt++ {
			if attempt > 0 {
				backoff := time.Duration(attempt*attempt) * 500 * time.Millisecond
				select {
				case <-ctx.Done():
					return
				case <-time.After(backoff):
				}
			}

			ch <- runtime.Event{Type: runtime.EventStart, Content: fmt.Sprintf("opencode attempt %d", attempt+1)}

			out, stderr, err := d.runOnce(ctx, prompt, timeout)
			if err == nil {
				// Parse JSON Lines and emit events
				for _, line := range strings.Split(out, "\n") {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					var ev opencodeEvent
					if err := json.Unmarshal([]byte(line), &ev); err != nil {
						continue
					}
					switch ev.Type {
					case "text":
						if ev.Part.Text != "" {
							ch <- runtime.Event{Type: runtime.EventMessage, Content: ev.Part.Text}
						}
					case "step_finish":
						// Task completed
					case "tool_call", "tool_result":
						ch <- runtime.Event{Type: runtime.EventToolCall, Content: ev.Part.Type}
					}
				}
				ch <- runtime.Event{Type: runtime.EventDone, Content: out}
				return
			}

			if isTransient(stderr, err) && attempt < d.maxRetries {
				lastErr = fmt.Errorf("attempt %d transient: %w; stderr: %s", attempt+1, err, stderr)
				ch <- runtime.Event{Type: runtime.EventError, Content: lastErr.Error(), Err: err}
				continue
			}

			ch <- runtime.Event{Type: runtime.EventError, Content: err.Error(), Err: err}
			return
		}

		ch <- runtime.Event{Type: runtime.EventError, Content: fmt.Sprintf("exhausted retries: %v", lastErr), Err: lastErr}
	}()

	return ch, nil
}

// runOnce executes a single opencode run call and returns (stdout, stderr, error).
func (d *Driver) runOnce(ctx context.Context, prompt string, timeoutSecs int) (string, string, error) {
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	args := []string{"run", "--format", "json", prompt}
	cmd := exec.CommandContext(taskCtx, d.binary, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return stdout.String(), filterStderr(stderr.String()), err
}

// filterStderr removes known noise lines from opencode stderr.
func filterStderr(s string) string {
	var kept []string
	scanner := bufio.NewScanner(bytes.NewReader([]byte(s)))
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		l := strings.TrimSpace(scanner.Text())
		if l == "" {
			continue
		}
		// Skip known noise patterns
		if strings.Contains(l, "downloading") ||
			strings.Contains(l, "plugin") ||
			strings.Contains(l, "warning:") {
			continue
		}
		kept = append(kept, l)
	}
	return strings.Join(kept, "\n")
}

// isTransient returns true for recoverable errors.
func isTransient(stderr string, err error) bool {
	s := stderr
	if err != nil {
		s += " " + err.Error()
	}
	if s == "" {
		return false
	}
	return strings.Contains(s, "503") ||
		strings.Contains(s, "Service temporarily unavailable") ||
		strings.Contains(s, "connection refused") ||
		strings.Contains(s, "deadline exceeded") ||
		strings.Contains(s, "ECONNREFUSED")
}

func (d *Driver) Capabilities() runtime.HostCapabilities {
	return runtime.HostCapabilities{SessionResume: false, ToolCalls: true, WebSearch: false, WriteHook: true}
}

func (d *Driver) ModelInfo() runtime.ModelInfo {
	return runtime.ModelInfo{ID: "opencode", Provider: "opencode", Context: 128000}
}

func (d *Driver) Pause(ctx context.Context, id runtime.SessionID) error { return nil }

func (d *Driver) Resume(ctx context.Context, id runtime.SessionID) error {
	return fmt.Errorf("opencode does not support session resume")
}

func (d *Driver) AuthStatus(ctx context.Context, id runtime.SessionID) (runtime.AuthStatus, error) {
	path, err := exec.LookPath(d.binary)
	if err != nil {
		return runtime.AuthStatus{Ready: false, Detail: fmt.Sprintf("opencode binary not found: %s", d.binary)}, nil
	}
	_ = path
	return runtime.AuthStatus{Ready: true, Detail: "opencode binary available"}, nil
}
