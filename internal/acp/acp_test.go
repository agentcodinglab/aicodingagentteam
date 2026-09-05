package acp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestServeReader_Initialize(t *testing.T) {
	s := New()
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
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected result map, got %T", resp.Result)
	}
	if result["protocolVersion"] == nil {
		t.Error("expected protocolVersion in initialize response")
	}
}

func TestServeReader_SessionStart(t *testing.T) {
	s := New()
	params := sessionStartParams{AgentID: "codex", Prompt: "build app"}
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage("1"),
		Method:  "session/start",
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
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected result map, got %T", resp.Result)
	}
	if result["sessionId"] == nil {
		t.Error("expected sessionId in response")
	}
	if result["status"] != "active" {
		t.Errorf("expected status active, got %v", result["status"])
	}
}

func TestServeReader_SessionStop(t *testing.T) {
	s := New()
	startParams := sessionStartParams{AgentID: "codex"}
	startReq := jsonRPCRequest{
		JSONRPC: "2.0", ID: json.RawMessage("1"), Method: "session/start",
		Params: mustMarshal(startParams),
	}
	startData, _ := json.Marshal(startReq)
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(string(startData)+"\n"), &buf)

	stopParams := sessionStopParams{SessionID: "session-0"}
	stopReq := jsonRPCRequest{
		JSONRPC: "2.0", ID: json.RawMessage("2"), Method: "session/stop",
		Params: mustMarshal(stopParams),
	}
	stopData, _ := json.Marshal(stopReq)
	buf.Reset()
	_ = s.ServeReader(context.Background(), strings.NewReader(string(stopData)+"\n"), &buf)

	var stopResp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &stopResp); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, buf.String())
	}
	if stopResp.Error != nil {
		t.Errorf("unexpected error: %s", stopResp.Error.Message)
	}
	result, ok := stopResp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected result map, got %T", stopResp.Result)
	}
	if result["status"] != "stopped" {
		t.Errorf("expected status stopped, got %v", result["status"])
	}
}

func TestServeReader_SessionList(t *testing.T) {
	s := New()
	for i := 0; i < 2; i++ {
		startReq := jsonRPCRequest{
			JSONRPC: "2.0", ID: json.RawMessage("1"), Method: "session/start",
			Params: mustMarshal(sessionStartParams{AgentID: "codex"}),
		}
		data, _ := json.Marshal(startReq)
		_ = s.ServeReader(context.Background(), strings.NewReader(string(data)+"\n"), &bytes.Buffer{})
	}

	listReq := jsonRPCRequest{JSONRPC: "2.0", ID: json.RawMessage("3"), Method: "session/list"}
	data, _ := json.Marshal(listReq)
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
	sessions, ok := result["sessions"].([]interface{})
	if !ok {
		t.Fatalf("expected sessions array, got %T", result["sessions"])
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
}

func TestServeReader_UnknownMethod(t *testing.T) {
	s := New()
	req := jsonRPCRequest{JSONRPC: "2.0", ID: json.RawMessage("1"), Method: "unknown/method"}
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
	if resp.Error.Code != -32601 {
		t.Errorf("expected code -32601, got %d", resp.Error.Code)
	}
}

func TestServeReader_InvalidJSON(t *testing.T) {
	s := New()
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
	s := New()
	req := jsonRPCRequest{JSONRPC: "2.0", ID: json.RawMessage("null"), Method: "initialize"}
	data, _ := json.Marshal(req)
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(string(data)+"\n"), &buf)
}

func TestServeReader_SessionStartInvalidParams(t *testing.T) {
	s := New()
	// Pass valid outer JSON but truncated params to trigger json.Unmarshal error in handleSessionStart
	input := `{"jsonrpc":"2.0","id":1,"method":"session/start","params":{"missing":}}`
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(input+"\n"), &buf)

	var resp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, buf.String())
	}
	if resp.Error == nil {
		t.Error("expected error for invalid params")
	}
}

func TestServeReader_SessionStopInvalidParams(t *testing.T) {
	s := New()
	input := `{"jsonrpc":"2.0","id":1,"method":"session/stop","params":{"missing":}}`
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(input+"\n"), &buf)

	var resp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, buf.String())
	}
	if resp.Error == nil {
		t.Error("expected error for invalid params")
	}
}

func TestServeReader_EmptyLine(t *testing.T) {
	s := New()
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader("\n\n"), &buf)
}

func TestServeReader_SessionStopNonexistent(t *testing.T) {
	s := New()
	stopReq := jsonRPCRequest{
		JSONRPC: "2.0", ID: json.RawMessage("1"), Method: "session/stop",
		Params: mustMarshal(sessionStopParams{SessionID: "nonexistent"}),
	}
	data, _ := json.Marshal(stopReq)
	var buf bytes.Buffer
	_ = s.ServeReader(context.Background(), strings.NewReader(string(data)+"\n"), &buf)

	var resp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %s", resp.Error.Message)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected result map, got %T", resp.Result)
	}
	if result["status"] != "stopped" {
		t.Errorf("expected status stopped, got %v", result["status"])
	}
}

func TestServer_New(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}
	if s.sessions == nil {
		t.Error("sessions map should be initialized")
	}
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
