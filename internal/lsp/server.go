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

	"selenelang/internal/format"
)

type Server struct {
	reader *bufio.Reader
	writer *bufio.Writer

	mu        sync.Mutex
	documents map[string]*document

	shuttingDown bool
}

type document struct {
	Text    string
	Version int
}

type requestMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type responseMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *responseError  `json:"error,omitempty"`
}

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const (
	jsonrpcVersion         = "2.0"
	methodInitialize       = "initialize"
	methodInitialized      = "initialized"
	methodShutdown         = "shutdown"
	methodExit             = "exit"
	methodDidOpen          = "textDocument/didOpen"
	methodDidChange        = "textDocument/didChange"
	methodDidClose         = "textDocument/didClose"
	methodDidSave          = "textDocument/didSave"
	methodCompletion       = "textDocument/completion"
	methodDidChangeConfig  = "workspace/didChangeConfiguration"
	methodDidChangeWatched = "workspace/didChangeWatchedFiles"
	methodDocumentSymbol   = "textDocument/documentSymbol"
	methodWorkspaceSymbol  = "workspace/symbol"
	methodDocumentFormat   = "textDocument/formatting"
)

var errClientExit = errors.New("client requested exit")

// NewServer constructs a Language Server Protocol server that communicates over the provided streams.
func NewServer(r io.Reader, w io.Writer) *Server {
	return &Server{
		reader:    bufio.NewReader(r),
		writer:    bufio.NewWriter(w),
		documents: make(map[string]*document),
	}
}

// Run processes incoming JSON-RPC requests until the client disconnects.
func (s *Server) Run() error {
	for {
		payload, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			if errors.Is(err, errClientExit) {
				return nil
			}
			return err
		}

		var msg requestMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			// Invalid messages are ignored per LSP robustness guidelines.
			continue
		}

		if msg.Method == "" {
			continue
		}

		if err := s.dispatch(msg); err != nil {
			if errors.Is(err, errClientExit) {
				return nil
			}
			return err
		}
	}
}

func (s *Server) dispatch(msg requestMessage) error {
	switch msg.Method {
	case methodInitialize:
		return s.handleInitialize(msg)
	case methodInitialized:
		return nil
	case methodShutdown:
		return s.handleShutdown(msg)
	case methodExit:
		return errClientExit
	case methodDidOpen:
		s.handleDidOpen(msg)
	case methodDidChange:
		s.handleDidChange(msg)
	case methodDidClose:
		s.handleDidClose(msg)
	case methodDidSave:
		s.handleDidSave(msg)
	case methodCompletion:
		return s.handleCompletion(msg)
	case methodDocumentSymbol:
		return s.handleDocumentSymbol(msg)
	case methodWorkspaceSymbol:
		return s.handleWorkspaceSymbol(msg)
	case methodDocumentFormat:
		return s.handleDocumentFormatting(msg)
	case methodDidChangeConfig, methodDidChangeWatched:
		return nil
	default:
		if len(msg.ID) > 0 {
			return s.replyError(msg.ID, -32601, fmt.Sprintf("method %s not found", msg.Method))
		}
	}
	return nil
}

func (s *Server) handleInitialize(msg requestMessage) error {
	type initializeParams struct {
		Capabilities map[string]interface{} `json:"capabilities"`
		ClientInfo   map[string]string      `json:"clientInfo"`
	}
	var params initializeParams
	_ = json.Unmarshal(msg.Params, &params)

	result := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"textDocumentSync": map[string]interface{}{
				"openClose": true,
				"change":    1, // TextDocumentSyncKindFull
				"save": map[string]bool{
					"includeText": true,
				},
			},
			"completionProvider": map[string]interface{}{
				"triggerCharacters": []string{".", ":", "@"},
			},
			"documentSymbolProvider":     true,
			"workspaceSymbolProvider":    true,
			"documentFormattingProvider": true,
			"hoverProvider":              false,
			"definitionProvider":         false,
		},
		"serverInfo": map[string]string{
			"name":    "selene-lsp",
			"version": "0.1.0",
		},
	}

	return s.reply(msg.ID, result)
}

func (s *Server) handleShutdown(msg requestMessage) error {
	s.shuttingDown = true
	return s.reply(msg.ID, nil)
}

func (s *Server) handleDidOpen(msg requestMessage) {
	type didOpenParams struct {
		TextDocument struct {
			URI      string `json:"uri"`
			Language string `json:"languageId"`
			Version  int    `json:"version"`
			Text     string `json:"text"`
		} `json:"textDocument"`
	}
	var params didOpenParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return
	}

	s.mu.Lock()
	s.documents[params.TextDocument.URI] = &document{Text: params.TextDocument.Text, Version: params.TextDocument.Version}
	s.mu.Unlock()

	s.publishDiagnostics(params.TextDocument.URI, analyzeDocument(params.TextDocument.Text))
}

