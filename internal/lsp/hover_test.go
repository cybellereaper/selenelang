package lsp

import (
	"strings"
	"testing"
)

func TestBuildHoverForFunction(t *testing.T) {
	source := "fn greet(name) {\n    return name\n}\n"
	analyzer := NewAnalyzer(NewLinter())
	docs := NewDocumentStore(analyzer)
	snapshot := docs.Open("file:///hover.sel", 1, source)
	hover, ok := buildHover(snapshot, Position{Line: 0, Character: 3})
	if !ok {
		t.Fatalf("expected hover information")
	}
	if hover.Contents.Kind != "markdown" {
		t.Fatalf("expected markdown hover, got %s", hover.Contents.Kind)
	}
	if !strings.Contains(hover.Contents.Value, "function") {
		t.Fatalf("unexpected hover contents: %s", hover.Contents.Value)
	}
}
