package lsp

import "testing"

func TestSemanticTokensIncludeStrings(t *testing.T) {
	source := "fn greet() {\n    let message = \"hello\"\n}\n"
	analyzer := NewAnalyzer(NewLinter())
	docs := NewDocumentStore(analyzer)
	snapshot := docs.Open("file:///highlight.sel", 1, source)
	highlighter := NewHighlighter()
	tokens := highlighter.Encode(snapshot)
	if len(tokens.Data) == 0 {
		t.Fatalf("expected semantic tokens, got none")
	}
	types, _ := highlighter.Legend()
	stringIndex := -1
	for i, name := range types {
		if name == "string" {
			stringIndex = i
			break
		}
	}
	if stringIndex == -1 {
		t.Fatalf("string token type missing from legend: %v", types)
	}
	if !containsTokenType(tokens, uint32(stringIndex)) {
		t.Fatalf("expected string token classification in %v", tokens.Data)
	}
}

func containsTokenType(tokens SemanticTokens, tokenType uint32) bool {
	for i := 0; i+3 < len(tokens.Data); i += 5 {
		if tokens.Data[i+3] == tokenType {
			return true
		}
	}
	return false
}
