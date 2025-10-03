package lsp

var keywordItems = []CompletionItem{
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
	{Label: "true", Kind: completionItemKeyword, Detail: "boolean literal"},
	{Label: "false", Kind: completionItemKeyword, Detail: "boolean literal"},
	{Label: "null", Kind: completionItemKeyword, Detail: "null literal"},
}

var builtinItems = []CompletionItem{
	{Label: "print", Kind: completionItemFunction, Detail: "builtin"},
	{Label: "spawn", Kind: completionItemFunction, Detail: "builtin"},
	{Label: "channel", Kind: completionItemFunction, Detail: "builtin"},
}

func completionItems() []CompletionItem {
	items := make([]CompletionItem, 0, len(keywordItems)+len(builtinItems))
	items = append(items, keywordItems...)
	items = append(items, builtinItems...)
	return items
}
