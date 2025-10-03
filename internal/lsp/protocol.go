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
