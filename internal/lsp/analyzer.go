// Package lsp implements the Selene language server analysis pipeline.
package lsp

import (
	"fmt"

	"github.com/cybellereaper/selenelang/internal/ast"
	"github.com/cybellereaper/selenelang/internal/lexer"
	"github.com/cybellereaper/selenelang/internal/parser"
	"github.com/cybellereaper/selenelang/internal/token"
)

const diagnosticSource = "selene"

// AnalysisResult aggregates data produced while analyzing a document.
type AnalysisResult struct {
	Tokens      []token.Token
	Program     *ast.Program
	Diagnostics []Diagnostic
	Symbols     *SymbolIndex
}

// Analyzer coordinates lexical, syntactic, and linting passes.
type Analyzer struct {
	linter *Linter
}

// NewAnalyzer constructs an Analyzer, defaulting to a fresh linter when nil.
func NewAnalyzer(linter *Linter) *Analyzer {
	if linter == nil {
		linter = NewLinter()
	}
	return &Analyzer{linter: linter}
}

// Analyze runs the lexer, parser, and linter to produce diagnostics and symbols.
func (a *Analyzer) Analyze(text string) AnalysisResult {
	tokens, lexDiagnostics := lexDocument(text)
	program, parseDiagnostics := parseDocument(text)
	symbols := buildSymbolIndex(program, tokens)

	diagnostics := append([]Diagnostic{}, lexDiagnostics...)
	diagnostics = append(diagnostics, parseDiagnostics...)
	diagnostics = append(diagnostics, a.linter.Lint(text, program, tokens, symbols)...)

	return AnalysisResult{
		Tokens:      tokens,
		Program:     program,
		Diagnostics: diagnostics,
		Symbols:     symbols,
	}
}

func lexDocument(source string) ([]token.Token, []Diagnostic) {
	lex := lexer.New(source)
	tokens := make([]token.Token, 0, len(source)/4)
	diagnostics := make([]Diagnostic, 0)
	for {
		tok := lex.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.ILLEGAL {
			diagnostics = append(diagnostics, Diagnostic{
				Range:    rangeFromToken(tok),
				Severity: severityError,
				Source:   diagnosticSource,
				Message:  fmt.Sprintf("illegal token %q", tok.Literal),
			})
		}
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens, diagnostics
}

func parseDocument(source string) (*ast.Program, []Diagnostic) {
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	diagnostics := make([]Diagnostic, 0, len(p.Errors()))
	for _, perr := range p.ErrorDetails() {
		diagnostics = append(diagnostics, Diagnostic{
			Range:    rangeFromPositions(perr.Position, perr.Position),
			Severity: severityError,
			Source:   diagnosticSource,
			Message:  perr.Message,
		})
	}
	return program, diagnostics
}
