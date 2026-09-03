// Package runtime defines the host driver abstraction interface.
// Each AI coding CLI driver implements Runtime to be schedulable by the Coordinator.
package runtime

import "context"

// SessionID uniquely identifies a host session.
type SessionID string

// SessionOpts configures a new host session.
type SessionOpts struct {
	Workspace   string // shared artifact volume mount path
	MaxTokens   int
	WorkingDir  string
	Permissions []string // allowed tool permissions
}

// TaskPayload is the task sent to a host CLI.
type TaskPayload struct {
	Instruction string   // natural language task instruction
	Files       []string // context files to read
	WritePaths  []string // allowed write scope
	Timeout     int      // seconds; 0 = use default
}

// EventType enumerates host event types.
type EventType string

const (
	EventStart    EventType = "start"
	EventMessage  EventType = "message"
	EventToolCall EventType = "tool_call"
	EventDone     EventType = "done"
	EventError    EventType = "error"
)

// Event is a single streamed event from the host CLI.
type Event struct {
	Type    EventType
	Content string
	Tool    string // tool name for EventToolCall
	Err     error  // non-nil for EventError
}

// HostCapabilities declares what a host can do; missing capabilities must be reported, never faked.
type HostCapabilities struct {
	SessionResume bool // can resume a paused session
	ToolCalls     bool // emits tool call events
	WebSearch     bool // can search the internet
	WriteHook     bool // supports pre-write hooks
}

// ModelInfo describes the underlying model of a host CLI.
type ModelInfo struct {
	ID       string
	Provider string
	Context  int // max context window
}

// AuthStatus reports whether the host CLI is authenticated.
type AuthStatus struct {
	Ready  bool
	Detail string
}

// Runtime is the host driver trait. Implementations must NOT hold API keys;
// authentication is delegated to the underlying CLI.
type Runtime interface {
	StartSession(ctx context.Context, opts SessionOpts) (SessionID, error)
	DestroySession(ctx context.Context, id SessionID) error
	SendTask(ctx context.Context, id SessionID, task TaskPayload) (<-chan Event, error)
	Capabilities() HostCapabilities
	ModelInfo() ModelInfo
	Pause(ctx context.Context, id SessionID) error
	Resume(ctx context.Context, id SessionID) error
	AuthStatus(ctx context.Context, id SessionID) (AuthStatus, error)
}

// Backend identifies a host driver.
type Backend string

const (
	BackendClaudeCode Backend = "claude-code"
	BackendCodex      Backend = "codex"
	BackendOpenCode   Backend = "opencode"
	BackendDSH        Backend = "deepseek-dsh"
)
