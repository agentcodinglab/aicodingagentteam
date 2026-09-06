// Package acp implements the Agent Client Protocol server.
// It exposes a stdio JSON-RPC 2.0 server for session lifecycle management.
package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"

	`github.com/agentcodinglab/aicodingagentteam/internal/types`
)

// Session represents an ACP agent session.
type Session struct {
	ID     string `json:"id"`
	Status string `json:"status"` // active / paused / stopped
	mu     sync.Mutex
	tasks  map[string]*Task // active task IDs for this session
}

// Task is a single newTask execution within a session (ADR-0020).
type Task struct {
	ID        string `json:"taskId"`
	SessionID string `json:"sessionId"`
	Status    string `json:"status"` // pending / running / done / error
	Events    []TaskEvent `json:"events"`
}

// TaskEvent is a single streamed event for a newTask.
type TaskEvent struct {
	Type    string `json:"type"` // start / message / tool_call / done / error
	Content string `json:"content"`
}

// DirectorLike is the minimal contract the ACP server needs from the coordinator.
// It lets us dispatch user requests without an import cycle.
type DirectorLike interface {
	Handle(ctx context.Context, req types.UserRequest) (*types.Delivery, error)
}

// Server handles ACP session lifecycle for standard agent clients.
type Server struct {
	mu          sync.Mutex
	sessions    map[string]*Session
	director    DirectorLike // optional: when set, session/newTask dispatches to it
	taskCounter atomic.Int64
	notifier    func(method string, params interface{}) // optional: write notifications to client (e.g. notifications/session/update)
}

// New creates an ACP Server without a director (lifecycle-only).
func New() *Server {
	return &Server{sessions: make(map[string]*Session)}
}

// NewWithDirector creates an ACP Server backed by a DirectorLike (ADR-0020).
// notifier is invoked for each streamed event; it must be safe to call from any goroutine.
func NewWithDirector(d DirectorLike, notifier func(method string, params interface{})) *Server {
	return &Server{
		sessions: make(map[string]*Session),
		director: d,
		notifier: notifier,
	}
}

// jsonRPCRequest is a JSON-RPC 2.0 request envelope.
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// jsonRPCResponse is a JSON-RPC 2.0 response envelope.
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *jsonRPCErr     `json:"error,omitempty"`
}

type jsonRPCErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// sessionStartParams holds arguments for session/start.
type sessionStartParams struct {
	AgentID string `json:"agentId"`
	Prompt  string `json:"prompt,omitempty"`
}

// sessionStopParams holds arguments for session/stop.
type sessionStopParams struct {
	SessionID string `json:"sessionId"`
}

// serveReader reads JSON-RPC messages line-by-line from r and writes responses to w.
func (s *Server) serveReader(ctx context.Context, r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	encoder := json.NewEncoder(w)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			_ = encoder.Encode(jsonRPCResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   &jsonRPCErr{Code: -32700, Message: "parse error"},
			})
			continue
		}

		resp := s.handleMethod(ctx, req)
		if len(req.ID) > 0 && string(req.ID) != "null" {
			_ = encoder.Encode(resp)
		}
	}
	return scanner.Err()
}

// handleMethod dispatches a JSON-RPC method to the appropriate handler.
func (s *Server) handleMethod(ctx context.Context, req jsonRPCRequest) jsonRPCResponse {
	switch req.Method {
	case "initialize":
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2025-03-26",
				"serverInfo": map[string]string{
					"name":    "aicodingagentteam-acp",
					"version": "0.1.0",
				},
				"capabilities": map[string]interface{}{
					"sessions": map[string]interface{}{},
				},
			},
		}
	case "session/start":
		return s.handleSessionStart(ctx, req)
	case "session/stop":
		return s.handleSessionStop(ctx, req)
	case "session/list":
		return s.handleSessionList(ctx, req)
	case "session/newTask":
		return s.handleSessionNewTask(ctx, req)
	default:
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonRPCErr{Code: -32601, Message: "method not found: " + req.Method},
		}
	}
}

func (s *Server) handleSessionStart(ctx context.Context, req jsonRPCRequest) jsonRPCResponse {
	var params sessionStartParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonRPCErr{Code: -32602, Message: "invalid params"}}
	}
	sess := &Session{
		ID:     fmt.Sprintf("session-%d", len(s.sessions)),
		Status: "active",
	}
	s.mu.Lock()
	s.sessions[sess.ID] = sess
	s.mu.Unlock()
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"sessionId": sess.ID,
			"status":    sess.Status,
		},
	}
}

