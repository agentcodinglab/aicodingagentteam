// Package opencode implements the OpenCode CLI host driver.
// It drives `opencode acp` (stdio JSON-RPC, Agent Client Protocol 2025-03-26).
// See ADR-0007 + ADR-0021 (B) for protocol and feasibility.
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

// Driver is the OpenCode CLI driver (opencode acp stdio mode).
type Driver struct {
	binary     string
	timeout    int
	maxRetries int
}

type Option func(*Driver)
func WithBinary(path string) Option { return func(d *Driver) { d.binary = path } }
func WithTimeout(secs int) Option { return func(d *Driver) { d.timeout = secs } }
func WithMaxRetries(n int) Option { return func(d *Driver) { d.maxRetries = n } }
func New(opts ...Option) *Driver {
	d := &Driver{binary: "opencode", timeout: 300, maxRetries: 3}
	for _, o := range opts { o(d) }
	return d
}

func (d *Driver) StartSession(ctx context.Context, opts runtime.SessionOpts) (runtime.SessionID, error) {
	return runtime.SessionID("opencode"), nil
}
func (d *Driver) DestroySession(ctx context.Context, id runtime.SessionID) error { return nil }

// SendTask drives opencode acp (ADR-0021, B) and streams notifications/session/update
// as runtime.Event (ADR-0019). On transient errors it retries up to maxRetries times.
func (d *Driver) SendTask(ctx context.Context, id runtime.SessionID, task runtime.TaskPayload) (<-chan runtime.Event, error) {
	ch := make(chan runtime.Event, 8)
	go func() {
		defer close(ch)
		timeout := d.timeout
		if task.Timeout > 0 { timeout = task.Timeout }
		var lastErr error
		for attempt := 0; attempt <= d.maxRetries; attempt++ {
			if attempt > 0 {
				backoff := time.Duration(attempt*attempt) * 500 * time.Millisecond
				select { case <-ctx.Done(): return; case <-time.After(backoff): }
			}
			ch <- runtime.Event{Type: runtime.EventStart, Content: fmt.Sprintf("opencode attempt %d", attempt+1)}
			out, stderr, err := d.runACPSession(ctx, task.Instruction, timeout, ch)
			if err == nil {
				ch <- runtime.Event{Type: runtime.EventDone, Content: out}
				return
			}
			if isTransient(stderr, err) && attempt < d.maxRetries {
				lastErr = fmt.Errorf("attempt %d: %w; stderr: %s", attempt+1, err, stderr)
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

// runACPSession starts `opencode acp` as a stdio JSON-RPC server (ADR-0021, B).
// It sends initialize, session/new, then session/prompt and reads the response.
// Streamed notifications/session/update are yielded as runtime.Event.
func (d *Driver) runACPSession(ctx context.Context, prompt string, timeoutSecs int, ch chan<- runtime.Event) (string, string, error) {
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	args := []string{"acp", "--pure"}
	cmd := exec.CommandContext(taskCtx, d.binary, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil { return "", filterStderr(stderr.String()), err }
	stdout, err := cmd.StdoutPipe()
	if err != nil { return "", filterStderr(stderr.String()), err }
	if err := cmd.Start(); err != nil { return "", filterStderr(stderr.String()), err }
	defer func() { _ = cmd.Process.Kill() }()

	type rpcReq struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      int             `json:"id"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params,omitempty"`
	}
	type rpcResp struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Result  json.RawMessage `json:"result,omitempty"`
		Error   *struct{
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}
	type updateParams struct {
		SessionID string          `json:"sessionId"`
		Update    json.RawMessage `json:"update"`
	}
	type updateMsg struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	}

	enc := json.NewEncoder(stdin)
	write := func(id int, method string, params interface{}) error {
		var raw json.RawMessage
		if params != nil {
			b, err := json.Marshal(params)
			if err != nil { return err }
			raw = b
		}
		return enc.Encode(rpcReq{JSONRPC: "2.0", ID: id, Method: method, Params: raw})
	}

	if err := write(1, "initialize", map[string]any{"protocolVersion": "2025-03-26", "clientInfo": map[string]string{"name": "aicodingagentteam", "version": "0.8.0"}}); err != nil {
		return "", filterStderr(stderr.String()), err
	}
	if err := write(2, "session/new", map[string]any{"cwd": ".", "mcpServers": []any{}}); err != nil {
		return "", filterStderr(stderr.String()), err
	}
	if err := write(3, "session/prompt", map[string]any{"sessionId": "pending", "prompt": []map[string]string{{"type": "text", "text": prompt}}}); err != nil {
		return "", filterStderr(stderr.String()), err
	}
	_ = stdin.Close()

	var agg strings.Builder
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" { continue }
		agg.WriteString(raw)
		agg.WriteByte('\n')
		var resp rpcResp
		if err := json.Unmarshal([]byte(raw), &resp); err != nil { continue }
		if len(resp.ID) > 0 && string(resp.ID) == "3" {
			if resp.Error != nil { return agg.String(), filterStderr(stderr.String()), fmt.Errorf("acp error: %s", resp.Error.Message) }
			break
		}
		var note struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal([]byte(raw), &note); err != nil { continue }
		if note.Method == "notifications/session/update" {
			var su updateParams
			if err := json.Unmarshal(note.Params, &su); err == nil {
				var u updateMsg
				if err := json.Unmarshal(su.Update, &u); err == nil {
					switch u.Type {
					case "agent_message_chunk", "text", "message_chunk":
						if u.Content != "" { ch <- runtime.Event{Type: runtime.EventMessage, Content: u.Content} }
					case "tool_call", "tool_use":
						ch <- runtime.Event{Type: runtime.EventToolCall, Content: u.Type}
					case "done", "session_complete":
						// handled at response boundary
					}
				}
			}
		}
	}
	werr := cmd.Wait()
	if werr != nil { return agg.String(), filterStderr(stderr.String()), werr }
	return agg.String(), filterStderr(stderr.String()), scanner.Err()
}

func filterStderr(s string) string {
	var kept []string
	scanner := bufio.NewScanner(bytes.NewReader([]byte(s)))
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		l := strings.TrimSpace(scanner.Text())
		if l == "" { continue }
		if strings.Contains(l, "downloading") || strings.Contains(l, "plugin") || strings.Contains(l, "warning:") { continue }
		kept = append(kept, l)
	}
	return strings.Join(kept, "\n")
}

func isTransient(stderr string, err error) bool {
	s := stderr
	if err != nil { s += " " + err.Error() }
	if s == "" { return false }
	return strings.Contains(s, "503") || strings.Contains(s, "Service temporarily unavailable") || strings.Contains(s, "connection refused") || strings.Contains(s, "deadline exceeded") || strings.Contains(s, "ECONNREFUSED")
}

func (d *Driver) Capabilities() runtime.HostCapabilities {
	return runtime.HostCapabilities{SessionResume: true, ToolCalls: true, WebSearch: false, WriteHook: true}
}
func (d *Driver) ModelInfo() runtime.ModelInfo {
	return runtime.ModelInfo{ID: "opencode", Provider: "opencode", Context: 128000}
}
func (d *Driver) Pause(ctx context.Context, id runtime.SessionID) error { return nil }
func (d *Driver) Resume(ctx context.Context, id runtime.SessionID) error { return nil }
func (d *Driver) AuthStatus(ctx context.Context, id runtime.SessionID) (runtime.AuthStatus, error) {
	path, err := exec.LookPath(d.binary)
	if err != nil { return runtime.AuthStatus{Ready: false, Detail: fmt.Sprintf("opencode binary not found: %s", d.binary)}, nil }
	_ = path
	return runtime.AuthStatus{Ready: true, Detail: "opencode binary available"}, nil
}