package lsp

// Position represents a location within a text document using zero-based indices.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a span of text within a document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Diagnostic describes a problem detected in a document.
type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity,omitempty"`
	Source   string `json:"source,omitempty"`
	Message  string `json:"message"`
}

const (
	severityError   = 1
	severityWarning = 2
)

// CompletionItem represents a single completion suggestion.
type CompletionItem struct {
	Label            string `json:"label"`
	Kind             int    `json:"kind,omitempty"`
	Detail           string `json:"detail,omitempty"`
	InsertText       string `json:"insertText,omitempty"`
	InsertTextFormat int    `json:"insertTextFormat,omitempty"`
}

// CompletionList is returned from completion requests.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

const (
	completionItemText          = 1
	completionItemMethod        = 2
	completionItemFunction      = 3
	completionItemConstructor   = 4
	completionItemField         = 5
	completionItemVariable      = 6
	completionItemClass         = 7
	completionItemInterface     = 8
	completionItemModule        = 9
	completionItemProperty      = 10
	completionItemUnit          = 11
	completionItemValue         = 12
	completionItemEnum          = 13
	completionItemKeyword       = 14
	completionItemSnippet       = 15
	completionItemColor         = 16
	completionItemFile          = 17
	completionItemReference     = 18
	completionItemFolder        = 19
	completionItemEnumMember    = 20
	completionItemConstant      = 21
	completionItemStruct        = 22
	completionItemEvent         = 23
	completionItemOperator      = 24
	completionItemTypeParameter = 25
)

const (
	insertTextPlainText = 1
	insertTextSnippet   = 2
)

// TextEdit represents a change applied to a text document.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// DocumentSymbol describes a named construct and its hierarchy within a document.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           int              `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// Location identifies a region of a document via URI and range.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// SymbolInformation represents a flattened symbol suitable for workspace searches.
type SymbolInformation struct {
	Name     string   `json:"name"`
	Kind     int      `json:"kind"`
	Location Location `json:"location"`
	Detail   string   `json:"detail,omitempty"`
}

// MarkupContent represents formatted hover text.
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// Hover contains hover information for a text position.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// SemanticTokens represents encoded semantic token data.
type SemanticTokens struct {
	Data []uint32 `json:"data"`
}

const (
	symbolKindFile        = 1
	symbolKindModule      = 2
	symbolKindNamespace   = 3
	symbolKindPackage     = 4
	symbolKindClass       = 5
	symbolKindMethod      = 6
	symbolKindProperty    = 7
	symbolKindField       = 8
	symbolKindConstructor = 9
	symbolKindEnum        = 10
	symbolKindInterface   = 11
	symbolKindFunction    = 12
	symbolKindVariable    = 13
	symbolKindConstant    = 14
	symbolKindString      = 15
	symbolKindNumber      = 16
	symbolKindBoolean     = 17
	symbolKindArray       = 18
)
