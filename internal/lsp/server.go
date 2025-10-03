package lsp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	"github.com/cybellereaper/selenelang/internal/format"
)

// Server implements the Selene language server protocol surface.
type Server struct {
	conn         *jsonRPCConnection
	documents    *DocumentStore
	completer    *Completer
	highlighter  *Highlighter
	shuttingDown int32
}

// NewServer wires together the JSON-RPC transport and language features.
func NewServer(r io.Reader, w io.Writer) *Server {
	analyzer := NewAnalyzer(nil)
	return &Server{
		conn:        newJSONRPCConnection(r, w),
		documents:   NewDocumentStore(analyzer),
		completer:   NewCompleter(),
		highlighter: NewHighlighter(),
	}
}

// Run processes incoming requests until the client disconnects.
func (s *Server) Run() error {
	for {
		msg, err := s.conn.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			var syntaxErr *json.SyntaxError
			if errors.As(err, &syntaxErr) {
				continue
			}
			if errors.Is(err, errClientExit) {
				return nil
			}
			return err
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

const (
	methodInitialize             = "initialize"
	methodInitialized            = "initialized"
	methodShutdown               = "shutdown"
	methodExit                   = "exit"
	methodDidOpen                = "textDocument/didOpen"
	methodDidChange              = "textDocument/didChange"
	methodDidClose               = "textDocument/didClose"
	methodDidSave                = "textDocument/didSave"
	methodCompletion             = "textDocument/completion"
	methodHover                  = "textDocument/hover"
	methodDocumentSymbol         = "textDocument/documentSymbol"
	methodWorkspaceSymbol        = "workspace/symbol"
	methodDocumentFormat         = "textDocument/formatting"
	methodSemanticTokensFull     = "textDocument/semanticTokens/full"
	methodSemanticTokensRange    = "textDocument/semanticTokens/range"
	methodDidChangeConfiguration = "workspace/didChangeConfiguration"
	methodDidChangeWatchedFiles  = "workspace/didChangeWatchedFiles"
)

func (s *Server) dispatch(msg requestMessage) error {
	switch msg.Method {
	case methodInitialize:
		return s.handleInitialize(msg)
	case methodInitialized:
		return nil
	case methodShutdown:
		return s.handleShutdown(msg)
	case methodExit:
		if atomic.LoadInt32(&s.shuttingDown) == 0 {
			return errClientExit
		}
		return nil
	case methodDidOpen:
		return s.handleDidOpen(msg)
	case methodDidChange:
		return s.handleDidChange(msg)
	case methodDidClose:
		return s.handleDidClose(msg)
	case methodDidSave:
		return s.handleDidSave(msg)
	case methodCompletion:
		return s.handleCompletion(msg)
	case methodHover:
		return s.handleHover(msg)
	case methodDocumentSymbol:
		return s.handleDocumentSymbol(msg)
	case methodWorkspaceSymbol:
		return s.handleWorkspaceSymbol(msg)
	case methodDocumentFormat:
		return s.handleDocumentFormatting(msg)
	case methodSemanticTokensFull:
		return s.handleSemanticTokensFull(msg)
	case methodSemanticTokensRange:
		return s.handleSemanticTokensRange(msg)
	case methodDidChangeConfiguration, methodDidChangeWatchedFiles:
		return nil
	default:
		if len(msg.ID) > 0 {
			return s.conn.ReplyError(msg.ID, -32601, fmt.Sprintf("method %s not found", msg.Method))
		}
		return nil
	}
}

func (s *Server) handleInitialize(msg requestMessage) error {
	var params struct {
		Capabilities map[string]any `json:"capabilities"`
		ClientInfo   map[string]any `json:"clientInfo"`
	}
	_ = json.Unmarshal(msg.Params, &params)
	tokenTypes, tokenModifiers := s.highlighter.Legend()
	result := map[string]any{
		"capabilities": map[string]any{
			"textDocumentSync": map[string]any{
				"openClose": true,
				"change":    1,
				"save": map[string]bool{
					"includeText": true,
				},
			},
			"completionProvider": map[string]any{
				"triggerCharacters": []string{".", ":", "@", "(", ">"},
			},
			"hoverProvider":              true,
			"documentSymbolProvider":     true,
			"workspaceSymbolProvider":    true,
			"documentFormattingProvider": true,
			"semanticTokensProvider": map[string]any{
				"legend": map[string]any{
					"tokenTypes":     tokenTypes,
					"tokenModifiers": tokenModifiers,
				},
				"range": true,
				"full":  true,
			},
		},
		"serverInfo": map[string]string{
			"name":    "selene-lsp",
			"version": "0.2.0",
		},
	}
	return s.conn.Reply(msg.ID, result)
}

func (s *Server) handleShutdown(msg requestMessage) error {
	atomic.StoreInt32(&s.shuttingDown, 1)
	return s.conn.Reply(msg.ID, nil)
}

func (s *Server) handleDidOpen(msg requestMessage) error {
	var params struct {
		TextDocument struct {
			URI     string `json:"uri"`
			Version int    `json:"version"`
			Text    string `json:"text"`
		} `json:"textDocument"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil
	}
	snapshot := s.documents.Open(params.TextDocument.URI, params.TextDocument.Version, params.TextDocument.Text)
	s.publishDiagnostics(params.TextDocument.URI, snapshot.Diagnostics)
	return nil
}

func (s *Server) handleDidChange(msg requestMessage) error {
	var params struct {
		TextDocument struct {
			URI     string `json:"uri"`
			Version int    `json:"version"`
		} `json:"textDocument"`
		ContentChanges []struct {
			Text string `json:"text"`
		} `json:"contentChanges"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil
	}
	if len(params.ContentChanges) == 0 {
		return nil
	}
	text := params.ContentChanges[len(params.ContentChanges)-1].Text
	snapshot := s.documents.Update(params.TextDocument.URI, params.TextDocument.Version, text)
	s.publishDiagnostics(params.TextDocument.URI, snapshot.Diagnostics)
	return nil
}

func (s *Server) handleDidSave(msg requestMessage) error {
	var params struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil
	}
	text := params.Text
	snapshot, ok := s.documents.Snapshot(params.TextDocument.URI)
	if !ok {
		return nil
	}
	version := snapshot.Version
	if text == "" {
		text = snapshot.Text
	}
	snapshot = s.documents.Save(params.TextDocument.URI, version, text)
	s.publishDiagnostics(params.TextDocument.URI, snapshot.Diagnostics)
	return nil
}

func (s *Server) handleDidClose(msg requestMessage) error {
	var params struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil
	}
	s.documents.Close(params.TextDocument.URI)
	s.publishDiagnostics(params.TextDocument.URI, nil)
	return nil
}

