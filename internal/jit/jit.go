package jit

import (
	"fmt"

	"selenelang/internal/ast"
	"selenelang/internal/runtime"
)

// Program represents a Selene program compiled into a sequence of eagerly
// specialized evaluation closures. The implementation reuses the interpreter's
// semantics but avoids repeatedly performing large type switches at runtime.
type Program struct {
	steps    []compiledStep
	analysis runtime.MainAnalysis
}

type compiledStep struct {
	run func(*runtime.Environment) (runtime.Value, error)
}

// Compile converts a parsed Selene AST into a JIT program that can be executed
// efficiently against an existing runtime.
func Compile(program *ast.Program) (*Program, error) {
	if program == nil {
		return nil, fmt.Errorf("jit: program cannot be nil")
	}
	steps := make([]compiledStep, 0, len(program.Items))
	for _, item := range program.Items {
		steps = append(steps, compileProgramItem(item))
	}
	analysis := runtime.AnalyzeMain(program)
	return &Program{steps: steps, analysis: analysis}, nil
}

func compileProgramItem(item ast.ProgramItem) compiledStep {
	switch node := item.(type) {
	case ast.Statement:
		stmt := node
		return compiledStep{run: func(env *runtime.Environment) (runtime.Value, error) {
			return runtime.ExecuteStatement(stmt, env)
		}}
	case *ast.ModuleDeclaration:
		module := node
		return compiledStep{run: func(env *runtime.Environment) (runtime.Value, error) {
			return runtime.ExecuteModuleDeclaration(module, env)
		}}
	case *ast.PackageDeclaration:
		pkg := node
		return compiledStep{run: func(env *runtime.Environment) (runtime.Value, error) {
			if pkg.Name != nil {
				env.Set("__package__", runtime.NewString(pkg.Name.Name))
			}
			return runtime.NullValue, nil
		}}
	default:
		return compiledStep{run: func(env *runtime.Environment) (runtime.Value, error) {
			return runtime.ExecuteProgramItem(item, env)
		}}
	}
}

// Run executes the compiled program within the supplied runtime instance.
func (p *Program) Run(rt *runtime.Runtime) (runtime.Value, error) {
	if p == nil {
		return nil, fmt.Errorf("jit: program is nil")
	}
	if rt == nil {
		return nil, fmt.Errorf("jit: runtime is nil")
	}
	env := rt.Environment()
	var last runtime.Value = runtime.NullValue
	for _, step := range p.steps {
		val, err := step.run(env)
		if err != nil {
			return nil, err
		}
		if val != nil {
			last = val
		} else {
			last = runtime.NullValue
		}
	}
	return runtime.InvokeMainIfNeeded(env, p.analysis, last)
}

// RunWithEnvironment executes the program against a specific environment. This
// is primarily useful for embedding scenarios where callers manage their own
// runtime scopes.
func (p *Program) RunWithEnvironment(env *runtime.Environment) (runtime.Value, error) {
	if p == nil {
		return nil, fmt.Errorf("jit: program is nil")
	}
	if env == nil {
		return nil, fmt.Errorf("jit: environment is nil")
	}
	var last runtime.Value = runtime.NullValue
	for _, step := range p.steps {
		val, err := step.run(env)
		if err != nil {
			return nil, err
		}
		if val != nil {
			last = val
		} else {
			last = runtime.NullValue
		}
	}
	return runtime.InvokeMainIfNeeded(env, p.analysis, last)
}
