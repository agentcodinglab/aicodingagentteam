package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentcodinglab/aicodingagentteam/internal/governance"
)

func TestGovernFile_CleanFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "clean.go")
	_ = os.WriteFile(p, []byte("package main\nfunc main() {}\n"), 0o644)

	s := New(governance.New())
	result, err := s.GovernFile(context.Background(), p)
	if err != nil {
		t.Fatalf("GovernFile failed: %v", err)
	}
	if result.Blocking {
		t.Error("clean file should not be blocking")
	}
}

func TestGovernFile_NonexistentFile(t *testing.T) {
	s := New(governance.New())
	_, err := s.GovernFile(context.Background(), "/nonexistent/file.go")
	if err == nil {
		t.Error("should fail on nonexistent file")
	}
}

func TestGovernFile_DetectsSecretLeak(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "leaky.go")
	_ = os.WriteFile(p, []byte("package main\nvar apiKey = \"sk-1234567890abcdef1234567890abcdef12345678\"\n"), 0o644)

	s := New(governance.New())
	result, err := s.GovernFile(context.Background(), p)
	if err != nil {
		t.Fatalf("GovernFile failed: %v", err)
	}
	if !result.Blocking {
		t.Error("secret leak should be blocking")
	}
	if len(result.Violations) == 0 {
		t.Error("should have violations")
	}
}

func TestGovernDirectory_CleanDir(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main\nfunc a() {}\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "b.go"), []byte("package main\nfunc b() {}\n"), 0o644)

	s := New(governance.New())
	results, err := s.GovernDirectory(context.Background(), dir)
	if err != nil {
		t.Fatalf("GovernDirectory failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("clean dir should have 0 results, got %d", len(results))
	}
}

func TestGovernDirectory_WithSubdirectory(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	_ = os.MkdirAll(sub, 0o755)
	_ = os.WriteFile(filepath.Join(sub, "file.go"), []byte("package main\nvar apiKey = \"sk-1234567890abcdef1234567890abcdef12345678\"\n"), 0o644)

	s := New(governance.New())
	results, err := s.GovernDirectory(context.Background(), dir)
	if err != nil {
		t.Fatalf("GovernDirectory failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected violations from subdirectory")
	}
}

func TestGovernDirectory_NonexistentRoot(t *testing.T) {
	s := New(governance.New())
	_, err := s.GovernDirectory(context.Background(), "/nonexistent/path/xyz")
	if err == nil {
		t.Error("should fail on nonexistent root")
	}
}

func TestServeReader_Initialize(t *testing.T) {
	s := New(governance.New())
	req := jsonRPCRequest{JSONRPC: "2.0", ID: json.RawMessage("1"), Method: "initialize"}
	data, _ := json.Marshal(req)
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(string(data)+"\n"), &buf)

	var resp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v\nraw: %s", err, buf.String())
	}
	if resp.JSONRPC != "2.0" {
		t.Error("expected jsonrpc 2.0")
	}
}

func TestServeReader_ToolsList(t *testing.T) {
	s := New(governance.New())
	req := jsonRPCRequest{JSONRPC: "2.0", ID: json.RawMessage("1"), Method: "tools/list"}
	data, _ := json.Marshal(req)
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(string(data)+"\n"), &buf)

	var resp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, buf.String())
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected result map, got %T", resp.Result)
	}
	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatal("expected tools array")
	}
	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}
	// Verify each tool has name, description, inputSchema
	for i, tool := range tools {
		toolMap, ok := tool.(map[string]interface{})
		if !ok {
			t.Fatalf("tool %d: expected map", i)
		}
		if toolMap["name"] == nil {
			t.Errorf("tool %d: missing name", i)
		}
		if toolMap["description"] == nil {
			t.Errorf("tool %d: missing description", i)
		}
		if toolMap["inputSchema"] == nil {
			t.Errorf("tool %d: missing inputSchema", i)
		}
	}
}

func TestServeReader_GovernFileTool(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "clean.go")
	_ = os.WriteFile(p, []byte("package main\nfunc main() {}\n"), 0o644)

	s := New(governance.New())
	params := toolCallParams{Name: "govern_file"}
	params.Arguments, _ = json.Marshal(governFileArgs{Path: p})
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage("1"),
		Method:  "tools/call",
		Params:  mustMarshal(params),
	}
	data, _ := json.Marshal(req)
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(string(data)+"\n"), &buf)

	var resp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, buf.String())
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %s", resp.Error.Message)
	}
}

func TestServeReader_GovernFileTool_NonexistentFile(t *testing.T) {
	s := New(governance.New())
	params := toolCallParams{Name: "govern_file"}
	params.Arguments, _ = json.Marshal(governFileArgs{Path: "/nonexistent/file.go"})
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage("1"),
		Method:  "tools/call",
		Params:  mustMarshal(params),
	}
	data, _ := json.Marshal(req)
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(string(data)+"\n"), &buf)

	var resp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestServeReader_GovernFileTool_InvalidArgs(t *testing.T) {
	s := New(governance.New())
	params := toolCallParams{Name: "govern_file"}
	params.Arguments = json.RawMessage("invalid")
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage("1"),
		Method:  "tools/call",
		Params:  mustMarshal(params),
	}
	data, _ := json.Marshal(req)
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(string(data)+"\n"), &buf)

	var resp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error for invalid args")
	}
}

func TestServeReader_GovernDirectoryTool(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644)

	s := New(governance.New())
	params := toolCallParams{Name: "govern_directory"}
	params.Arguments, _ = json.Marshal(governDirArgs{Root: dir})
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage("1"),
		Method:  "tools/call",
		Params:  mustMarshal(params),
	}
	data, _ := json.Marshal(req)
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(string(data)+"\n"), &buf)

	var resp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, buf.String())
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %s", resp.Error.Message)
	}
}

func TestServeReader_GovernDirectoryTool_InvalidArgs(t *testing.T) {
	s := New(governance.New())
	params := toolCallParams{Name: "govern_directory"}
	params.Arguments = json.RawMessage("invalid")
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage("1"),
		Method:  "tools/call",
		Params:  mustMarshal(params),
	}
	data, _ := json.Marshal(req)
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(string(data)+"\n"), &buf)

	var resp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error for invalid args")
	}
}

func TestServeReader_UnknownMethod(t *testing.T) {
	s := New(governance.New())
	req := jsonRPCRequest{JSONRPC: "2.0", ID: json.RawMessage("1"), Method: "unknown_method"}
	data, _ := json.Marshal(req)
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(string(data)+"\n"), &buf)

	var resp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error for unknown method")
	}
}

func TestServeReader_InvalidJSON(t *testing.T) {
	s := New(governance.New())
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader("not-valid-json\n"), &buf)

	var resp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != -32700 {
		t.Error("expected parse error")
	}
}

func TestServeReader_Notification(t *testing.T) {
	s := New(governance.New())
	req := jsonRPCRequest{JSONRPC: "2.0", ID: json.RawMessage("null"), Method: "initialize"}
	data, _ := json.Marshal(req)
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(string(data)+"\n"), &buf)
	_ = buf.Len() // notifications should not crash
}

func TestServeReader_UnknownTool(t *testing.T) {
	s := New(governance.New())
	params := toolCallParams{Name: "nonexistent_tool"}
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage("1"),
		Method:  "tools/call",
		Params:  mustMarshal(params),
	}
	data, _ := json.Marshal(req)
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(string(data)+"\n"), &buf)

	var resp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error for unknown tool")
	}
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
