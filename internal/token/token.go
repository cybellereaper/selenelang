package token

import "fmt"

// Type represents the type of lexical token.
type Type string

// Token represents a lexical token with positional metadata.
type Token struct {
	Type    Type
	Literal string
	Pos     Position
	End     Position
}

// Position describes a location within a source file.
type Position struct {
	Offset int
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

const (
	// Special tokens
	ILLEGAL Type = "ILLEGAL"
	EOF     Type = "EOF"

	// Identifiers + literals
	IDENT  Type = "IDENT"
	NUMBER Type = "NUMBER"
	STRING Type = "STRING"
	TRUE   Type = "TRUE"
	FALSE  Type = "FALSE"
	NULL   Type = "NULL"

	// Operators
	ASSIGN   Type = "="
	PLUS     Type = "+"
	MINUS    Type = "-"
	ASTERISK Type = "*"
	SLASH    Type = "/"
	PERCENT  Type = "%"
	BANG     Type = "!"
	QUESTION Type = "?"
	COLON    Type = ":"
	ELVIS    Type = "?:"
	SAFE_DOT Type = "?."
	NON_NULL Type = "!!"
	EQ       Type = "=="
	NOT_EQ   Type = "!="
	LT       Type = "<"
	LTE      Type = "<="
	GT       Type = ">"
	GTE      Type = ">="
	OR       Type = "||"
	AND      Type = "&&"
	ARROW    Type = "=>"
	IS       Type = "is"
	NOT_IS   Type = "!is"

	// Delimiters
	COMMA     Type = ","
	DOT       Type = "."
	SEMICOLON Type = ";"
	LPAREN    Type = "("
	RPAREN    Type = ")"
	LBRACE    Type = "{"
	RBRACE    Type = "}"
	LBRACKET  Type = "["
	RBRACKET  Type = "]"
)

const (
	// Keywords
	LET      Type = "let"
	VAR      Type = "var"
	FN       Type = "fn"
	ASYNC    Type = "async"
	CONTRACT Type = "contract"
	RETURNS  Type = "returns"
	CLASS    Type = "class"
	STRUCT   Type = "struct"
	ENUM     Type = "enum"
	MATCH    Type = "match"
	MODULE   Type = "module"
	IMPORT   Type = "import"
	AS       Type = "as"
	AWAIT    Type = "await"
)

var keywords = map[string]Type{
	"let":      LET,
	"var":      VAR,
	"fn":       FN,
	"async":    ASYNC,
	"contract": CONTRACT,
	"returns":  RETURNS,
	"class":    CLASS,
	"struct":   STRUCT,
	"enum":     ENUM,
	"match":    MATCH,
	"module":   MODULE,
	"import":   IMPORT,
	"as":       AS,
	"true":     TRUE,
	"false":    FALSE,
	"null":     NULL,
	"is":       IS,
	"await":    AWAIT,
}

// LookupIdent identifies reserved keywords.
func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
