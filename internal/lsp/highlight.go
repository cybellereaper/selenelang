package lsp

import (
	"math"
	"slices"
	"sort"
	"strings"

	"github.com/cybellereaper/selenelang/internal/token"
)

// Highlighter produces semantic tokens for editors.
type Highlighter struct {
	tokenTypes      []string
	tokenTypeLookup map[string]int
}

type semanticToken struct {
	line      int
	start     int
	length    int
	tokenType int
	modifiers int
}

const maxUint32 = uint64(math.MaxUint32)

func safeUint32(value int) (uint32, bool) {
	if value < 0 {
		return 0, false
	}
	if uint64(value) > maxUint32 {
		return 0, false
	}
	return uint32(value), true
}

// NewHighlighter constructs a semantic token highlighter with Selene token types.
func NewHighlighter() *Highlighter {
	types := []string{
		"namespace",
		"type",
		"class",
		"enum",
		"interface",
		"struct",
		"function",
		"method",
		"variable",
		"parameter",
		"property",
		"keyword",
		"number",
		"string",
		"operator",
		"comment",
		"enumMember",
	}
	lookup := make(map[string]int, len(types))
	for i, t := range types {
		lookup[t] = i
	}
	return &Highlighter{tokenTypes: types, tokenTypeLookup: lookup}
}

// Legend returns the supported semantic token types and modifiers.
func (h *Highlighter) Legend() (tokenTypes []string, tokenModifiers []string) {
	return slices.Clone(h.tokenTypes), []string{}
}

// Encode builds semantic tokens for an entire document snapshot.
func (h *Highlighter) Encode(doc *DocumentSnapshot) SemanticTokens {
	if doc == nil || doc.Symbols == nil {
		return SemanticTokens{}
	}
	segments := h.collectTokens(doc)
	return encodeSemanticSegments(segments)
}

// EncodeRange builds semantic tokens covering a specific range.
func (h *Highlighter) EncodeRange(doc *DocumentSnapshot, rng Range) SemanticTokens {
	if doc == nil || doc.Symbols == nil {
		return SemanticTokens{}
	}
	segments := h.collectTokens(doc)
	filtered := make([]semanticToken, 0, len(segments))
	for _, seg := range segments {
		segRange := Range{
			Start: Position{Line: seg.line, Character: seg.start},
			End:   Position{Line: seg.line, Character: seg.start + seg.length},
		}
		if rangesOverlap(segRange, rng) {
			filtered = append(filtered, seg)
		}
	}
	return encodeSemanticSegments(filtered)
}

func (h *Highlighter) collectTokens(doc *DocumentSnapshot) []semanticToken {
	lines := splitLines(doc.Text)
	symbols := doc.Symbols

	typeClassifications := make(map[string]string)
	for _, t := range symbols.TypeSymbols {
		typeClassifications[t.Name] = classificationForType(t)
	}
	functionNames := make(map[string]struct{})
	parameterNames := make(map[string]struct{})
	for _, fn := range symbols.FunctionSymbols {
		functionNames[fn.Name] = struct{}{}
		for _, param := range fn.Params {
			if param.Name != "" {
				parameterNames[param.Name] = struct{}{}
			}
		}
	}
	variableNames := make(map[string]struct{})
	for _, v := range symbols.VariableSymbols {
		if v.Name != "" {
			variableNames[v.Name] = struct{}{}
		}
	}

	segments := make([]semanticToken, 0, len(doc.Tokens))
	for _, tok := range doc.Tokens {
		if tok.Type == token.EOF {
			continue
		}
		rng := rangeFromToken(tok)
		switch tok.Type {
		case token.STRING, token.RAWSTRING, token.FORMATSTRING:
			segments = append(segments, h.makeSegments(lines, rng, h.indexFor("string"))...)
		case token.NUMBER:
			segments = append(segments, h.makeSegments(lines, rng, h.indexFor("number"))...)
		default:
			if isKeywordToken(tok.Type) {
				segments = append(segments, h.makeSegments(lines, rng, h.indexFor("keyword"))...)
				continue
			}
			if isOperatorToken(tok.Type) {
				segments = append(segments, h.makeSegments(lines, rng, h.indexFor("operator"))...)
				continue
			}
			if tok.Type == token.IDENT {
				classification := "variable"
				if class, ok := typeClassifications[tok.Literal]; ok {
					classification = class
				} else if _, ok := parameterNames[tok.Literal]; ok {
					classification = "parameter"
				} else if _, ok := functionNames[tok.Literal]; ok {
					classification = "function"
				} else if _, ok := variableNames[tok.Literal]; ok {
					classification = "variable"
				} else {
					classification = "variable"
				}
				segments = append(segments, h.makeSegments(lines, rng, h.indexFor(classification))...)
			}
		}
	}
	return segments
}

func (h *Highlighter) indexFor(name string) int {
	if idx, ok := h.tokenTypeLookup[name]; ok {
		return idx
	}
	return h.tokenTypeLookup["variable"]
}

