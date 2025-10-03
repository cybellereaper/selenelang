package lsp

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

func TestReplyIncludesNullResult(t *testing.T) {
	writer := &bytes.Buffer{}
	conn := newJSONRPCConnection(strings.NewReader(""), writer)

	if err := conn.Reply(json.RawMessage("1"), nil); err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}

	payload := writer.String()
	parts := strings.SplitN(payload, "\r\n\r\n", 2)
	if len(parts) != 2 {
		t.Fatalf("expected header and body, got %q", payload)
	}
	body := parts[1]
	if want := len(body); parts[0] != "Content-Length: "+strconv.Itoa(want) {
		t.Fatalf("unexpected header %q", parts[0])
	}

	var msg map[string]interface{}
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if _, ok := msg["error"]; ok {
		t.Fatalf("unexpected error field: %v", msg["error"])
	}
	result, ok := msg["result"]
	if !ok {
		t.Fatalf("result field missing: %v", msg)
	}
	if result != nil {
		t.Fatalf("expected result null, got %v", result)
	}
}

func TestReplyErrorOmitsResult(t *testing.T) {
	writer := &bytes.Buffer{}
	conn := newJSONRPCConnection(strings.NewReader(""), writer)

	if err := conn.ReplyError(json.RawMessage("1"), -32603, "boom"); err != nil {
		t.Fatalf("ReplyError returned error: %v", err)
	}

	payload := writer.String()
	parts := strings.SplitN(payload, "\r\n\r\n", 2)
	if len(parts) != 2 {
		t.Fatalf("expected header and body, got %q", payload)
	}
	body := parts[1]

	var msg map[string]interface{}
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if _, ok := msg["result"]; ok {
		t.Fatalf("unexpected result field: %v", msg["result"])
	}
	if _, ok := msg["error"]; !ok {
		t.Fatalf("expected error field: %v", msg)
	}
}
