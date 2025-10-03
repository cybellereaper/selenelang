package lsp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

const jsonrpcVersion = "2.0"

var errClientExit = errors.New("client requested exit")

type requestMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type responseMessage struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      json.RawMessage  `json:"id,omitempty"`
	Result  *json.RawMessage `json:"result,omitempty"`
	Error   *responseError   `json:"error,omitempty"`
}

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type jsonRPCConnection struct {
	reader  *bufio.Reader
	writer  *bufio.Writer
	writeMu sync.Mutex
}

func newJSONRPCConnection(r io.Reader, w io.Writer) *jsonRPCConnection {
	return &jsonRPCConnection{
		reader: bufio.NewReader(r),
		writer: bufio.NewWriter(w),
	}
}

func (c *jsonRPCConnection) Read() (requestMessage, error) {
	payload, err := c.readPayload()
	if err != nil {
		return requestMessage{}, err
	}
	var msg requestMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return requestMessage{}, err
	}
	if msg.JSONRPC != "" && msg.JSONRPC != jsonrpcVersion {
		return requestMessage{}, fmt.Errorf("unsupported jsonrpc version: %s", msg.JSONRPC)
	}
	return msg, nil
}

func (c *jsonRPCConnection) Reply(id json.RawMessage, result interface{}) error {
	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}
	raw := json.RawMessage(payload)
	resp := responseMessage{JSONRPC: jsonrpcVersion, ID: id, Result: &raw}
	return c.writeMessage(resp)
}

func (c *jsonRPCConnection) ReplyError(id json.RawMessage, code int, message string) error {
	resp := responseMessage{JSONRPC: jsonrpcVersion, ID: id, Error: &responseError{Code: code, Message: message}}
	return c.writeMessage(resp)
}

func (c *jsonRPCConnection) Notify(method string, params interface{}) error {
	msg := map[string]interface{}{
		"jsonrpc": jsonrpcVersion,
		"method":  method,
		"params":  params,
	}
	return c.writeMessage(msg)
}

func (c *jsonRPCConnection) readPayload() ([]byte, error) {
	headers := make(map[string]string)
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if line == "\r\n" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(strings.TrimSuffix(parts[1], "\r\n"))
		headers[strings.ToLower(key)] = value
	}
	lengthStr, ok := headers[strings.ToLower("Content-Length")]
	if !ok {
		return nil, fmt.Errorf("missing Content-Length header")
	}
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return nil, err
	}
	if length == 0 {
		return []byte("{}"), nil
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(c.reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *jsonRPCConnection) writeMessage(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if _, err := c.writer.WriteString(header); err != nil {
		return err
	}
	if _, err := c.writer.Write(data); err != nil {
		return err
	}
	return c.writer.Flush()
}
