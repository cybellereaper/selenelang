package lsp

import (
	"fmt"
	"strings"
	"unicode"

	"selenelang/internal/ast"
	"selenelang/internal/token"
)

type Linter struct{}

func NewLinter() *Linter {
	return &Linter{}
}

func (l *Linter) Lint(text string, program *ast.Program, tokens []token.Token, symbols *SymbolIndex) []Diagnostic {
	diagnostics := make([]Diagnostic, 0)
	diagnostics = append(diagnostics, l.trailingWhitespace(text)...)
	diagnostics = append(diagnostics, l.longLines(text)...)
	diagnostics = append(diagnostics, l.missingFinalNewline(text)...)
	diagnostics = append(diagnostics, l.todoComments(text)...)
	diagnostics = append(diagnostics, l.unusedVariables(tokens, symbols)...)
	diagnostics = append(diagnostics, l.functionsWithoutBody(symbols)...)
	return diagnostics
}

func (l *Linter) trailingWhitespace(text string) []Diagnostic {
	lines := strings.Split(text, "\n")
	diags := make([]Diagnostic, 0)
	for i, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if trimmed != line {
			start := len([]rune(trimmed))
			end := len([]rune(line))
			diags = append(diags, Diagnostic{
				Range: Range{
					Start: Position{Line: i, Character: start},
					End:   Position{Line: i, Character: end},
				},
				Severity: severityWarning,
				Source:   diagnosticSource,
				Message:  "trailing whitespace",
			})
		}
	}
	return diags
}

func (l *Linter) longLines(text string) []Diagnostic {
	lines := strings.Split(text, "\n")
	diags := make([]Diagnostic, 0)
	const limit = 120
	for i, line := range lines {
		runeCount := len([]rune(line))
		if runeCount > limit {
			diags = append(diags, Diagnostic{
				Range: Range{
					Start: Position{Line: i, Character: limit},
					End:   Position{Line: i, Character: runeCount},
				},
				Severity: severityWarning,
				Source:   diagnosticSource,
				Message:  fmt.Sprintf("line exceeds %d characters (%d)", limit, runeCount),
			})
		}
	}
	return diags
}

func (l *Linter) missingFinalNewline(text string) []Diagnostic {
	if text == "" {
		return nil
	}
	runes := []rune(text)
	if runes[len(runes)-1] == '\n' {
		return nil
	}
	pos := positionForRuneOffset(text, len(runes))
	return []Diagnostic{{
		Range:    Range{Start: pos, End: pos},
		Severity: severityWarning,
		Source:   diagnosticSource,
		Message:  "file does not end with a newline",
	}}
}

func (l *Linter) todoComments(text string) []Diagnostic {
	lines := strings.Split(text, "\n")
	diags := make([]Diagnostic, 0)
	for i, line := range lines {
		idx := strings.Index(line, "//")
		if idx >= 0 {
			comment := line[idx+2:]
			upper := strings.ToUpper(comment)
			todoIdx := strings.Index(upper, "TODO")
			if todoIdx >= 0 {
				startChar := len([]rune(line[:idx+2+todoIdx]))
				endChar := startChar + len("TODO")
				diags = append(diags, Diagnostic{
					Range: Range{
						Start: Position{Line: i, Character: startChar},
						End:   Position{Line: i, Character: endChar},
					},
					Severity: severityWarning,
					Source:   diagnosticSource,
					Message:  "TODO comment",
				})
			}
		}
	}
	return diags
}

func (l *Linter) unusedVariables(tokens []token.Token, symbols *SymbolIndex) []Diagnostic {
	if symbols == nil {
		return nil
	}
	counts := make(map[string]int)
	for _, tok := range tokens {
		if tok.Type == token.IDENT {
			counts[tok.Literal]++
		}
	}
	diags := make([]Diagnostic, 0)
	for _, variable := range symbols.VariableSymbols {
		name := variable.Name
		if name == "" || name == "_" {
			continue
		}
		if counts[name] <= 1 {
			if first := []rune(name)[0]; unicode.IsUpper(first) {
				continue
			}
			diags = append(diags, Diagnostic{
				Range:    variable.Range,
				Severity: severityWarning,
				Source:   diagnosticSource,
				Message:  fmt.Sprintf("variable %q declared but never used", name),
			})
		}
	}
	return diags
}

func (l *Linter) functionsWithoutBody(symbols *SymbolIndex) []Diagnostic {
	if symbols == nil {
		return nil
	}
	diags := make([]Diagnostic, 0)
	for _, fn := range symbols.FunctionSymbols {
		if fn.HasBody {
			continue
		}
		diags = append(diags, Diagnostic{
			Range:    fn.Range,
			Severity: severityWarning,
			Source:   diagnosticSource,
			Message:  fmt.Sprintf("function %q has no implementation", fn.Name),
		})
	}
	return diags
}
