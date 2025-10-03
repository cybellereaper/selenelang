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
	completionItemKeyword  = 14
	completionItemFunction = 3
	insertTextPlainText    = 1
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
