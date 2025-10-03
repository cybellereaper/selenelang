package lsp

import "testing"

func TestCompletionSuggestsVariablesFunctionsAndParameters(t *testing.T) {
	source := "fn greet(name) {\n    let greeting = \"hi\"\n    gre\n}\n"
	analyzer := NewAnalyzer(NewLinter())
	docs := NewDocumentStore(analyzer)
	snapshot := docs.Open("file:///test.sel", 1, source)
	completer := NewCompleter()
	posWithPrefix := Position{Line: 2, Character: 7}
	if snapshot.Symbols.FunctionForPosition(posWithPrefix) == nil {
		t.Fatalf("expected to find enclosing function; functions: %#v", snapshot.Symbols.FunctionSymbols)
	}
	list := completer.Completion(snapshot, posWithPrefix)
	if len(list.Items) == 0 {
		t.Fatalf("expected completion items, got none")
	}
	if !hasCompletion(list, "greeting") {
		t.Fatalf("expected greeting variable in completion list: %#v", list.Items)
	}
	if !hasCompletion(list, "greet") {
		t.Fatalf("expected greet function in completion list: %#v", list.Items)
	}

	posNoPrefix := Position{Line: 2, Character: 4}
	paramList := completer.Completion(snapshot, posNoPrefix)
	if !hasCompletion(paramList, "name") {
		t.Fatalf("expected parameter completion for name: %#v", paramList.Items)
	}
}

func hasCompletion(list CompletionList, label string) bool {
	for _, item := range list.Items {
		if item.Label == label {
			return true
		}
	}
	return false
}