func (h *Highlighter) makeSegments(lines [][]rune, rng Range, tokenType int) []semanticToken {
	if tokenType < 0 {
		return nil
	}
	if rng.Start.Line < 0 {
		rng.Start.Line = 0
	}
	if rng.End.Line < rng.Start.Line {
		rng.End.Line = rng.Start.Line
	}
	segments := make([]semanticToken, 0, 1)
	if rng.Start.Line >= len(lines) {
		length := rng.End.Character - rng.Start.Character
		if length <= 0 {
			length = 1
		}
		segments = append(segments, semanticToken{line: rng.Start.Line, start: rng.Start.Character, length: length, tokenType: tokenType})
		return segments
	}
	if rng.End.Line >= len(lines) {
		rng.End.Line = len(lines) - 1
	}
	if rng.Start.Line == rng.End.Line {
		lineLen := 0
		if rng.Start.Line < len(lines) {
			lineLen = len(lines[rng.Start.Line])
		}
		length := rng.End.Character - rng.Start.Character
		if length <= 0 {
			if lineLen > rng.Start.Character {
				length = lineLen - rng.Start.Character
			} else {
				length = 1
			}
		}
		if length <= 0 {
			length = 1
		}
		segments = append(segments, semanticToken{line: rng.Start.Line, start: rng.Start.Character, length: length, tokenType: tokenType})
		return segments
	}
	// First line
	firstLineLen := 0
	if rng.Start.Line < len(lines) {
		firstLineLen = len(lines[rng.Start.Line])
	}
	firstLength := firstLineLen - rng.Start.Character
	if firstLength <= 0 {
		firstLength = 1
	}
	segments = append(segments, semanticToken{line: rng.Start.Line, start: rng.Start.Character, length: firstLength, tokenType: tokenType})
	// Middle lines
	for line := rng.Start.Line + 1; line < rng.End.Line; line++ {
		if line < 0 || line >= len(lines) {
			continue
		}
		length := len(lines[line])
		if length <= 0 {
			continue
		}
		segments = append(segments, semanticToken{line: line, start: 0, length: length, tokenType: tokenType})
	}
	// Last line
	lastLength := rng.End.Character
	if lastLength <= 0 {
		lastLength = 1
	}
	segments = append(segments, semanticToken{line: rng.End.Line, start: 0, length: lastLength, tokenType: tokenType})
	return segments
}

func splitLines(text string) [][]rune {
	if text == "" {
		return [][]rune{{}}
	}
	rawLines := strings.Split(text, "\n")
	lines := make([][]rune, len(rawLines))
	for i, line := range rawLines {
		lines[i] = []rune(line)
	}
	return lines
}

func encodeSemanticSegments(segments []semanticToken) SemanticTokens {
	if len(segments) == 0 {
		return SemanticTokens{}
	}
	sort.Slice(segments, func(i, j int) bool {
		if segments[i].line == segments[j].line {
			return segments[i].start < segments[j].start
		}
		return segments[i].line < segments[j].line
	})
	data := make([]uint32, 0, len(segments)*5)
	prevLine := 0
	prevStart := 0
	first := true
	for _, seg := range segments {
		if seg.length <= 0 {
			continue
		}
		deltaLine := seg.line
		deltaStart := seg.start
		if !first {
			deltaLine = seg.line - prevLine
			if deltaLine == 0 {
				deltaStart = seg.start - prevStart
			}
		}
		if deltaLine < 0 || deltaStart < 0 || seg.length < 0 || seg.tokenType < 0 || seg.modifiers < 0 {
			continue
		}
		lineVal, ok := safeUint32(deltaLine)
		if !ok {
			continue
		}
		startVal, ok := safeUint32(deltaStart)
		if !ok {
			continue
		}
		lengthVal, ok := safeUint32(seg.length)
		if !ok {
			continue
		}
		typeVal, ok := safeUint32(seg.tokenType)
		if !ok {
			continue
		}
		modifierVal, ok := safeUint32(seg.modifiers)
		if !ok {
			continue
		}
		data = append(data, lineVal, startVal, lengthVal, typeVal, modifierVal)
		prevLine = seg.line
		prevStart = seg.start
		first = false
	}
	return SemanticTokens{Data: data}
}

func rangesOverlap(a, b Range) bool {
	if !rangeIsValid(a) || !rangeIsValid(b) {
		return false
	}
	if comparePosition(a.End, b.Start) <= 0 {
		return false
	}
	if comparePosition(b.End, a.Start) <= 0 {
		return false
	}
	return true
}

func classificationForType(t TypeSymbol) string {
	switch strings.ToLower(t.Detail) {
	case "class", "contract":
		return "class"
	case "struct":
		return "struct"
	case "enum":
		return "enum"
	case "interface":
		return "interface"
	default:
		return "type"
	}
}

func isKeywordToken(t token.Type) bool {
	switch t {
	case token.LET, token.VAR, token.FN, token.ASYNC, token.CONTRACT, token.RETURNS,
		token.CLASS, token.STRUCT, token.ENUM, token.MATCH, token.MODULE, token.IMPORT,
		token.AS, token.PACKAGE, token.INTERFACE, token.IF, token.ELSE, token.WHILE,
		token.FOR, token.RETURN, token.BREAK, token.CONTINUE, token.AWAIT, token.TRY,
		token.CATCH, token.FINALLY, token.THROW, token.USING, token.EXT, token.CONDITION,
		token.WHEN, token.TRUE, token.FALSE, token.NULL:
		return true
	default:
		return false
	}
}

func isOperatorToken(t token.Type) bool {
	switch t {
	case token.ASSIGN, token.PLUS, token.MINUS, token.ASTERISK, token.SLASH, token.PERCENT,
		token.PLUS_ASSIGN, token.MINUS_ASSIGN, token.STAR_ASSIGN, token.SLASH_ASSIGN, token.PERCENT_ASSIGN,
		token.BANG, token.QUESTION, token.COLON, token.ELVIS, token.SAFE_DOT, token.NON_NULL,
		token.AMPERSAND, token.EQ, token.NOT_EQ, token.LT, token.LTE, token.GT, token.GTE,
		token.OR, token.AND, token.ARROW, token.IS, token.NOT_IS:
		return true
	default:
		return false
	}
}
