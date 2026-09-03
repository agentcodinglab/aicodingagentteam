// Package audit writes structured JSONL audit logs for all key operations.
package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger writes audit entries as JSONL to .aicodingagentteam/audit/.
type Logger struct {
	mu  sync.Mutex
	dir string
}

// Entry is a single audit log entry.
type Entry struct {
	TS         time.Time `json:"ts"`
	Type       string    `json:"type"` // tool_call / a2a_message / verify
	Agent      string    `json:"agent,omitempty"`
	Task       string    `json:"task,omitempty"`
	Tool       string    `json:"tool,omitempty"`
	Result     string    `json:"result"`
	Detail     string    `json:"detail,omitempty"`
	DurationMs int64     `json:"duration_ms,omitempty"`
}

// New creates an audit Logger writing to the given directory.
func New(dir string) *Logger { return &Logger{dir: dir} }

// Log appends a JSONL entry.
func (l *Logger) Log(e Entry) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e.TS.IsZero() {
		e.TS = time.Now()
	}
	if err := os.MkdirAll(l.dir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(l.dir, "events.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	return enc.Encode(e)
}
