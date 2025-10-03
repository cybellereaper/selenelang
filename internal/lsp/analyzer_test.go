package lsp

import (
	"strings"
	"testing"
)

func TestAnalyzerProducesDiagnostics(t *testing.T) {
	text := "let foo = 1  \nfn bar() {}\n// TODO: revisit\nfn unused() {}\nlet unusedVar = 42\n"
	analyzer := NewAnalyzer(NewLinter())
	result := analyzer.Analyze(text)
	if len(result.Diagnostics) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}
	if !containsDiagnostic(result.Diagnostics, "trailing whitespace") {
		t.Fatalf("expected trailing whitespace diagnostic, got %v", result.Diagnostics)
	}
	if !containsDiagnostic(result.Diagnostics, "TODO comment") {
		t.Fatalf("expected TODO diagnostic, got %v", result.Diagnostics)
	}
	if !containsDiagnostic(result.Diagnostics, "declared but never used") {
		t.Fatalf("expected unused variable diagnostic, got %v", result.Diagnostics)
	}
}

func containsDiagnostic(diags []Diagnostic, substr string) bool {
	for _, d := range diags {
		if strings.Contains(d.Message, substr) {
			return true
		}
	}
	return false
}
