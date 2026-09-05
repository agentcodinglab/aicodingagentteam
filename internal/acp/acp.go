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
)

// Session represents an ACP agent session.
type Session struct {
	ID     string `json:"id"`
	Status string `json:"status"` // active / paused / stopped
}

// Server handles ACP session lifecycle for standard agent clients.
type Server struct {
	mu       sync.Mutex
	sessions map[string]*Session
}

// New creates an ACP Server.
func New() *Server {
	return &Server{sessions: make(map[string]*Session)}
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

// Serve starts the ACP server over stdio JSON-RPC.
func (s *Server) Serve(ctx context.Context) error {
	return s.serveReader(ctx, os.Stdin, os.Stdout)
}

// ServeReader is a testable variant of Serve that reads from r and writes to w.
func (s *Server) ServeReader(ctx context.Context, r io.Reader, w io.Writer) error {
	return s.serveReader(ctx, r, w)
}
