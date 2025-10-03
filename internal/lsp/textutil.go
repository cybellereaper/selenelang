package lsp

import "unicode"

func runeOffsetForPosition(text string, pos Position) (int, bool) {
	if pos.Line < 0 || pos.Character < 0 {
		return 0, false
	}
	runes := []rune(text)
	line := 0
	character := 0
	for i, r := range runes {
		if line == pos.Line && character == pos.Character {
			return i, true
		}
		if r == '\n' {
			line++
			character = 0
		} else {
			character++
		}
	}
	if line == pos.Line && character == pos.Character {
		return len(runes), true
	}
	return len(runes), false
}

func positionForRuneOffset(text string, offset int) Position {
	if offset < 0 {
		offset = 0
	}
	runes := []rune(text)
	if offset > len(runes) {
		offset = len(runes)
	}
	line := 0
	character := 0
	for i := 0; i < offset; i++ {
		r := runes[i]
		if r == '\n' {
			line++
			character = 0
		} else {
			character++
		}
	}
	return Position{Line: line, Character: character}
}

func identifierPrefixAt(text string, pos Position) (string, Position) {
	runes := []rune(text)
	idx, ok := runeOffsetForPosition(text, pos)
	if !ok {
		idx = len(runes)
	}
	if idx > len(runes) {
		idx = len(runes)
	}
	start := idx
	for start > 0 {
		r := runes[start-1]
		if r == '\n' {
			break
		}
		if !isIdentifierRune(r) {
			break
		}
		start--
	}
	prefixRunes := runes[start:idx]
	startPos := positionForRuneOffset(text, start)
	return string(prefixRunes), startPos
}

func identifierAt(text string, pos Position) (string, Range) {
	runes := []rune(text)
	idx, ok := runeOffsetForPosition(text, pos)
	if !ok {
		idx = len(runes)
	}
	if idx > len(runes) {
		idx = len(runes)
	}
	start := idx
	for start > 0 {
		r := runes[start-1]
		if r == '\n' {
			break
		}
		if !isIdentifierRune(r) {
			break
		}
		start--
	}
	end := idx
	for end < len(runes) {
		r := runes[end]
		if r == '\n' {
			break
		}
		if !isIdentifierRune(r) {
			break
		}
		end++
	}
	value := string(runes[start:end])
	return value, Range{Start: positionForRuneOffset(text, start), End: positionForRuneOffset(text, end)}
}

func isIdentifierRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func runeBefore(text string, pos Position) rune {
	runes := []rune(text)
	idx, ok := runeOffsetForPosition(text, pos)
	if !ok {
		idx = len(runes)
	}
	if idx <= 0 || idx > len(runes) {
		return 0
	}
	return runes[idx-1]
}
