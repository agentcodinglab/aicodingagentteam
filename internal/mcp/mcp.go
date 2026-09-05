// Package mcp exposes governance and orchestration capabilities via MCP.
// It implements a stdio JSON-RPC 2.0 server conforming to the Model Context Protocol.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentcodinglab/aicodingagentteam/internal/governance"
)

// GovernanceResult holds the output of a govern_file scan.
type GovernanceResult struct {
	Path       string                 `json:"path"`
	Violations []governance.Violation `json:"violations"`
	Blocking   bool                   `json:"blocking"`
}

// Server exposes MCP tools to external MCP clients.
type Server struct {
	engine *governance.Engine
}

// New creates an MCP Server with the given governance engine.
func New(engine *governance.Engine) *Server {
	return &Server{engine: engine}
}

// GovernFile runs governance on a single file and returns violations.
func (s *Server) GovernFile(ctx context.Context, path string) (*GovernanceResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}
	violations := s.engine.Check(ctx, path, string(content))
	return &GovernanceResult{
		Path:       path,
		Violations: violations,
		Blocking:   governance.HasBlocking(violations),
	}, nil
}

// GovernDirectory runs governance on all code files in a directory tree.
func (s *Server) GovernDirectory(ctx context.Context, root string) ([]GovernanceResult, error) {
	var results []GovernanceResult
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".go" && ext != ".ts" && ext != ".tsx" && ext != ".js" &&
			ext != ".jsx" && ext != ".py" && ext != ".java" && ext != ".rb" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		violations := s.engine.Check(ctx, path, string(content))
		if len(violations) > 0 {
			results = append(results, GovernanceResult{
				Path:       path,
				Violations: violations,
				Blocking:   governance.HasBlocking(violations),
			})
		}
		return nil
	})
	return results, err
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

// serveReader reads JSON-RPC messages line-by-line from r and writes responses to w.
func (s *Server) serveReader(ctx context.Context, r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	// Allow larger lines (file content in params)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	encoder := json.NewEncoder(w)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
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
		// Only send response if ID is present (notification has no ID)
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
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]string{
					"name":    "aicodingagentteam-mcp",
					"version": "0.1.0",
				},
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
			},
		}
	case "tools/list":
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"tools": s.toolDefinitions(),
			},
		}
	case "tools/call":
		return s.handleToolCall(ctx, req)
	default:
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonRPCErr{Code: -32601, Message: "method not found: " + req.Method},
		}
	}
}

// toolCallParams holds the arguments for a tools/call request.
type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// governFileArgs are arguments for the govern_file tool.
type governFileArgs struct {
	Path string `json:"path"`
}

// governDirArgs are arguments for the govern_directory tool.
type governDirArgs struct {
	Root string `json:"root"`
}

func (s *Server) handleToolCall(ctx context.Context, req jsonRPCRequest) jsonRPCResponse {
	var params toolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonRPCErr{Code: -32602, Message: "invalid params"},
		}
	}

	switch params.Name {
	case "govern_file":
		var args governFileArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonRPCErr{Code: -32602, Message: "invalid arguments"}}
		}
		result, err := s.GovernFile(ctx, args.Path)
		if err != nil {
			return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonRPCErr{Code: -32603, Message: err.Error()}}
		}
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{"content": []map[string]string{{"type": "text", "text": fmt.Sprintf("%s: %d violations, blocking=%v", result.Path, len(result.Violations), result.Blocking)}}}}
	case "govern_directory":
		var args governDirArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonRPCErr{Code: -32602, Message: "invalid arguments"}}
		}
		results, err := s.GovernDirectory(ctx, args.Root)
		if err != nil {
			return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonRPCErr{Code: -32603, Message: err.Error()}}
		}
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{"content": []map[string]string{{"type": "text", "text": fmt.Sprintf("%s: %d files with violations", args.Root, len(results))}}}}
	default:
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonRPCErr{Code: -32602, Message: "unknown tool: " + params.Name}}
	}
}

// toolDefinitions returns the MCP tool definitions.
func (s *Server) toolDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "govern_file",
			"description": "Run governance checks on a single source file",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "file path to scan"},
				},
				"required": []string{"path"},
			},
		},
		{
			"name":        "govern_directory",
			"description": "Run governance checks on all code files in a directory tree",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"root": map[string]interface{}{"type": "string", "description": "directory root to scan"},
				},
				"required": []string{"root"},
			},
		},
	}
}

// Serve starts the MCP server over stdio JSON-RPC.
// Reads JSON-RPC requests from stdin, writes responses to stdout.
func (s *Server) Serve(ctx context.Context) error {
	return s.serveReader(ctx, os.Stdin, os.Stdout)
}

// ServeReader is a testable variant of Serve that reads from r and writes to w.
func (s *Server) ServeReader(ctx context.Context, r io.Reader, w io.Writer) error {
	return s.serveReader(ctx, r, w)
}
