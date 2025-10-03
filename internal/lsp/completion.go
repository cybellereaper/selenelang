package lsp

import (
	"fmt"
	"sort"
	"strings"
)

// Completer provides completion suggestions for Selene source.
type Completer struct {
	keywordItems []CompletionItem
	builtinItems []CompletionItem
}

// NewCompleter builds a completer with keyword and builtin suggestions.
func NewCompleter() *Completer {
	keywords := []CompletionItem{
		{Label: "let", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "var", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "fn", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "async", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "contract", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "returns", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "class", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "struct", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "enum", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "match", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "module", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "import", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "as", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "package", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "interface", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "if", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "else", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "while", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "for", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "return", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "break", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "continue", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "await", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "try", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "catch", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "finally", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "throw", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "using", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "ext", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "condition", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "when", Kind: completionItemKeyword, Detail: "keyword"},
		{Label: "true", Kind: completionItemKeyword, Detail: "boolean"},
		{Label: "false", Kind: completionItemKeyword, Detail: "boolean"},
		{Label: "null", Kind: completionItemKeyword, Detail: "null"},
	}
	builtins := []CompletionItem{
		{Label: "print", Kind: completionItemFunction, Detail: "builtin"},
		{Label: "spawn", Kind: completionItemFunction, Detail: "builtin"},
		{Label: "channel", Kind: completionItemFunction, Detail: "builtin"},
	}
	return &Completer{keywordItems: keywords, builtinItems: builtins}
}

// Completion returns completion candidates for the provided snapshot and position.
func (c *Completer) Completion(doc *DocumentSnapshot, pos Position) CompletionList {
	if doc == nil || doc.Symbols == nil {
		return CompletionList{IsIncomplete: false, Items: nil}
	}
	prefix, start := identifierPrefixAt(doc.Text, pos)
	prev := runeBefore(doc.Text, start)
	lowerPrefix := strings.ToLower(prefix)
	suggestions := make([]CompletionItem, 0)
	seen := make(map[string]struct{})

	add := func(item CompletionItem) {
		if _, ok := seen[item.Label]; ok {
			return
		}
		if !prefixMatches(lowerPrefix, item.Label) {
			return
		}
		suggestions = append(suggestions, item)
		seen[item.Label] = struct{}{}
	}

	// Prioritize local variables and parameters.
	if fn := doc.Symbols.FunctionForPosition(pos); fn != nil {
		for _, param := range fn.Params {
			if param.Name == "" {
				continue
			}
			item := CompletionItem{
				Label:  param.Name,
				Kind:   completionItemVariable,
				Detail: "parameter",
			}
			add(item)
		}
	}

	for _, variable := range doc.Symbols.VariableSymbols {
		declPos := Position{Line: variable.DeclLine, Character: variable.DeclColumn}
		if comparePosition(declPos, pos) > 0 {
			continue
		}
		detail := "variable"
		if variable.Mutable {
			detail = "mutable variable"
		}
		item := CompletionItem{Label: variable.Name, Kind: completionItemVariable, Detail: detail}
		add(item)
	}

	for _, fn := range doc.Symbols.FunctionSymbols {
		snippet := fn.Name
		insertFormat := insertTextPlainText
		if len(fn.Params) > 0 {
			parts := make([]string, 0, len(fn.Params))
			for i, param := range fn.Params {
				placeholder := fmt.Sprintf("${%d:%s}", i+1, param.Name)
				parts = append(parts, placeholder)
			}
			snippet = fmt.Sprintf("%s(%s)", fn.Name, strings.Join(parts, ", "))
			insertFormat = insertTextSnippet
		} else {
			snippet = fmt.Sprintf("%s()", fn.Name)
		}
		item := CompletionItem{
			Label:            fn.Name,
			Kind:             completionItemFunction,
			Detail:           fn.Detail,
			InsertText:       snippet,
			InsertTextFormat: insertFormat,
		}
		add(item)
	}

	for _, t := range doc.Symbols.TypeSymbols {
		kind := completionItemClass
		switch strings.ToLower(t.Detail) {
		case "interface":
			kind = completionItemInterface
		case "enum":
			kind = completionItemEnum
		case "struct":
			kind = completionItemStruct
		}
		item := CompletionItem{Label: t.Name, Kind: kind, Detail: t.Detail}
		add(item)
	}

	if prev != '.' {
		for _, item := range c.keywordItems {
			add(item)
		}
	}
	for _, item := range c.builtinItems {
		add(item)
	}

	sort.SliceStable(suggestions, func(i, j int) bool {
		if strings.HasPrefix(strings.ToLower(suggestions[i].Label), lowerPrefix) &&
			!strings.HasPrefix(strings.ToLower(suggestions[j].Label), lowerPrefix) {
			return true
		}
		if len(suggestions[i].Label) == len(suggestions[j].Label) {
			return suggestions[i].Label < suggestions[j].Label
		}
		return len(suggestions[i].Label) < len(suggestions[j].Label)
	})

	return CompletionList{IsIncomplete: false, Items: suggestions}
}

func prefixMatches(prefix string, candidate string) bool {
	if prefix == "" {
		return true
	}
	return strings.HasPrefix(strings.ToLower(candidate), prefix)
}
