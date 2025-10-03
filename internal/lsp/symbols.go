package lsp

import (
	"sort"
	"strings"

	"selenelang/internal/ast"
	"selenelang/internal/lexer"
	"selenelang/internal/parser"
)

func documentSymbols(source string) []DocumentSymbol {
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return nil
	}
	symbols := make([]DocumentSymbol, 0, len(program.Items))
	for _, item := range program.Items {
		if sym, ok := symbolFromItem(item); ok {
			symbols = append(symbols, sym)
		}
	}
	return symbols
}

func workspaceSymbols(docs map[string]*document, query string) []SymbolInformation {
	lower := strings.ToLower(query)
	infos := make([]SymbolInformation, 0)
	for uri, doc := range docs {
		ds := documentSymbols(doc.Text)
		flattened := flattenSymbols(uri, ds)
		for _, info := range flattened {
			if lower == "" || strings.Contains(strings.ToLower(info.Name), lower) {
				infos = append(infos, info)
			}
		}
	}
	sort.Slice(infos, func(i, j int) bool {
		if infos[i].Name == infos[j].Name {
			return infos[i].Location.URI < infos[j].Location.URI
		}
		return infos[i].Name < infos[j].Name
	})
	return infos
}

func flattenSymbols(uri string, symbols []DocumentSymbol) []SymbolInformation {
	out := make([]SymbolInformation, 0, len(symbols))
	for _, sym := range symbols {
		info := SymbolInformation{
			Name:     sym.Name,
			Kind:     sym.Kind,
			Detail:   sym.Detail,
			Location: Location{URI: uri, Range: sym.Range},
		}
		out = append(out, info)
		if len(sym.Children) > 0 {
			out = append(out, flattenSymbols(uri, sym.Children)...)
		}
	}
	return out
}

func symbolFromItem(item ast.ProgramItem) (DocumentSymbol, bool) {
	switch node := item.(type) {
	case *ast.PackageDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		return DocumentSymbol{
			Name:           node.Name.Name,
			Kind:           symbolKindPackage,
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}, true
	case *ast.ModuleDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		return DocumentSymbol{
			Name:           node.Name.Name,
			Kind:           symbolKindModule,
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}, true
	case *ast.FunctionDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		sym := DocumentSymbol{
			Name:           node.Name.Name,
			Kind:           symbolKindFunction,
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}
		return sym, true
	case *ast.VariableDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		detail := "let"
		if node.Mutable {
			detail = "var"
		}
		return DocumentSymbol{
			Name:           node.Name.Name,
			Detail:         detail,
			Kind:           symbolKindVariable,
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}, true
	case *ast.StructDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		return DocumentSymbol{
			Name:           node.Name.Name,
			Kind:           symbolKindClass,
			Detail:         "struct",
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}, true
	case *ast.ClassDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		return DocumentSymbol{
			Name:           node.Name.Name,
			Kind:           symbolKindClass,
			Detail:         "class",
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}, true
	case *ast.InterfaceDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		return DocumentSymbol{
			Name:           node.Name.Name,
			Kind:           symbolKindInterface,
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}, true
	case *ast.EnumDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		return DocumentSymbol{
			Name:           node.Name.Name,
			Kind:           symbolKindEnum,
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}, true
	case *ast.ContractDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		return DocumentSymbol{
			Name:           node.Name.Name,
			Kind:           symbolKindClass,
			Detail:         "contract",
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}, true
	default:
		return DocumentSymbol{}, false
	}
}

func rangeFromNode(node ast.Node) Range {
	start := positionFromTokenPos(node.Pos())
	end := positionFromTokenPos(node.End())
	return Range{Start: start, End: end}
}

func rangeFromIdentifier(id *ast.Identifier) Range {
	if id == nil {
		return Range{}
	}
	start := positionFromTokenPos(id.Pos())
	end := positionFromTokenPos(id.End())
	return Range{Start: start, End: end}
}
