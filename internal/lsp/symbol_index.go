package lsp

import (
	"fmt"
	"strings"

	"selenelang/internal/ast"
	"selenelang/internal/token"
)

// SymbolIndex stores aggregated symbol information for a document.
type SymbolIndex struct {
	DocumentSymbols []DocumentSymbol
	FunctionSymbols []FunctionSymbol
	TypeSymbols     []TypeSymbol
	VariableSymbols []VariableSymbol
}

// FunctionSymbol describes a function declaration discovered in the source.
type FunctionSymbol struct {
	Name           string
	Detail         string
	Range          Range
	SelectionRange Range
	BodyRange      Range
	Params         []ParameterSymbol
	HasBody        bool
	Async          bool
}

// ParameterSymbol captures metadata about a function parameter.
type ParameterSymbol struct {
	Name  string
	Range Range
}

// TypeSymbol represents a declared type such as a class or interface.
type TypeSymbol struct {
	Name   string
	Kind   int
	Detail string
	Range  Range
}

// VariableSymbol records information about a variable declaration.
type VariableSymbol struct {
	Name       string
	Range      Range
	Mutable    bool
	DeclLine   int
	DeclColumn int
}

// FunctionForPosition returns the innermost function covering the position.
func (i *SymbolIndex) FunctionForPosition(pos Position) *FunctionSymbol {
	if i == nil {
		return nil
	}
	for idx := range i.FunctionSymbols {
		fn := &i.FunctionSymbols[idx]
		if rangeContains(fn.BodyRange, pos) || rangeContains(fn.Range, pos) {
			return fn
		}
	}
	return nil
}

func buildSymbolIndex(program *ast.Program, tokens []token.Token) *SymbolIndex {
	index := &SymbolIndex{
		DocumentSymbols: make([]DocumentSymbol, 0),
		FunctionSymbols: make([]FunctionSymbol, 0),
		TypeSymbols:     make([]TypeSymbol, 0),
		VariableSymbols: make([]VariableSymbol, 0),
	}
	if program != nil {
		for _, item := range program.Items {
			if sym, ok := index.symbolFromItem(item); ok {
				index.DocumentSymbols = append(index.DocumentSymbols, sym)
			}
		}
	}
	index.VariableSymbols = append(index.VariableSymbols, extractVariableSymbols(tokens)...)
	return index
}

