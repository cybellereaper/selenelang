package lsp

import (
	"github.com/cybellereaper/selenelang/internal/ast"
	"github.com/cybellereaper/selenelang/internal/token"
)

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

func rangeFromToken(tok token.Token) Range {
	start := positionFromTokenPos(tok.Pos)
	end := positionFromTokenPos(tok.End)
	if end.Line == start.Line && end.Character <= start.Character {
		end.Character = start.Character + len([]rune(tok.Literal))
		if end.Character == start.Character {
			end.Character++
		}
	}
	return Range{Start: start, End: end}
}

func rangeFromPositions(start, end token.Position) Range {
	return Range{Start: positionFromTokenPos(start), End: positionFromTokenPos(end)}
}

func rangeFromNode(node ast.Node) Range {
	if node == nil {
		return Range{}
	}
	return rangeFromPositions(node.Pos(), node.End())
}

func rangeFromIdentifier(id *ast.Identifier) Range {
	if id == nil {
		return Range{}
	}
	return rangeFromPositions(id.Pos(), id.End())
}

func rangeContains(r Range, pos Position) bool {
	if !rangeIsValid(r) {
		return false
	}
	if comparePosition(pos, r.Start) < 0 {
		return false
	}
	if comparePosition(pos, r.End) > 0 {
		return false
	}
	return true
}

func rangeIsValid(r Range) bool {
	return r.Start.Line >= 0 && r.Start.Character >= 0 &&
		(r.End.Line > r.Start.Line || (r.End.Line == r.Start.Line && r.End.Character >= r.Start.Character))
}

func comparePosition(a, b Position) int {
	if a.Line < b.Line {
		return -1
	}
	if a.Line > b.Line {
		return 1
	}
	switch {
	case a.Character < b.Character:
		return -1
	case a.Character > b.Character:
		return 1
	default:
		return 0
	}
}
