package jit

import (
	"testing"

	"github.com/cybellereaper/selenelang/internal/lexer"
	"github.com/cybellereaper/selenelang/internal/parser"
	"github.com/cybellereaper/selenelang/internal/runtime"
)

func TestJITRunSimpleProgram(t *testing.T) {
	source := `
let x = 1;
let y = 2;
x + y;
`
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}

	rt := runtime.New()
	compiled, err := Compile(program)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	value, err := compiled.Run(rt)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	num, ok := value.(*runtime.Number)
	if !ok {
		t.Fatalf("expected Number result, got %T", value)
	}
	if num.Value != 3 {
		t.Fatalf("expected 3, got %v", num.Value)
	}
}

func TestJITRunInvokesMainAutomatically(t *testing.T) {
	source := `
fn main() {
    record();
}
`
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}

	rt := runtime.New()
	var calls int
	rt.Environment().Set("record", runtime.NewBuiltin("record", func(args []runtime.Value) (runtime.Value, error) {
		calls++
		return runtime.NullValue, nil
	}))

	compiled, err := Compile(program)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if _, err := compiled.Run(rt); err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected record to be called once, got %d", calls)
	}
}
