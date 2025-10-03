package runtime

import "github.com/cybellereaper/selenelang/internal/ast"

// MainAnalysis captures metadata about a program's main function usage.
type MainAnalysis struct {
	HasMainFunction       bool
	HasTopLevelInvocation bool
}

// AnalyzeMain inspects the provided program and reports whether it defines a
// top-level main function and whether that function is already invoked via a
// top-level expression statement.
func AnalyzeMain(program *ast.Program) MainAnalysis {
	analysis := MainAnalysis{}
	if program == nil {
		return analysis
	}
	for _, item := range program.Items {
		switch node := item.(type) {
		case *ast.FunctionDeclaration:
			if node.Name != nil && node.Name.Name == "main" {
				analysis.HasMainFunction = true
			}
		case *ast.ExpressionStatement:
			if call, ok := node.Expression.(*ast.CallExpression); ok {
				if ident, ok := call.Callee.(*ast.Identifier); ok && ident.Name == "main" {
					analysis.HasTopLevelInvocation = true
				}
			}
		}
	}
	return analysis
}

// InvokeMainIfNeeded executes the runtime's main function when a program
// defines one but does not call it explicitly at the top level. The provided
// last value is returned unchanged when no invocation is required.
func InvokeMainIfNeeded(env *Environment, analysis MainAnalysis, last Value) (Value, error) {
	if !analysis.HasMainFunction || analysis.HasTopLevelInvocation || env == nil {
		return last, nil
	}
	candidate, ok := env.Get("main")
	if !ok {
		return last, nil
	}
	fn, ok := candidate.(*Function)
	if !ok {
		return last, nil
	}
	result, err := applyFunction(fn, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}
