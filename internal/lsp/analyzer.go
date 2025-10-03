package lsp

import (
	"fmt"

	"selenelang/internal/lexer"
	"selenelang/internal/parser"
	"selenelang/internal/token"
)

const diagnosticSource = "selene"

func analyzeDocument(source string) []Diagnostic {
	diags := make([]Diagnostic, 0)

	lex := lexer.New(source)
	for {
		tok := lex.NextToken()
		if tok.Type == token.ILLEGAL {
			message := fmt.Sprintf("illegal token %q", tok.Literal)
			diags = append(diags, Diagnostic{
				Range:    rangeFromToken(tok),
				Severity: severityError,
				Source:   diagnosticSource,
				Message:  message,
			})
		}
		if tok.Type == token.EOF {
			break
		}
	}

	p := parser.New(lexer.New(source))
	_ = p.ParseProgram()
	for _, perr := range p.ErrorDetails() {
		diags = append(diags, Diagnostic{
			Range:    rangeFromPosition(perr.Position),
			Severity: severityError,
			Source:   diagnosticSource,
			Message:  perr.Message,
		})
	}

	return diags
}

func rangeFromToken(tok token.Token) Range {
	start := positionFromTokenPos(tok.Pos)
	end := positionFromTokenPos(tok.End)
	if end.Line == start.Line && end.Character <= start.Character {
		end.Character = start.Character + 1
	}
	return Range{Start: start, End: end}
}

func rangeFromPosition(pos token.Position) Range {
	start := positionFromTokenPos(pos)
	end := start
	end.Character = start.Character + 1
	return Range{Start: start, End: end}
}

func positionFromTokenPos(pos token.Position) Position {
	line := pos.Line - 1
	if line < 0 {
		line = 0
	}
	character := pos.Column - 1
	if character < 0 {
		character = 0
	}
	return Position{Line: line, Character: character}
}