func (s *Server) handleCompletion(msg requestMessage) error {
	var params struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
		Position Position `json:"position"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.conn.ReplyError(msg.ID, -32602, "invalid params")
	}
	snapshot, ok := s.documents.Snapshot(params.TextDocument.URI)
	if !ok {
		return s.conn.Reply(msg.ID, CompletionList{IsIncomplete: false, Items: nil})
	}
	list := s.completer.Completion(snapshot, params.Position)
	return s.conn.Reply(msg.ID, list)
}

func (s *Server) handleHover(msg requestMessage) error {
	var params struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
		Position Position `json:"position"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.conn.ReplyError(msg.ID, -32602, "invalid params")
	}
	snapshot, ok := s.documents.Snapshot(params.TextDocument.URI)
	if !ok {
		return s.conn.Reply(msg.ID, nil)
	}
	hover, ok := buildHover(snapshot, params.Position)
	if !ok {
		return s.conn.Reply(msg.ID, nil)
	}
	return s.conn.Reply(msg.ID, hover)
}

func (s *Server) handleDocumentSymbol(msg requestMessage) error {
	var params struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.conn.ReplyError(msg.ID, -32602, "invalid params")
	}
	snapshot, ok := s.documents.Snapshot(params.TextDocument.URI)
	if !ok || snapshot.Symbols == nil {
		return s.conn.Reply(msg.ID, []DocumentSymbol{})
	}
	return s.conn.Reply(msg.ID, snapshot.Symbols.DocumentSymbols)
}

func (s *Server) handleWorkspaceSymbol(msg requestMessage) error {
	var params struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.conn.ReplyError(msg.ID, -32602, "invalid params")
	}
	infos := s.documents.WorkspaceSymbols(params.Query)
	return s.conn.Reply(msg.ID, infos)
}

