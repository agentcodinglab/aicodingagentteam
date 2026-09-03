// Package codex implements the Codex CLI host driver.
// It calls `codex exec` via exec.Command in non-interactive mode.
// See ADR-0007 for feasibility verification.
package codex

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/agentcodinglab/aicodingagentteam/pkg/runtime"
)

// Driver is the Codex CLI driver (codex exec non-interactive mode).
type Driver struct {
	binary     string
	timeout    int
	maxRetries int
}

// Option configures a Driver.
type Option func(*Driver)

// WithBinary overrides the codex binary path.
func WithBinary(path string) Option { return func(d *Driver) { d.binary = path } }

// WithTimeout sets the per-task timeout in seconds.
func WithTimeout(secs int) Option { return func(d *Driver) { d.timeout = secs } }

// WithMaxRetries sets the max retry count for transient failures.
func WithMaxRetries(n int) Option { return func(d *Driver) { d.maxRetries = n } }

// New creates a Codex driver. Defaults: binary="codex", timeout=300, retries=3.
func New(opts ...Option) *Driver {
	d := &Driver{binary: "codex", timeout: 300, maxRetries: 3}
	for _, o := range opts {
		o(d)
	}
	return d
}

func (d *Driver) StartSession(ctx context.Context, opts runtime.SessionOpts) (runtime.SessionID, error) {
	return runtime.SessionID("codex"), nil
}

func (d *Driver) DestroySession(ctx context.Context, id runtime.SessionID) error { return nil }

// SendTask runs `codex exec --skip-git-repo-check <prompt>` and streams events.
// On transient errors (503, stream disconnect) it retries up to maxRetries times.
func (d *Driver) SendTask(ctx context.Context, id runtime.SessionID, task runtime.TaskPayload) (<-chan runtime.Event, error) {
	ch := make(chan runtime.Event, 4)

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

			ch <- runtime.Event{Type: runtime.EventStart, Content: fmt.Sprintf("codex attempt %d", attempt+1)}

			out, stderr, err := d.runOnce(ctx, prompt, timeout)

			if err == nil {
				ch <- runtime.Event{Type: runtime.EventMessage, Content: out}
				ch <- runtime.Event{Type: runtime.EventDone, Content: out}
				return
			}

			if isTransient(stderr, err) && attempt < d.maxRetries {
				lastErr = fmt.Errorf("attempt %d transient: %w; stderr: %s", attempt+1, err, stderr)
				ch <- runtime.Event{Type: runtime.EventError, Content: lastErr.Error()}
				continue
			}

			ch <- runtime.Event{Type: runtime.EventError, Content: err.Error(), Err: err}
			return
		}

		ch <- runtime.Event{Type: runtime.EventError, Content: fmt.Sprintf("exhausted retries: %v", lastErr), Err: lastErr}
	}()

	return ch, nil
}

// runOnce executes a single codex exec call and returns (stdout, stderr, error).
func (d *Driver) runOnce(ctx context.Context, prompt string, timeoutSecs int) (string, string, error) {
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	args := []string{"exec", "--skip-git-repo-check", prompt}
	cmd := exec.CommandContext(taskCtx, d.binary, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Filter stderr noise: keep only meaningful lines
	filteredStderr := filterStderr(stderr.String())

	return stdout.String(), filteredStderr, err
}

// filterStderr removes known noise lines (plugin sync warnings, model metadata warnings).
func filterStderr(s string) string {
	var kept []string
	for _, line := range strings.Split(s, "\n") {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		// Skip known noise patterns
		if strings.Contains(l, "remote installed plugin bundle sync failed") ||
			strings.Contains(l, "Unknown model") ||
			strings.Contains(l, "Shell snapshot not supported") ||
			strings.Contains(l, "failed to warm featured plugin") ||
			strings.Contains(l, "ignoring interface.defaultPrompt") ||
			strings.Contains(l, "Reading additional input from stdin") {
			continue
		}
		kept = append(kept, l)
	}
	return strings.Join(kept, "\n")
}

// isTransient returns true for recoverable errors (503, stream disconnect, timeout).
// Checks stderr and error text together; either may carry the signal.
func isTransient(stderr string, err error) bool {
	s := stderr
	if err != nil {
		s += " " + err.Error()
	}
	if s == "" {
		return false
	}
	return strings.Contains(s, "503") ||
		strings.Contains(s, "stream disconnected") ||
		strings.Contains(s, "Service temporarily unavailable") ||
		strings.Contains(s, "Reconnecting") ||
		strings.Contains(s, "deadline exceeded")
}

func (d *Driver) Capabilities() runtime.HostCapabilities {
	return runtime.HostCapabilities{SessionResume: true, ToolCalls: true, WebSearch: true, WriteHook: false}
}

func (d *Driver) ModelInfo() runtime.ModelInfo {
	return runtime.ModelInfo{ID: "codex", Provider: "openai", Context: 200000}
}

func (d *Driver) Pause(ctx context.Context, id runtime.SessionID) error  { return nil }
func (d *Driver) Resume(ctx context.Context, id runtime.SessionID) error { return nil }

func (d *Driver) AuthStatus(ctx context.Context, id runtime.SessionID) (runtime.AuthStatus, error) {
	// Check if codex binary is available and authenticated
	path, err := exec.LookPath(d.binary)
	if err != nil {
		return runtime.AuthStatus{Ready: false, Detail: fmt.Sprintf("codex binary not found: %s", d.binary)}, nil
	}
	_ = path
	return runtime.AuthStatus{Ready: true, Detail: "codex binary available"}, nil
}
