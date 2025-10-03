package lsp

import "testing"

func TestWorkspaceSymbolsFiltering(t *testing.T) {
	analyzer := NewAnalyzer(NewLinter())
	docs := NewDocumentStore(analyzer)
	docs.Open("file:///symbols.sel", 1, "package demo\nfn alpha() {}\nfn beta() {}\n")
	results := docs.WorkspaceSymbols("alp")
	if len(results) != 1 {
		t.Fatalf("expected 1 workspace symbol, got %d", len(results))
	}
	if results[0].Name != "alpha" {
		t.Fatalf("expected alpha symbol, got %s", results[0].Name)
	}
}
