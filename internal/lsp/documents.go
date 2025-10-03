package lsp

import (
	"sort"
	"strings"
	"sync"

	"selenelang/internal/ast"
	"selenelang/internal/token"
)

type DocumentStore struct {
	mu       sync.RWMutex
	analyzer *Analyzer
	docs     map[string]*documentState
}

type documentState struct {
	uri      string
	version  int
	text     string
	analysis AnalysisResult
}

type DocumentSnapshot struct {
	URI         string
	Version     int
	Text        string
	Tokens      []token.Token
	Program     *ast.Program
	Symbols     *SymbolIndex
	Diagnostics []Diagnostic
}

func NewDocumentStore(analyzer *Analyzer) *DocumentStore {
	if analyzer == nil {
		analyzer = NewAnalyzer(nil)
	}
	return &DocumentStore{
		analyzer: analyzer,
		docs:     make(map[string]*documentState),
	}
}

func (ds *DocumentStore) Open(uri string, version int, text string) *DocumentSnapshot {
	analysis := ds.analyzer.Analyze(text)
	state := &documentState{uri: uri, version: version, text: text, analysis: analysis}
	ds.mu.Lock()
	ds.docs[uri] = state
	ds.mu.Unlock()
	return state.snapshot()
}

func (ds *DocumentStore) Update(uri string, version int, text string) *DocumentSnapshot {
	analysis := ds.analyzer.Analyze(text)
	ds.mu.Lock()
	state := &documentState{uri: uri, version: version, text: text, analysis: analysis}
	ds.docs[uri] = state
	ds.mu.Unlock()
	return state.snapshot()
}

func (ds *DocumentStore) Save(uri string, version int, text string) *DocumentSnapshot {
	return ds.Update(uri, version, text)
}

func (ds *DocumentStore) Close(uri string) {
	ds.mu.Lock()
	delete(ds.docs, uri)
	ds.mu.Unlock()
}

func (ds *DocumentStore) Snapshot(uri string) (*DocumentSnapshot, bool) {
	ds.mu.RLock()
	state, ok := ds.docs[uri]
	ds.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return state.snapshot(), true
}

func (ds *DocumentStore) AllSnapshots() []*DocumentSnapshot {
	ds.mu.RLock()
	result := make([]*DocumentSnapshot, 0, len(ds.docs))
	for _, state := range ds.docs {
		result = append(result, state.snapshot())
	}
	ds.mu.RUnlock()
	return result
}

func (ds *DocumentStore) WorkspaceSymbols(query string) []SymbolInformation {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	lower := strings.ToLower(query)
	infos := make([]SymbolInformation, 0)
	for uri, state := range ds.docs {
		if state.analysis.Symbols == nil {
			continue
		}
		flattened := flattenDocumentSymbols(uri, state.analysis.Symbols.DocumentSymbols)
		for _, info := range flattened {
			if lower == "" || strings.Contains(strings.ToLower(info.Name), lower) {
				infos = append(infos, info)
			}
		}
	}
	sort.Slice(infos, func(i, j int) bool {
		if infos[i].Name == infos[j].Name {
			if infos[i].Location.URI == infos[j].Location.URI {
				if infos[i].Location.Range.Start.Line == infos[j].Location.Range.Start.Line {
					return infos[i].Location.Range.Start.Character < infos[j].Location.Range.Start.Character
				}
				return infos[i].Location.Range.Start.Line < infos[j].Location.Range.Start.Line
			}
			return infos[i].Location.URI < infos[j].Location.URI
		}
		return infos[i].Name < infos[j].Name
	})
	return infos
}

func (ds *DocumentStore) Text(uri string) string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if state, ok := ds.docs[uri]; ok {
		return state.text
	}
	return ""
}

func (d *documentState) snapshot() *DocumentSnapshot {
	tokens := make([]token.Token, len(d.analysis.Tokens))
	copy(tokens, d.analysis.Tokens)
	diags := append([]Diagnostic(nil), d.analysis.Diagnostics...)
	return &DocumentSnapshot{
		URI:         d.uri,
		Version:     d.version,
		Text:        d.text,
		Tokens:      tokens,
		Program:     d.analysis.Program,
		Symbols:     d.analysis.Symbols,
		Diagnostics: diags,
	}
}