func (s *Server) handleDidChange(msg requestMessage) {
	type contentChange struct {
		Text string `json:"text"`
	}
	type didChangeParams struct {
		TextDocument struct {
			URI     string `json:"uri"`
			Version int    `json:"version"`
		} `json:"textDocument"`
		ContentChanges []contentChange `json:"contentChanges"`
	}
	var params didChangeParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return
	}
	if len(params.ContentChanges) == 0 {
		return
	}

	newText := params.ContentChanges[len(params.ContentChanges)-1].Text

	s.mu.Lock()
	s.documents[params.TextDocument.URI] = &document{Text: newText, Version: params.TextDocument.Version}
	s.mu.Unlock()

	s.publishDiagnostics(params.TextDocument.URI, analyzeDocument(newText))
}

func (s *Server) handleDidSave(msg requestMessage) {
	type didSaveParams struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
		Text string `json:"text"`
	}
	var params didSaveParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return
	}

	text := params.Text
	if text == "" {
		s.mu.Lock()
		if doc, ok := s.documents[params.TextDocument.URI]; ok {
			text = doc.Text
		}
		s.mu.Unlock()
	}
	if text == "" {
		return
	}
	s.publishDiagnostics(params.TextDocument.URI, analyzeDocument(text))
}

func (s *Server) handleDidClose(msg requestMessage) {
	type didCloseParams struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	var params didCloseParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return
	}

	s.mu.Lock()
	delete(s.documents, params.TextDocument.URI)
	s.mu.Unlock()

	s.publishDiagnostics(params.TextDocument.URI, []Diagnostic{})
}

func (s *Server) handleCompletion(msg requestMessage) error {
	type completionParams struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	var params completionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.replyError(msg.ID, -32602, "invalid params")
	}

	items := completionItems()
	for i := range items {
		if items[i].InsertText == "" {
			items[i].InsertText = items[i].Label
			items[i].InsertTextFormat = insertTextPlainText
		}
	}

	return s.reply(msg.ID, CompletionList{IsIncomplete: false, Items: items})
}

func (s *Server) handleDocumentSymbol(msg requestMessage) error {
	type documentSymbolParams struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	var params documentSymbolParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.replyError(msg.ID, -32602, "invalid params")
	}
	text := s.lookupDocument(params.TextDocument.URI)
	if text == "" {
		return s.reply(msg.ID, []DocumentSymbol{})
	}
	symbols := documentSymbols(text)
	return s.reply(msg.ID, symbols)
}

func (s *Server) handleWorkspaceSymbol(msg requestMessage) error {
	type workspaceSymbolParams struct {
		Query string `json:"query"`
	}
	var params workspaceSymbolParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.replyError(msg.ID, -32602, "invalid params")
	}
	s.mu.Lock()
	docs := make(map[string]*document, len(s.documents))
	for uri, doc := range s.documents {
		docs[uri] = &document{Text: doc.Text, Version: doc.Version}
	}
	s.mu.Unlock()
	symbols := workspaceSymbols(docs, params.Query)
	return s.reply(msg.ID, symbols)
}

func (s *Server) handleDocumentFormatting(msg requestMessage) error {
	type formattingParams struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	var params formattingParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.replyError(msg.ID, -32602, "invalid params")
	}
	uri := params.TextDocument.URI
	text := s.lookupDocument(uri)
	formatted, err := format.Source(text)
	if err != nil {
		return s.replyError(msg.ID, -32603, err.Error())
	}
	if formatted == text {
		return s.reply(msg.ID, []TextEdit{})
	}
	edit := TextEdit{Range: fullDocumentRange(text), NewText: formatted}
	s.mu.Lock()
	if doc, ok := s.documents[uri]; ok {
		doc.Text = formatted
	}
	s.mu.Unlock()
	return s.reply(msg.ID, []TextEdit{edit})
}

func (s *Server) lookupDocument(uri string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if doc, ok := s.documents[uri]; ok && doc != nil {
		return doc.Text
	}
	return ""
}

func (s *Server) publishDiagnostics(uri string, diagnostics []Diagnostic) {
	notification := map[string]interface{}{
		"jsonrpc": jsonrpcVersion,
		"method":  "textDocument/publishDiagnostics",
		"params": map[string]interface{}{
			"uri":         uri,
			"diagnostics": diagnostics,
		},
	}
	_ = s.writeMessage(notification)
}

func (s *Server) reply(id json.RawMessage, result interface{}) error {
	resp := responseMessage{JSONRPC: jsonrpcVersion, ID: id, Result: result}
	return s.writeMessage(resp)
}

func (s *Server) replyError(id json.RawMessage, code int, message string) error {
	resp := responseMessage{JSONRPC: jsonrpcVersion, ID: id, Error: &responseError{Code: code, Message: message}}
	return s.writeMessage(resp)
}

func (s *Server) writeMessage(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := s.writer.WriteString(header); err != nil {
		return err
	}
	if _, err := s.writer.Write(data); err != nil {
		return err
	}
	return s.writer.Flush()
}

func (s *Server) readMessage() ([]byte, error) {
	headers := make(map[string]string)
	for {
		line, err := s.reader.ReadString('\n')
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
	if _, err := io.ReadFull(s.reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func fullDocumentRange(text string) Range {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return Range{Start: Position{}, End: Position{}}
	}
	endLine := len(lines) - 1
	endChar := len([]rune(lines[endLine]))
	return Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: endLine, Character: endChar}}
}
