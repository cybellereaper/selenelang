package format

import (
	"fmt"
	"strings"

	"selenelang/internal/lexer"
	"selenelang/internal/token"
)

var noSpaceAfter = map[token.Type]bool{
	token.LPAREN:   true,
	token.LBRACKET: true,
	token.DOT:      true,
	token.SAFE_DOT: true,
}

var noSpaceBefore = map[token.Type]bool{
	token.RPAREN:    true,
	token.RBRACKET:  true,
	token.RBRACE:    true,
	token.COMMA:     true,
	token.SEMICOLON: true,
	token.DOT:       true,
	token.SAFE_DOT:  true,
	token.COLON:     true,
	token.NON_NULL:  true,
}

// Source formats Selene source code into a canonical layout.
func Source(src string) (string, error) {
	lex := lexer.New(src)
	tokens := make([]token.Token, 0, len(src)/4)
	for {
		tok := lex.NextToken()
		if tok.Type == token.ILLEGAL {
			return "", fmt.Errorf("illegal token %q at %s", tok.Literal, tok.Pos)
		}
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	var b strings.Builder
	indent := 0
	newLine := true
	var prev token.Token
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.RBRACE {
			if indent > 0 {
				indent--
			}
			if !newLine {
				b.WriteByte('\n')
			}
			writeIndent(&b, indent)
			newLine = false
		} else if newLine {
			writeIndent(&b, indent)
			newLine = false
		} else if needsSpace(prev.Type, tok.Type) {
			b.WriteByte(' ')
		}

		b.WriteString(tok.Literal)

		switch tok.Type {
		case token.LBRACE:
			b.WriteByte('\n')
			indent++
			newLine = true
		case token.RBRACE:
			b.WriteByte('\n')
			newLine = true
		case token.SEMICOLON:
			b.WriteByte('\n')
			newLine = true
		case token.COMMA:
			b.WriteByte(' ')
			newLine = false
		case token.COLON:
			b.WriteByte(' ')
			newLine = false
		case token.ARROW, token.ELVIS, token.ASSIGN, token.PLUS_ASSIGN, token.MINUS_ASSIGN, token.STAR_ASSIGN, token.SLASH_ASSIGN, token.PERCENT_ASSIGN,
			token.PLUS, token.MINUS, token.ASTERISK, token.SLASH, token.PERCENT,
			token.EQ, token.NOT_EQ, token.LT, token.LTE, token.GT, token.GTE, token.OR, token.AND, token.IS, token.NOT_IS:
			b.WriteByte(' ')
			newLine = false
		default:
			newLine = false
		}
		prev = tok
	}
	out := strings.TrimRight(b.String(), "\n") + "\n"
	return out, nil
}

func needsSpace(prev, curr token.Type) bool {
	if prev == "" {
		return false
	}
	if noSpaceAfter[prev] || noSpaceBefore[curr] {
		return false
	}
	if isIdentifierLike(prev) && isIdentifierLike(curr) {
		return true
	}
	if isLiteral(prev) && isLiteral(curr) {
		return true
	}
	if isIdentifierLike(prev) && (isLiteral(curr) || isKeyword(curr)) {
		return true
	}
	if isIdentifierLike(curr) && (isLiteral(prev) || isKeyword(prev)) {
		return true
	}
	if isKeyword(prev) && isKeyword(curr) {
		return true
	}
	if isKeyword(prev) && (isIdentifierLike(curr) || isLiteral(curr)) {
		return true
	}
	return false
}

func isIdentifierLike(t token.Type) bool {
	switch t {
	case token.IDENT:
		return true
	}
	return false
}

func isKeyword(t token.Type) bool {
	switch t {
	case token.LET, token.VAR, token.FN, token.ASYNC, token.CONTRACT, token.RETURNS, token.CLASS,
		token.STRUCT, token.ENUM, token.MATCH, token.MODULE, token.IMPORT, token.AS, token.PACKAGE,
		token.INTERFACE, token.IF, token.ELSE, token.WHILE, token.FOR, token.RETURN, token.BREAK,
		token.CONTINUE, token.AWAIT, token.TRY, token.CATCH, token.FINALLY, token.THROW, token.USING,
		token.EXT, token.CONDITION, token.WHEN:
		return true
	}
	return false
}

func isLiteral(t token.Type) bool {
	switch t {
	case token.NUMBER, token.STRING, token.FORMATSTRING, token.RAWSTRING, token.TRUE, token.FALSE, token.NULL:
		return true
	}
	return false
}

func writeIndent(b *strings.Builder, indent int) {
	for i := 0; i < indent; i++ {
		b.WriteString("    ")
	}
}