func (s *Server) handleSessionStop(ctx context.Context, req jsonRPCRequest) jsonRPCResponse {
	var params sessionStopParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonRPCErr{Code: -32602, Message: "invalid params"}}
	}
	s.mu.Lock()
	if sess, ok := s.sessions[params.SessionID]; ok {
		sess.Status = "stopped"
	}
	s.mu.Unlock()
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"sessionId": params.SessionID,
			"status":    "stopped",
		},
	}
}

func (s *Server) handleSessionList(ctx context.Context, req jsonRPCRequest) jsonRPCResponse {
	s.mu.Lock()
	defer s.mu.Unlock()
	sessions := make([]map[string]string, 0, len(s.sessions))
	for id, sess := range s.sessions {
		sessions = append(sessions, map[string]string{
			"sessionId": id,
			"status":    sess.Status,
		})
	}
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"sessions": sessions,
		},
	}
}


// sessionNewTaskParams holds arguments for session/newTask (ADR-0020).
type sessionNewTaskParams struct {
	SessionID string `json:"sessionId"`
	AgentID   string `json:"agentId"`
	Prompt    string `json:"prompt"`
}

// sessionUpdateParams is the payload of notifications/session/update events.
type sessionUpdateParams struct {
	SessionID string    `json:"sessionId"`
	TaskID    string    `json:"taskId"`
	Event     TaskEvent `json:"event"`
}

// handleSessionNewTask creates a new task in a session and dispatches the
// request to the wired Director in a goroutine. Events are streamed via the
// notifier (if set) as JSON-RPC notifications/session/update.
func (s *Server) handleSessionNewTask(ctx context.Context, req jsonRPCRequest) jsonRPCResponse {
	var params sessionNewTaskParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonRPCErr{Code: -32602, Message: "invalid params"}}
	}
	if s.director == nil {
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonRPCErr{Code: -32000, Message: "session/newTask requires Director"}}
	}
	s.mu.Lock()
	sess, ok := s.sessions[params.SessionID]
	if !ok {
		s.mu.Unlock()
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonRPCErr{Code: -32001, Message: "session not found: " + params.SessionID}}
	}
	taskID := fmt.Sprintf("task-%d", s.taskCounter.Add(1))
	task := &Task{ID: taskID, SessionID: params.SessionID, Status: "pending"}
	if sess.tasks == nil {
		sess.tasks = make(map[string]*Task)
	}
	sess.tasks[taskID] = task
	s.mu.Unlock()

	go s.dispatchTaskAsync(ctx, sess, task, params)

	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"taskId": taskID,
			"status": "pending",
		},
	}
}

// dispatchTaskAsync runs the task and emits streamed events.
func (s *Server) dispatchTaskAsync(ctx context.Context, sess *Session, task *Task, params sessionNewTaskParams) {
	s.emitUpdate(sess, task, "start", params.Prompt)

	delivery, err := s.director.Handle(ctx, types.UserRequest{
		Message: params.Prompt,
		Backend: params.AgentID,
	})
	if err != nil {
		task.Status = "error"
		s.emitUpdate(sess, task, "error", err.Error())
		return
	}

	s.emitUpdate(sess, task, "message", fmt.Sprintf("planID=%s score=%d passed=%v", delivery.PlanID, delivery.Score, delivery.Passed))
	for _, a := range delivery.Artifacts {
		s.emitUpdate(sess, task, "tool_call", a)
	}
	if delivery.Passed {
		task.Status = "done"
		s.emitUpdate(sess, task, "done", "ok")
	} else {
		task.Status = "error"
		s.emitUpdate(sess, task, "error", "delivery did not pass")
	}
}

// emitUpdate records the event on the task and (if wired) pushes a JSON-RPC
// notifications/session/update to the client.
func (s *Server) emitUpdate(sess *Session, task *Task, eventType, content string) {
	ev := TaskEvent{Type: eventType, Content: content}
	sess.mu.Lock()
	task.Events = append(task.Events, ev)
	sess.mu.Unlock()
	if s.notifier != nil {
		s.notifier("notifications/session/update", sessionUpdateParams{
			SessionID: sess.ID,
			TaskID:    task.ID,
			Event:     ev,
		})
	}
}


// Serve starts the ACP server over stdio JSON-RPC.
func (s *Server) Serve(ctx context.Context) error {
	return s.serveReader(ctx, os.Stdin, os.Stdout)
}

// ServeReader is a testable variant of Serve that reads from r and writes to w.
func (s *Server) ServeReader(ctx context.Context, r io.Reader, w io.Writer) error {
	return s.serveReader(ctx, r, w)
}
