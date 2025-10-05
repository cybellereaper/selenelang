package lsp

import (
	"math"
	"strconv"
	"testing"
)

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

func TestSafeUint32(t *testing.T) {
	t.Parallel()

	maxUint32 := uint64(math.MaxUint32)
	tests := []struct {
		name  string
		input int
		ok    bool
		want  uint32
	}{
		{name: "zero", input: 0, ok: true, want: 0},
		{name: "max", input: int(maxUint32), ok: true, want: uint32(maxUint32)},
		{name: "negative", input: -1, ok: false, want: 0},
	}

	if strconv.IntSize > 32 {
		tests = append(tests, struct {
			name  string
			input int
			ok    bool
			want  uint32
		}{name: "overflow", input: int(maxUint32 + 1), ok: false, want: 0})
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := safeUint32(tc.input)
			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v", tc.ok, ok)
			}
			if got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}
