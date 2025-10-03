package lexer

import (
	"strings"

	"selenelang/internal/token"
)

type Lexer struct {
	input        []rune
	position     int
	readPosition int
	ch           rune
	line         int
	column       int
}

func New(input string) *Lexer {
	l := &Lexer{
		input: []rune(input),
		line:  1,
	}
	l.readRune()
	return l
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespaceAndComments()

	startOffset := l.position
	startLine := l.line
	startColumn := l.column

	tok := token.Token{
		Pos: token.Position{Offset: startOffset, Line: startLine, Column: startColumn},
	}

	switch l.ch {
	case 0:
		tok.Type = token.EOF
		tok.Literal = ""
		tok.End = tok.Pos
		return tok
	case '+':
		tok.Type = token.PLUS
		tok.Literal = "+"
		l.readRune()
	case '-':
		tok.Type = token.MINUS
		tok.Literal = "-"
		l.readRune()
	case '*':
		tok.Type = token.ASTERISK
		tok.Literal = "*"
		l.readRune()
	case '%':
		tok.Type = token.PERCENT
		tok.Literal = "%"
		l.readRune()
	case ',':
		tok.Type = token.COMMA
		tok.Literal = ","
		l.readRune()
	case ';':
		tok.Type = token.SEMICOLON
		tok.Literal = ";"
		l.readRune()
	case '(':
		tok.Type = token.LPAREN
		tok.Literal = "("
		l.readRune()
	case ')':
		tok.Type = token.RPAREN
		tok.Literal = ")"
		l.readRune()
	case '{':
		tok.Type = token.LBRACE
		tok.Literal = "{"
		l.readRune()
	case '}':
		tok.Type = token.RBRACE
		tok.Literal = "}"
		l.readRune()
	case '[':
		tok.Type = token.LBRACKET
		tok.Literal = "["
		l.readRune()
	case ']':
		tok.Type = token.RBRACKET
		tok.Literal = "]"
		l.readRune()
	case '.':
		tok.Type = token.DOT
		tok.Literal = "."
		l.readRune()
	case ':':
		tok.Type = token.COLON
		tok.Literal = ":"
		l.readRune()
	case '?':
		switch l.peekRune() {
		case ':':
			tok.Type = token.ELVIS
			tok.Literal = "?:"
			l.readRune()
			l.readRune()
		case '.':
			tok.Type = token.SAFE_DOT
			tok.Literal = "?."
			l.readRune()
			l.readRune()
		default:
			tok.Type = token.QUESTION
			tok.Literal = "?"
			l.readRune()
		}
	case '!':
		switch l.peekRune() {
		case '=':
			tok.Type = token.NOT_EQ
			tok.Literal = "!="
			l.readRune()
			l.readRune()
		case '!':
			tok.Type = token.NON_NULL
			tok.Literal = "!!"
			l.readRune()
			l.readRune()
		default:
			if l.peekWord("is") {
				tok.Type = token.NOT_IS
				tok.Literal = "!is"
				l.readRune()
				l.readRune() // consume 'i'
				l.readRune() // consume 's'
			} else {
				tok.Type = token.BANG
				tok.Literal = "!"
				l.readRune()
			}
		}
	case '=':
		switch l.peekRune() {
		case '=':
			tok.Type = token.EQ
			tok.Literal = "=="
			l.readRune()
			l.readRune()
		case '>':
			tok.Type = token.ARROW
			tok.Literal = "=>"
			l.readRune()
			l.readRune()
		default:
			tok.Type = token.ASSIGN
			tok.Literal = "="
			l.readRune()
		}
	case '<':
		if l.peekRune() == '=' {
			tok.Type = token.LTE
			tok.Literal = "<="
			l.readRune()
			l.readRune()
		} else {
			tok.Type = token.LT
			tok.Literal = "<"
			l.readRune()
		}
	case '>':
		if l.peekRune() == '=' {
			tok.Type = token.GTE
			tok.Literal = ">="
			l.readRune()
			l.readRune()
		} else {
			tok.Type = token.GT
			tok.Literal = ">"
			l.readRune()
		}
	case '&':
		if l.peekRune() == '&' {
			tok.Type = token.AND
			tok.Literal = "&&"
			l.readRune()
			l.readRune()
		} else {
			tok.Type = token.ILLEGAL
			tok.Literal = string(l.ch)
			l.readRune()
		}
	case '|':
		if l.peekRune() == '|' {
			tok.Type = token.OR
			tok.Literal = "||"
			l.readRune()
			l.readRune()
		} else {
			tok.Type = token.ILLEGAL
			tok.Literal = string(l.ch)
			l.readRune()
		}
	case '/':
		tok.Type = token.SLASH
		tok.Literal = "/"
		l.readRune()
	case '"':
		lit := l.readString()
		tok.Type = token.STRING
		tok.Literal = lit
	default:
		if isLetter(l.ch) {
			literal := l.readIdentifier()
			tok.Type = token.LookupIdent(literal)
			tok.Literal = literal
		} else if isDigit(l.ch) {
			literal := l.readNumber()
			tok.Type = token.NUMBER
			tok.Literal = literal
		} else {
			tok.Type = token.ILLEGAL
			tok.Literal = string(l.ch)
			l.readRune()
		}
	}

	tok.End = token.Position{Offset: l.position, Line: l.line, Column: l.column}
	return tok
}

func (l *Lexer) readRune() {
	if l.readPosition >= len(l.input) {
		l.position = len(l.input)
		l.ch = 0
		return
	}
	l.position = l.readPosition
	l.ch = l.input[l.readPosition]
	l.readPosition++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

func (l *Lexer) peekRune() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) peekWord(word string) bool {
	if len(word) == 0 {
		return false
	}
	if l.peekRune() != rune(word[0]) {
		return false
	}
	if l.readPosition+len(word) > len(l.input) {
		return false
	}
	for i, r := range word {
		if l.readPosition+i >= len(l.input) {
			return false
		}
		if l.input[l.readPosition+i] != r {
			return false
		}
	}
	nextIndex := l.readPosition + len(word)
	if nextIndex < len(l.input) {
		next := l.input[nextIndex]
		if isLetter(next) || isDigit(next) || next == '_' {
			return false
		}
	}
	return true
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readRune()
	}
	return string(l.input[start:l.position])
}

func (l *Lexer) readNumber() string {
	start := l.position
	for isDigit(l.ch) {
		l.readRune()
	}
	return string(l.input[start:l.position])
}

func (l *Lexer) readString() string {
	l.readRune() // consume opening quote
	start := l.position
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readRune()
		}
		l.readRune()
	}
	literal := string(l.input[start:l.position])
	if l.ch == '"' {
		l.readRune() // consume closing quote
	}
	return literal
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			l.readRune()
		}
		if l.ch == '/' {
			switch l.peekRune() {
			case '/':
				l.consumeLineComment()
				continue
			case '*':
				l.consumeBlockComment()
				continue
			}
		}
		break
	}
}

func (l *Lexer) consumeLineComment() {
	l.readRune() // consume first '/'
	l.readRune() // consume second '/'
	for l.ch != '\n' && l.ch != 0 {
		l.readRune()
	}
}

func (l *Lexer) consumeBlockComment() {
	l.readRune() // consume first '/'
	l.readRune() // consume '*'
	for {
		if l.ch == 0 {
			return
		}
		if l.ch == '*' && l.peekRune() == '/' {
			l.readRune()
			l.readRune()
			return
		}
		l.readRune()
	}
}

func isLetter(ch rune) bool {
	return ch == '_' || strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", ch)
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}
