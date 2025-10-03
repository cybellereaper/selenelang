package token

import "testing"

func TestLookupIdentRecognizesKeywords(t *testing.T) {
	keywords := []Type{
		LET, VAR, FN, ASYNC, CONTRACT, RETURNS, CLASS, STRUCT, ENUM, MATCH,
		MODULE, IMPORT, AS, PACKAGE, INTERFACE, IF, ELSE, WHILE, FOR, RETURN,
		BREAK, CONTINUE, AWAIT, TRY, CATCH, FINALLY, THROW, USING, EXT,
		CONDITION, WHEN,
	}
	for _, kw := range keywords {
		t.Run(string(kw), func(t *testing.T) {
			if got := LookupIdent(string(kw)); got != kw {
				t.Fatalf("LookupIdent(%q) = %s, want %s", kw, got, kw)
			}
		})
	}
}

func TestLookupIdentFallsBackToIdentifier(t *testing.T) {
	cases := []string{"value", "Result", "_ignored", "Type42"}
	for _, ident := range cases {
		if got := LookupIdent(ident); got != IDENT {
			t.Fatalf("LookupIdent(%q) = %s, want IDENT", ident, got)
		}
	}
}

func TestPositionString(t *testing.T) {
	pos := Position{Line: 12, Column: 8}
	if got, want := pos.String(), "12:8"; got != want {
		t.Fatalf("Position.String() = %q, want %q", got, want)
	}
}
