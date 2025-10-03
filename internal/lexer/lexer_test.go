package lexer

import (
	"testing"

	"selenelang/internal/token"
)

func TestLexerRecognizesCoreTokens(t *testing.T) {
	input := `
package module import as
let var fn async contract returns class struct enum interface ext match if else while for using try catch finally throw return break continue condition when await
true false null
is !is
+= -= *= /= %= ?: ?. !! && || == != < <= > >= =>
& + - * / % = . , ; ( ) { } [ ]
`

	tests := []struct {
		typ token.Type
		lit string
	}{
		{token.PACKAGE, "package"},
		{token.MODULE, "module"},
		{token.IMPORT, "import"},
		{token.AS, "as"},
		{token.LET, "let"},
		{token.VAR, "var"},
		{token.FN, "fn"},
		{token.ASYNC, "async"},
		{token.CONTRACT, "contract"},
		{token.RETURNS, "returns"},
		{token.CLASS, "class"},
		{token.STRUCT, "struct"},
		{token.ENUM, "enum"},
		{token.INTERFACE, "interface"},
		{token.EXT, "ext"},
		{token.MATCH, "match"},
		{token.IF, "if"},
		{token.ELSE, "else"},
		{token.WHILE, "while"},
		{token.FOR, "for"},
		{token.USING, "using"},
		{token.TRY, "try"},
		{token.CATCH, "catch"},
		{token.FINALLY, "finally"},
		{token.THROW, "throw"},
		{token.RETURN, "return"},
		{token.BREAK, "break"},
		{token.CONTINUE, "continue"},
		{token.CONDITION, "condition"},
		{token.WHEN, "when"},
		{token.AWAIT, "await"},
		{token.TRUE, "true"},
		{token.FALSE, "false"},
		{token.NULL, "null"},
		{token.IS, "is"},
		{token.NOT_IS, "!is"},
		{token.PLUS_ASSIGN, "+="},
		{token.MINUS_ASSIGN, "-="},
		{token.STAR_ASSIGN, "*="},
		{token.SLASH_ASSIGN, "/="},
		{token.PERCENT_ASSIGN, "%="},
		{token.ELVIS, "?:"},
		{token.SAFE_DOT, "?."},
		{token.NON_NULL, "!!"},
		{token.AND, "&&"},
		{token.OR, "||"},
		{token.EQ, "=="},
		{token.NOT_EQ, "!="},
		{token.LT, "<"},
		{token.LTE, "<="},
		{token.GT, ">"},
		{token.GTE, ">="},
		{token.ARROW, "=>"},
		{token.AMPERSAND, "&"},
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.PERCENT, "%"},
		{token.ASSIGN, "="},
		{token.DOT, "."},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.LBRACKET, "["},
		{token.RBRACKET, "]"},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.typ {
			t.Fatalf("test %d: expected token type %s, got %s", i, tt.typ, tok.Type)
		}
		if tok.Literal != tt.lit {
			t.Fatalf("test %d: expected literal %q, got %q", i, tt.lit, tok.Literal)
		}
	}

	eof := l.NextToken()
	if eof.Type != token.EOF {
		t.Fatalf("expected EOF, got %s", eof.Type)
	}
}

func TestLexerParsesStringVariants(t *testing.T) {
	input := "\"plain\" f\"format\" r\"raw\" `tick`"

	l := New(input)

	tok := l.NextToken()
	if tok.Type != token.STRING || tok.Literal != "plain" {
		t.Fatalf("expected plain string, got %s (%q)", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != token.FORMATSTRING || tok.Literal != "format" {
		t.Fatalf("expected format string, got %s (%q)", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != token.RAWSTRING || tok.Literal != "raw" {
		t.Fatalf("expected raw string, got %s (%q)", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != token.RAWSTRING || tok.Literal != "tick" {
		t.Fatalf("expected backtick raw string, got %s (%q)", tok.Type, tok.Literal)
	}
}