func (s *Server) handleDocumentFormatting(msg requestMessage) error {
	var params struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.conn.ReplyError(msg.ID, -32602, "invalid params")
	}
	snapshot, ok := s.documents.Snapshot(params.TextDocument.URI)
	if !ok {
		return s.conn.Reply(msg.ID, []TextEdit{})
	}
	formatted, err := format.Source(snapshot.Text)
	if err != nil {
		return s.conn.ReplyError(msg.ID, -32603, err.Error())
	}
	if formatted == snapshot.Text {
		return s.conn.Reply(msg.ID, []TextEdit{})
	}
	edit := TextEdit{Range: fullDocumentRange(snapshot.Text), NewText: formatted}
	return s.conn.Reply(msg.ID, []TextEdit{edit})
}

func (s *Server) handleSemanticTokensFull(msg requestMessage) error {
	var params struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.conn.ReplyError(msg.ID, -32602, "invalid params")
	}
	snapshot, ok := s.documents.Snapshot(params.TextDocument.URI)
	if !ok {
		return s.conn.Reply(msg.ID, SemanticTokens{})
	}
	tokens := s.highlighter.Encode(snapshot)
	return s.conn.Reply(msg.ID, tokens)
}

func (s *Server) handleSemanticTokensRange(msg requestMessage) error {
	var params struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
		Range Range `json:"range"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.conn.ReplyError(msg.ID, -32602, "invalid params")
	}
	snapshot, ok := s.documents.Snapshot(params.TextDocument.URI)
	if !ok {
		return s.conn.Reply(msg.ID, SemanticTokens{})
	}
	tokens := s.highlighter.EncodeRange(snapshot, params.Range)
	return s.conn.Reply(msg.ID, tokens)
}

func (s *Server) publishDiagnostics(uri string, diagnostics []Diagnostic) {
	params := map[string]any{
		"uri":         uri,
		"diagnostics": diagnostics,
	}
	_ = s.conn.Notify("textDocument/publishDiagnostics", params)
}

func buildHover(doc *DocumentSnapshot, pos Position) (Hover, bool) {
	name, rng := identifierAt(doc.Text, pos)
	if name == "" {
		return Hover{}, false
	}
	var content strings.Builder
	if doc.Symbols != nil {
		for _, fn := range doc.Symbols.FunctionSymbols {
			if fn.Name == name {
				content.WriteString(fmt.Sprintf("**function** `%s`\n\n", fn.Detail))
				return Hover{Contents: MarkupContent{Kind: "markdown", Value: content.String()}, Range: &rng}, true
			}
		}
		for _, param := range collectParameters(doc.Symbols) {
			if param.Name == name {
				content.WriteString(fmt.Sprintf("**parameter** `%s`", param.Name))
				return Hover{Contents: MarkupContent{Kind: "markdown", Value: content.String()}, Range: &rng}, true
			}
		}
		for _, variable := range doc.Symbols.VariableSymbols {
			if variable.Name == name {
				detail := "immutable"
				if variable.Mutable {
					detail = "mutable"
				}
				content.WriteString(fmt.Sprintf("**variable** `%s` (%s)", variable.Name, detail))
				return Hover{Contents: MarkupContent{Kind: "markdown", Value: content.String()}, Range: &rng}, true
			}
		}
		for _, t := range doc.Symbols.TypeSymbols {
			if t.Name == name {
				content.WriteString(fmt.Sprintf("**%s** `%s`", strings.ToLower(t.Detail), t.Name))
				return Hover{Contents: MarkupContent{Kind: "markdown", Value: content.String()}, Range: &rng}, true
			}
		}
	}
	if content.Len() == 0 {
		content.WriteString(fmt.Sprintf("`%s`", name))
	}
	return Hover{Contents: MarkupContent{Kind: "markdown", Value: content.String()}, Range: &rng}, true
}

func collectParameters(index *SymbolIndex) []ParameterSymbol {
	params := make([]ParameterSymbol, 0)
	if index == nil {
		return params
	}
	for _, fn := range index.FunctionSymbols {
		params = append(params, fn.Params...)
	}
	return params
}

func fullDocumentRange(text string) Range {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 0}}
	}
	endLine := len(lines) - 1
	endChar := len([]rune(lines[endLine]))
	return Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: endLine, Character: endChar}}
}