func (i *SymbolIndex) symbolFromItem(item ast.ProgramItem) (DocumentSymbol, bool) {
	switch node := item.(type) {
	case *ast.PackageDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		sym := DocumentSymbol{
			Name:           node.Name.Name,
			Kind:           symbolKindPackage,
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}
		return sym, true
	case *ast.ModuleDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		sym := DocumentSymbol{
			Name:           node.Name.Name,
			Kind:           symbolKindModule,
			Detail:         "module",
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}
		if node.Body != nil {
			children := i.symbolsFromStatements(node.Body.Statements)
			if len(children) > 0 {
				sym.Children = append(sym.Children, children...)
			}
		}
		return sym, true
	case *ast.FunctionDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		fnSym := FunctionSymbol{
			Name:           node.Name.Name,
			Detail:         functionSignature(node),
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
			BodyRange:      determineFunctionBodyRange(node),
			HasBody:        node.Body != nil || node.BodyExpr != nil,
			Async:          node.Async,
		}
		paramChildren := make([]DocumentSymbol, 0, len(node.Params))
		for _, param := range node.Params {
			if param.Name == nil {
				continue
			}
			ps := ParameterSymbol{Name: param.Name.Name, Range: rangeFromIdentifier(param.Name)}
			fnSym.Params = append(fnSym.Params, ps)
			paramChildren = append(paramChildren, DocumentSymbol{
				Name:           param.Name.Name,
				Detail:         "parameter",
				Kind:           symbolKindVariable,
				Range:          rangeFromIdentifier(param.Name),
				SelectionRange: rangeFromIdentifier(param.Name),
			})
		}
		i.FunctionSymbols = append(i.FunctionSymbols, fnSym)
		sym := DocumentSymbol{
			Name:           node.Name.Name,
			Detail:         fnSym.Detail,
			Kind:           symbolKindFunction,
			Range:          fnSym.Range,
			SelectionRange: fnSym.SelectionRange,
		}
		if len(paramChildren) > 0 {
			sym.Children = append(sym.Children, paramChildren...)
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
		sym := DocumentSymbol{
			Name:           node.Name.Name,
			Detail:         detail,
			Kind:           symbolKindVariable,
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}
		return sym, true
	case *ast.StructDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		sym := DocumentSymbol{
			Name:           node.Name.Name,
			Detail:         "struct",
			Kind:           symbolKindClass,
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}
		i.TypeSymbols = append(i.TypeSymbols, TypeSymbol{Name: node.Name.Name, Kind: symbolKindClass, Detail: "struct", Range: sym.Range})
		return sym, true
	case *ast.ClassDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		sym := DocumentSymbol{
			Name:           node.Name.Name,
			Detail:         "class",
			Kind:           symbolKindClass,
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}
		i.TypeSymbols = append(i.TypeSymbols, TypeSymbol{Name: node.Name.Name, Kind: symbolKindClass, Detail: "class", Range: sym.Range})
		return sym, true
	case *ast.InterfaceDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		sym := DocumentSymbol{
			Name:           node.Name.Name,
			Detail:         "interface",
			Kind:           symbolKindInterface,
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}
		i.TypeSymbols = append(i.TypeSymbols, TypeSymbol{Name: node.Name.Name, Kind: symbolKindInterface, Detail: "interface", Range: sym.Range})
		return sym, true
	case *ast.EnumDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		sym := DocumentSymbol{
			Name:           node.Name.Name,
			Detail:         "enum",
			Kind:           symbolKindEnum,
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}
		if len(node.Cases) > 0 {
			children := make([]DocumentSymbol, 0, len(node.Cases))
			for _, c := range node.Cases {
				if c.Name == nil {
					continue
				}
				children = append(children, DocumentSymbol{
					Name:           c.Name.Name,
					Detail:         "case",
					Kind:           symbolKindEnum,
					Range:          rangeFromIdentifier(c.Name),
					SelectionRange: rangeFromIdentifier(c.Name),
				})
			}
			sym.Children = append(sym.Children, children...)
		}
		i.TypeSymbols = append(i.TypeSymbols, TypeSymbol{Name: node.Name.Name, Kind: symbolKindEnum, Detail: "enum", Range: sym.Range})
		return sym, true
	case *ast.ContractDeclaration:
		if node.Name == nil {
			return DocumentSymbol{}, false
		}
		sym := DocumentSymbol{
			Name:           node.Name.Name,
			Detail:         "contract",
			Kind:           symbolKindClass,
			Range:          rangeFromNode(node),
			SelectionRange: rangeFromIdentifier(node.Name),
		}
		i.TypeSymbols = append(i.TypeSymbols, TypeSymbol{Name: node.Name.Name, Kind: symbolKindClass, Detail: "contract", Range: sym.Range})
		return sym, true
	default:
		return DocumentSymbol{}, false
	}
}

func (i *SymbolIndex) symbolsFromStatements(stmts []ast.Statement) []DocumentSymbol {
	if len(stmts) == 0 {
		return nil
	}
	symbols := make([]DocumentSymbol, 0)
	for _, stmt := range stmts {
		if stmt == nil {
			continue
		}
		if sym, ok := i.symbolFromStatement(stmt); ok {
			symbols = append(symbols, sym)
		}
	}
	return symbols
}

func (i *SymbolIndex) symbolFromStatement(stmt ast.Statement) (DocumentSymbol, bool) {
	switch node := stmt.(type) {
	case *ast.FunctionDeclaration:
		return i.symbolFromItem(node)
	case *ast.VariableDeclaration:
		return i.symbolFromItem(node)
	case *ast.ClassDeclaration:
		return i.symbolFromItem(node)
	case *ast.StructDeclaration:
		return i.symbolFromItem(node)
	case *ast.EnumDeclaration:
		return i.symbolFromItem(node)
	case *ast.InterfaceDeclaration:
		return i.symbolFromItem(node)
	case *ast.ContractDeclaration:
		return i.symbolFromItem(node)
	case *ast.BlockStatement:
		return DocumentSymbol{}, false
	default:
		return DocumentSymbol{}, false
	}
}

func extractVariableSymbols(tokens []token.Token) []VariableSymbol {
	vars := make([]VariableSymbol, 0)
	for i := 0; i < len(tokens)-1; i++ {
		tok := tokens[i]
		if tok.Type != token.LET && tok.Type != token.VAR {
			continue
		}
		next := tokens[i+1]
		if next.Type != token.IDENT {
			continue
		}
		rng := rangeFromToken(next)
		vars = append(vars, VariableSymbol{
			Name:       next.Literal,
			Range:      rng,
			Mutable:    tok.Type == token.VAR,
			DeclLine:   rng.Start.Line,
			DeclColumn: rng.Start.Character,
		})
	}
	return vars
}

func functionSignature(fn *ast.FunctionDeclaration) string {
	if fn == nil {
		return "fn"
	}
	params := make([]string, 0, len(fn.Params))
	for _, param := range fn.Params {
		if param.Name == nil {
			continue
		}
		name := param.Name.Name
		if param.Type != nil {
			name = fmt.Sprintf("%s: %s", name, formatTypeAnnotation(param.Type))
		}
		params = append(params, name)
	}
	prefix := "fn"
	if fn.Async {
		prefix = "async fn"
	}
	result := fmt.Sprintf("%s(%s)", prefix, strings.Join(params, ", "))
	if fn.ReturnType != nil {
		result = fmt.Sprintf("%s -> %s", result, formatTypeAnnotation(fn.ReturnType))
	}
	if fn.Receiver != nil && fn.Receiver.Name != nil {
		result = fmt.Sprintf("%s %s", fn.Receiver.Name.Name, result)
	}
	return result
}

func formatTypeAnnotation(t *ast.TypeAnnotation) string {
	if t == nil {
		return ""
	}
	parts := make([]string, 0)
	if t.Name != nil {
		parts = append(parts, t.Name.Name)
	}
	if len(t.TypeArgs) > 0 {
		args := make([]string, 0, len(t.TypeArgs))
		for _, arg := range t.TypeArgs {
			args = append(args, formatTypeAnnotation(arg))
		}
		parts = append(parts, fmt.Sprintf("<%s>", strings.Join(args, ", ")))
	}
	typeStr := strings.Join(parts, "")
	if typeStr == "" {
		typeStr = "type"
	}
	if t.Nullable {
		typeStr += "?"
	}
	return typeStr
}

func determineFunctionBodyRange(fn *ast.FunctionDeclaration) Range {
	if fn == nil {
		return Range{}
	}
	if fn.Body != nil {
		return rangeFromNode(fn.Body)
	}
	if fn.BodyExpr != nil {
		return rangeFromNode(fn.BodyExpr)
	}
	return rangeFromNode(fn)
}

func flattenDocumentSymbols(uri string, symbols []DocumentSymbol) []SymbolInformation {
	infos := make([]SymbolInformation, 0)
	for _, sym := range symbols {
		info := SymbolInformation{
			Name:     sym.Name,
			Kind:     sym.Kind,
			Detail:   sym.Detail,
			Location: Location{URI: uri, Range: sym.Range},
		}
		infos = append(infos, info)
		if len(sym.Children) > 0 {
			infos = append(infos, flattenDocumentSymbols(uri, sym.Children)...)
		}
	}
	return infos
}
