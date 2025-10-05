// Package runtime implements the Selene interpreter and execution model.
package runtime

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/cybellereaper/selenelang/internal/ast"
	"github.com/cybellereaper/selenelang/internal/lexer"
	"github.com/cybellereaper/selenelang/internal/parser"
	"github.com/cybellereaper/selenelang/internal/token"
)

// Value is implemented by every runtime value and exposes its type and display form.
type Value interface {
	Type() string
	Inspect() string
}

var extensionRegistry = make(map[string]map[string]*Function)

func registerExtension(typeName, method string, fn *Function) {
	name := normalizeTypeName(typeName)
	if _, ok := extensionRegistry[name]; !ok {
		extensionRegistry[name] = make(map[string]*Function)
	}
	extensionRegistry[name][method] = fn
}

func lookupExtension(typeName, method string) (*Function, bool) {
	name := normalizeTypeName(typeName)
	if methods, ok := extensionRegistry[name]; ok {
		fn, ok := methods[method]
		return fn, ok
	}
	return nil, false
}

func normalizeTypeName(name string) string {
	switch name {
	case "Int", "Integer", "Float", "Number":
		return "Number"
	case "Bool", "Boolean":
		return "Boolean"
	case "String":
		return "String"
	default:
		return name
	}
}

// Number wraps a numeric Selene runtime value.
type Number struct {
	Value float64
}

// Type implements the Value interface for Number.
func (n *Number) Type() string { return "Number" }

// Inspect returns a human-readable representation of Number.
func (n *Number) Inspect() string { return strconv.FormatFloat(n.Value, 'f', -1, 64) }

// String wraps a UTF-8 Selene string value.
type String struct {
	Value string
}

// Type implements the Value interface for String.
func (s *String) Type() string { return "String" }

// Inspect returns a human-readable representation of String.
func (s *String) Inspect() string { return s.Value }

// Boolean wraps a true or false runtime value.
type Boolean struct {
	Value bool
}

// Type implements the Value interface for Boolean.
func (b *Boolean) Type() string { return "Boolean" }

// Inspect returns a human-readable representation of Boolean.
func (b *Boolean) Inspect() string {
	if b.Value {
		return "true"
	}
	return "false"
}

// Null represents the Selene null literal.
type Null struct{}

// Type implements the Value interface for Null.
func (n *Null) Type() string { return "Null" }

// Inspect returns a human-readable representation of Null.
func (n *Null) Inspect() string { return "null" }

var NullValue Value = &Null{}
var TrueValue Value = &Boolean{Value: true}
var FalseValue Value = &Boolean{Value: false}

// ErrorValue captures runtime errors with optional causes.
type ErrorValue struct {
	Message string
	Cause   Value
}

// Type implements the Value interface for ErrorValue.
func (e *ErrorValue) Type() string { return "Error" }

// Inspect returns a human-readable representation of ErrorValue.
func (e *ErrorValue) Inspect() string {
	b := borrowBuilder()
	b.WriteString("<error ")
	b.WriteString(e.Message)
	if e.Cause != nil {
		b.WriteString(" : ")
		b.WriteString(e.Cause.Inspect())
	}
	b.WriteByte('>')
	return finishBuilder(b)
}

func toErrorValue(val Value) *ErrorValue {
	switch v := val.(type) {
	case *ErrorValue:
		return v
	case *String:
		return &ErrorValue{Message: v.Value}
	default:
		return &ErrorValue{Message: v.Inspect(), Cause: val}
	}
}

// Array represents an ordered collection of values.
type Array struct {
	Elements []Value
}

// Type implements the Value interface for Array.
func (a *Array) Type() string { return "Array" }

// Inspect returns a human-readable representation of Array.
func (a *Array) Inspect() string {
	if len(a.Elements) == 0 {
		return "[]"
	}

	b := borrowBuilder()
	b.Grow(1 + len(a.Elements)*4)
	b.WriteByte('[')
	for i, el := range a.Elements {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(el.Inspect())
	}
	b.WriteByte(']')
	return finishBuilder(b)
}

// Object models a dynamic map of string keys to values.
type Object struct {
	Properties map[string]Value
}

// Type implements the Value interface for Object.
func (o *Object) Type() string { return "Object" }

// Inspect returns a human-readable representation of Object.
func (o *Object) Inspect() string {
	return inspectFields("", o.Properties)
}

// Module captures exported bindings from a loaded module.
type Module struct {
	Name    string
	Exports map[string]Value
}

// NewModule constructs a module value with the provided exports.
func NewModule(name string, exports map[string]Value) *Module {
	clone := maps.Clone(exports)
	if clone == nil {
		clone = make(map[string]Value)
	}
	return &Module{Name: name, Exports: clone}
}

// Type implements the Value interface for Module.
func (m *Module) Type() string { return "Module" }

// Inspect returns a human-readable representation of Module.
func (m *Module) Inspect() string {
	keys := sortedKeys(m.Exports)

	b := borrowBuilder()
	b.Grow(len(m.Name) + len(keys)*4 + 10)
	b.WriteString("<module ")
	b.WriteString(m.Name)
	b.WriteString(" [")
	for i, key := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(key)
	}
	b.WriteString("]>")
	return finishBuilder(b)
}

// StructType describes the shape of a user-defined struct.
type StructType struct {
	Name    string
	Fields  []string
	Methods map[string]*Function
	Static  map[string]Value
}

// Type implements the Value interface for StructType.
func (s *StructType) Type() string { return "Struct" }

// Inspect returns a human-readable representation of StructType.
func (s *StructType) Inspect() string {
	b := borrowBuilder()
	b.WriteString("<struct ")
	b.WriteString(s.Name)
	b.WriteByte('>')
	return finishBuilder(b)
}

// StructInstance stores field values for a struct.
type StructInstance struct {
	Definition *StructType
	Fields     map[string]Value
}

// Type implements the Value interface for StructInstance.
func (s *StructInstance) Type() string { return s.Definition.Name }

// Inspect returns a human-readable representation of StructInstance.
func (s *StructInstance) Inspect() string {
	return inspectFields(s.Definition.Name, s.Fields)
}

// ClassType represents the metadata for a Selene class.
type ClassType struct {
	Name    string
	Fields  []string
	Methods map[string]*Function
	Static  map[string]Value
	Super   *ClassType
}

// Type implements the Value interface for ClassType.
func (c *ClassType) Type() string { return "Class" }

// Inspect returns a human-readable representation of ClassType.
func (c *ClassType) Inspect() string {
	b := borrowBuilder()
	b.WriteString("<class ")
	b.WriteString(c.Name)
	b.WriteByte('>')
	return finishBuilder(b)
}

// ClassInstance stores fields and methods for a class instance.
type ClassInstance struct {
	Definition *ClassType
	Fields     map[string]Value
}

// Type implements the Value interface for ClassInstance.
func (c *ClassInstance) Type() string { return c.Definition.Name }

// Inspect returns a human-readable representation of ClassInstance.
func (c *ClassInstance) Inspect() string {
	return inspectFields(c.Definition.Name, c.Fields)
}

func (c *ClassType) lookupMethod(name string) (*Function, bool) {
	if c == nil {
		return nil, false
	}
	if fn, ok := c.Methods[name]; ok {
		return fn, true
	}
	if c.Super != nil {
		return c.Super.lookupMethod(name)
	}
	return nil, false
}

func (c *ClassType) lookupStatic(name string) (Value, bool) {
	if c == nil {
		return nil, false
	}
	if val, ok := c.Static[name]; ok {
		return val, true
	}
	if c.Super != nil {
		return c.Super.lookupStatic(name)
	}
	return nil, false
}

// EnumType describes an enumeration and its cases.
type EnumType struct {
	Name         string
	Cases        map[string][]string
	Constructors map[string]Value
}

// Type implements the Value interface for EnumType.
func (e *EnumType) Type() string { return "Enum" }

// Inspect returns a human-readable representation of EnumType.
func (e *EnumType) Inspect() string {
	keys := sortedKeys(e.Cases)

	b := borrowBuilder()
	b.Grow(len(e.Name) + len(keys)*4 + 10)
	b.WriteString("<enum ")
	b.WriteString(e.Name)
	b.WriteString(" [")
	for i, key := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(key)
	}
	b.WriteString("]>")
	return finishBuilder(b)
}

// EnumInstance stores the active case of an enumeration.
type EnumInstance struct {
	Enum   *EnumType
	Case   string
	Fields map[string]Value
	Order  []string
}

// Type implements the Value interface for EnumInstance.
func (e *EnumInstance) Type() string { return e.Enum.Name }

// Inspect returns a human-readable representation of EnumInstance.
func (e *EnumInstance) Inspect() string {
	if len(e.Order) == 0 {
		return e.Enum.Name + "." + e.Case
	}

	b := borrowBuilder()
	b.Grow(len(e.Enum.Name) + len(e.Case) + len(e.Order)*4 + 4)
	b.WriteString(e.Enum.Name)
	b.WriteByte('.')
	b.WriteString(e.Case)
	b.WriteByte('(')
	for i, name := range e.Order {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(e.Fields[name].Inspect())
	}
	b.WriteByte(')')
	return finishBuilder(b)
}

// ChannelValue represents a concurrent communication channel.
type ChannelValue struct {
	ch       chan Value
	capacity int
	closed   bool
}

// Type implements the Value interface for ChannelValue.
func (c *ChannelValue) Type() string { return "Channel" }

// Inspect returns a human-readable representation of ChannelValue.
func (c *ChannelValue) Inspect() string {
	state := "open"
	if c.closed {
		state = "closed"
	}

	b := borrowBuilder()
	b.WriteString("<channel ")
	b.WriteString(state)
	b.WriteString(" capacity=")
	b.WriteString(strconv.Itoa(c.capacity))
	b.WriteByte('>')
	return finishBuilder(b)
}

func (c *ChannelValue) send(val Value) error {
	if c.closed {
		return errors.New("send on closed channel")
	}
	c.ch <- val
	return nil
}

func (c *ChannelValue) recv() (Value, error) {
	v, ok := <-c.ch
	if !ok {
		return NullValue, errors.New("receive on closed channel")
	}
	return v, nil
}

func (c *ChannelValue) close() error {
	if c.closed {
		return nil
	}
	close(c.ch)
	c.closed = true
	return nil
}

type taskResult struct {
	value Value
	err   error
}

// Task tracks asynchronous execution state.
type Task struct {
	once   sync.Once
	result taskResult
	ch     chan taskResult
}

// NewTask creates a pending task with synchronization primitives.
func NewTask() *Task {
	return &Task{ch: make(chan taskResult, 1)}
}

// Type implements the Value interface for Task.
func (t *Task) Type() string { return "Task" }

// Inspect returns a human-readable representation of Task.
func (t *Task) Inspect() string { return "<task>" }

func (t *Task) deliver(val Value, err error) {
	t.ch <- taskResult{value: val, err: err}
	close(t.ch)
}

func (t *Task) await() taskResult {
	t.once.Do(func() {
		res, ok := <-t.ch
		if ok {
			t.result = res
			return
		}
		t.result = taskResult{value: NullValue, err: nil}
	})
	return t.result
}

// Join waits for the task to complete and returns its result.
func (t *Task) Join() (Value, error) {
	res := t.await()
	return res.value, res.err
}

// Contract captures runtime contract requirements.
type Contract struct {
	Name    string
	Exports map[string]Value
}

// Type implements the Value interface for Contract.
func (c *Contract) Type() string { return "Contract" }

// Inspect returns a human-readable representation of Contract.
func (c *Contract) Inspect() string {
	keys := sortedKeys(c.Exports)
	b := borrowBuilder()
	b.Grow(len(c.Name) + len(keys)*4 + 12)
	b.WriteString("<contract ")
	b.WriteString(c.Name)
	b.WriteString(" [")
	for i, key := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(key)
	}
	b.WriteString("]>")
	return finishBuilder(b)
}

// InterfaceType records an interface name and required methods.
type InterfaceType struct {
	Name    string
	Methods map[string]int
}

// Type implements the Value interface for InterfaceType.
func (i *InterfaceType) Type() string { return "Interface" }

// Inspect returns a human-readable representation of InterfaceType.
func (i *InterfaceType) Inspect() string {
	keys := sortedKeys(i.Methods)
	b := borrowBuilder()
	b.Grow(len(i.Name) + len(keys)*4 + 12)
	b.WriteString("<interface ")
	b.WriteString(i.Name)
	b.WriteString(" [")
	for j, key := range keys {
		if j > 0 {
			b.WriteString(", ")
		}
		b.WriteString(key)
	}
	b.WriteString("]>")
	return finishBuilder(b)
}

// BuiltinFunction is a Go function exposed as a Selene builtin.
type BuiltinFunction func(args []Value) (Value, error)

// Function wraps a Selene function declaration and its environment.
type Function struct {
	Declaration *ast.FunctionDeclaration
	Env         *Environment
	Builtin     BuiltinFunction
	Name        string
}

// Type implements the Value interface for Function.
func (f *Function) Type() string { return "Function" }

// Inspect returns a human-readable representation of Function.
func (f *Function) Inspect() string {
	if f.Name != "" {
		b := borrowBuilder()
		b.WriteString("<fn ")
		b.WriteString(f.Name)
		b.WriteByte('>')
		return finishBuilder(b)
	}
	return "<fn>"
}

// NewBoolean wraps a Go bool as a Selene boolean value.
func NewBoolean(v bool) Value {
	if v {
		return TrueValue
	}
	return FalseValue
}

// NewNumber wraps a Go float64 as a Selene number.
func NewNumber(v float64) Value { return &Number{Value: v} }

// NewString wraps a Go string as a Selene string.
func NewString(v string) Value { return &String{Value: v} }

// NewBuiltin creates a runtime value for a builtin function.
func NewBuiltin(name string, fn BuiltinFunction) Value {
	return &Function{Name: name, Builtin: fn}
}

type returnSignal struct {
	value Value
}

// Error implements the error interface for return signals.
func (r *returnSignal) Error() string { return "return" }

type breakSignal struct{}

// Error implements the error interface for break signals.
func (b *breakSignal) Error() string { return "break" }

type continueSignal struct{}

// Error implements the error interface for continue signals.
func (c *continueSignal) Error() string { return "continue" }

type runtimeError struct {
	value *ErrorValue
}

// Error exposes the error message stored in the runtime error.
func (r *runtimeError) Error() string { return r.value.Message }

func wrapRuntimeError(err error) *runtimeError {
	if err == nil {
		return nil
	}
	if rt, ok := err.(*runtimeError); ok {
		return rt
	}
	return &runtimeError{value: &ErrorValue{Message: err.Error()}}
}

// Environment stores variable bindings with optional outer scopes.
type Environment struct {
	store map[string]Value
	outer *Environment
}

// NewEnvironment creates a fresh environment with no outer scope.
func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]Value)}
}

// NewEnclosedEnvironment creates a child scope linked to an outer environment.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get looks up a value in the environment chain.
func (e *Environment) Get(name string) (Value, bool) {
	for env := e; env != nil; env = env.outer {
		if val, ok := env.store[name]; ok {
			return val, true
		}
	}
	return nil, false
}

// Set stores a binding in the current environment scope.
func (e *Environment) Set(name string, val Value) Value {
	e.store[name] = val
	return val
}

// Snapshot returns a copy of the environment bindings.
func (e *Environment) Snapshot() map[string]Value {
	if len(e.store) == 0 {
		return make(map[string]Value)
	}
	return maps.Clone(e.store)
}

// Assign updates an existing binding in the environment chain.
func (e *Environment) Assign(name string, val Value) (Value, error) {
	for env := e; env != nil; env = env.outer {
		if _, ok := env.store[name]; ok {
			env.store[name] = val
			return val, nil
		}
	}
	return nil, fmt.Errorf("undefined variable %s", name)
}

func (e *Environment) resolve(name string) (*Environment, bool) {
	for env := e; env != nil; env = env.outer {
		if _, ok := env.store[name]; ok {
			return env, true
		}
	}
	return nil, false
}

// Pointer references a binding inside an environment chain.
type Pointer struct {
	env  *Environment
	name string
}

// Type implements the Value interface for Pointer.
func (p *Pointer) Type() string { return "Pointer" }

// Inspect returns a human-readable representation of Pointer.
func (p *Pointer) Inspect() string {
	b := borrowBuilder()
	b.WriteByte('&')
	b.WriteString(p.name)
	return finishBuilder(b)
}

func (p *Pointer) get() (Value, error) {
	if p == nil || p.env == nil {
		return nil, errors.New("dereference of nil pointer")
	}
	if val, ok := p.env.store[p.name]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("dangling pointer to %s", p.name)
}

func (p *Pointer) set(val Value) error {
	if p == nil || p.env == nil {
		return errors.New("assignment to nil pointer")
	}
	if _, ok := p.env.store[p.name]; !ok {
		return fmt.Errorf("dangling pointer to %s", p.name)
	}
	p.env.store[p.name] = val
	return nil
}

// Runtime executes Selene programs and holds the global environment.
type Runtime struct {
	env *Environment
}

// New constructs a runtime with built-in functions installed.
func New() *Runtime {
	env := NewEnvironment()
	env.Set("print", NewBuiltin("print", builtinPrint))
	env.Set("format", NewBuiltin("format", builtinFormat))
	env.Set("spawn", NewBuiltin("spawn", builtinSpawn))
	env.Set("channel", NewBuiltin("channel", builtinChannel))
	return &Runtime{env: env}
}

// Environment returns the runtime's global environment.
func (r *Runtime) Environment() *Environment {
	return r.env
}

// Run executes a Selene program and returns the last value.
func (r *Runtime) Run(program *ast.Program) (Value, error) {
	analysis := AnalyzeMain(program)
	result, err := evalProgram(program, r.env)
	if err != nil {
		return nil, err
	}
	return InvokeMainIfNeeded(r.env, analysis, result)
}

func evalProgram(program *ast.Program, env *Environment) (Value, error) {
	result := NullValue
	for _, item := range program.Items {
		val, err := evalProgramItem(item, env)
		if err != nil {
			return nil, err
		}
		result = val
	}
	return result, nil
}

func evalProgramItem(item ast.ProgramItem, env *Environment) (Value, error) {
	switch node := item.(type) {
	case ast.Statement:
		val, err := evalStatement(node, env)
		if err != nil {
			if _, ok := err.(*returnSignal); ok {
				return nil, errors.New("return outside of function")
			}
			if _, ok := err.(*breakSignal); ok {
				return nil, errors.New("break outside of loop")
			}
			if _, ok := err.(*continueSignal); ok {
				return nil, errors.New("continue outside of loop")
			}
			return nil, err
		}
		return val, nil
	case *ast.ModuleDeclaration:
		return evalModuleDeclaration(node, env)
	case *ast.PackageDeclaration:
		if node.Name != nil {
			env.Set("__package__", NewString(node.Name.Name))
		}
		return NullValue, nil
	default:
		return nil, fmt.Errorf("runtime does not support top-level %T", item)
	}
}

// ExecuteProgramItem evaluates a top-level program item within the provided environment.
// It is a thin wrapper around the interpreter's internal evalProgramItem helper so that
// other packages (notably the JIT backend) can reuse the existing semantics without
// re-implementing the large statement/type switch.
func ExecuteProgramItem(item ast.ProgramItem, env *Environment) (Value, error) {
	return evalProgramItem(item, env)
}

func evalStatement(stmt ast.Statement, env *Environment) (Value, error) {
	switch node := stmt.(type) {
	case *ast.ExpressionStatement:
		if node.Expression == nil {
			return NullValue, nil
		}
		return evalExpression(node.Expression, env)
	case *ast.VariableDeclaration:
		var val Value = NullValue
		var err error
		if node.Value != nil {
			val, err = evalExpression(node.Value, env)
			if err != nil {
				return nil, err
			}
		}
		env.Set(node.Name.Name, val)
		return val, nil
	case *ast.FunctionDeclaration:
		fn := &Function{Declaration: node, Env: env, Name: node.Name.Name}
		if node.IsExtension {
			if node.Receiver == nil || node.Receiver.Name == nil {
				return nil, errors.New("extension function requires receiver type")
			}
			registerExtension(node.Receiver.Name.Name, node.Name.Name, fn)
			return fn, nil
		}
		env.Set(node.Name.Name, fn)
		return fn, nil
	case *ast.InterfaceDeclaration:
		methods := make(map[string]int)
		for _, method := range node.Methods {
			if method.Name != nil {
				methods[method.Name.Name] = len(method.Params)
			}
		}
		iface := &InterfaceType{Name: node.Name.Name, Methods: methods}
		env.Set(node.Name.Name, iface)
		return iface, nil
	case *ast.ImportDeclaration:
		return evalImportDeclaration(node, env)
	case *ast.StructDeclaration:
		return evalStructDeclaration(node, env)
	case *ast.ClassDeclaration:
		return evalClassDeclaration(node, env)
	case *ast.EnumDeclaration:
		return evalEnumDeclaration(node, env)
	case *ast.ContractDeclaration:
		return evalContractDeclaration(node, env)
	case *ast.BlockStatement:
		blockEnv := NewEnclosedEnvironment(env)
		return evalBlock(node, blockEnv)
	case *ast.MatchStatement:
		return evalMatchStatement(node, env)
	case *ast.IfStatement:
		return evalIfStatement(node, env)
	case *ast.WhileStatement:
		return evalWhileStatement(node, env)
	case *ast.ForStatement:
		return evalForStatement(node, env)
	case *ast.UsingStatement:
		return evalUsingStatement(node, env)
	case *ast.TryStatement:
		return evalTryStatement(node, env)
	case *ast.ThrowStatement:
		val, err := evalExpression(node.Value, env)
		if err != nil {
			return nil, err
		}
		return NullValue, &runtimeError{value: toErrorValue(val)}
	case *ast.ConditionStatement:
		return evalConditionStatement(node, env)
	case *ast.ReturnStatement:
		var val Value = NullValue
		var err error
		if node.Value != nil {
			val, err = evalExpression(node.Value, env)
			if err != nil {
				return nil, err
			}
		}
		return val, &returnSignal{value: val}
	case *ast.BreakStatement:
		return NullValue, &breakSignal{}
	case *ast.ContinueStatement:
		return NullValue, &continueSignal{}
	default:
		return nil, fmt.Errorf("runtime does not support statement %T", stmt)
	}
}

// ExecuteStatement evaluates a statement node in the provided environment.
func ExecuteStatement(stmt ast.Statement, env *Environment) (Value, error) {
	return evalStatement(stmt, env)
}

func evalBlock(block *ast.BlockStatement, env *Environment) (Value, error) {
	result := NullValue
	for _, stmt := range block.Statements {
		val, err := evalStatement(stmt, env)
		if err != nil {
			switch sig := err.(type) {
			case *returnSignal:
				return sig.value, err
			case *breakSignal, *continueSignal:
				return NullValue, err
			default:
				return nil, err
			}
		}
		result = val
	}
	return result, nil
}

func evalIfStatement(stmt *ast.IfStatement, env *Environment) (Value, error) {
	if stmt.Condition == nil {
		return NullValue, nil
	}
	condition, err := evalExpression(stmt.Condition, env)
	if err != nil {
		return nil, err
	}
	if isTruthy(condition) {
		if stmt.Consequence == nil {
			return NullValue, nil
		}
		return evalStatement(stmt.Consequence, env)
	}
	if stmt.Alternative != nil {
		return evalStatement(stmt.Alternative, env)
	}
	return NullValue, nil
}

func evalWhileStatement(stmt *ast.WhileStatement, env *Environment) (Value, error) {
	result := NullValue
	for {
		if stmt.Condition != nil {
			cond, err := evalExpression(stmt.Condition, env)
			if err != nil {
				return nil, err
			}
			if !isTruthy(cond) {
				break
			}
		}
		if stmt.Body == nil {
			continue
		}
		val, err := evalStatement(stmt.Body, env)
		if err != nil {
			switch sig := err.(type) {
			case *breakSignal:
				return result, nil
			case *continueSignal:
				continue
			case *returnSignal:
				return sig.value, err
			default:
				return nil, err
			}
		}
		result = val
	}
	return result, nil
}

func evalForStatement(stmt *ast.ForStatement, env *Environment) (Value, error) {
	loopEnv := NewEnclosedEnvironment(env)
	if stmt.Init != nil {
		if _, err := evalStatement(stmt.Init, loopEnv); err != nil {
			switch err.(type) {
			case *returnSignal:
				return nil, errors.New("return not allowed in for initializer")
			case *breakSignal, *continueSignal:
				return nil, errors.New("loop control not allowed in for initializer")
			default:
				return nil, err
			}
		}
	}

	result := NullValue
	for {
		if stmt.Condition != nil {
			cond, err := evalExpression(stmt.Condition, loopEnv)
			if err != nil {
				return nil, err
			}
			if !isTruthy(cond) {
				break
			}
		}

		if stmt.Body != nil {
			val, err := evalStatement(stmt.Body, loopEnv)
			if err != nil {
				switch sig := err.(type) {
				case *breakSignal:
					return result, nil
				case *continueSignal:
					if stmt.Post != nil {
						if _, err := evalExpression(stmt.Post, loopEnv); err != nil {
							return nil, err
						}
					}
					continue
				case *returnSignal:
					return sig.value, err
				default:
					return nil, err
				}
			} else {
				result = val
			}
		}

		if stmt.Post != nil {
			if _, err := evalExpression(stmt.Post, loopEnv); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func evalExpression(expr ast.Expression, env *Environment) (Value, error) {
	switch node := expr.(type) {
	case *ast.Identifier:
		if val, ok := env.Get(node.Name); ok {
			return val, nil
		}
		return nil, fmt.Errorf("undefined identifier %s", node.Name)
	case *ast.NumberLiteral:
		num, err := strconv.ParseFloat(node.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number literal %q", node.Value)
		}
		return NewNumber(num), nil
	case *ast.StringLiteral:
		return renderStringLiteral(node, env)
	case *ast.BooleanLiteral:
		return NewBoolean(node.Value), nil
	case *ast.NullLiteral:
		return NullValue, nil
	case *ast.AwaitExpression:
		val, err := evalExpression(node.Expression, env)
		if err != nil {
			return nil, err
		}
		switch async := val.(type) {
		case *Task:
			return async.Join()
		case *ChannelValue:
			return async.recv()
		default:
			return val, nil
		}
	case *ast.PrefixExpression:
		if node.Operator == "&" {
			ident, ok := node.Right.(*ast.Identifier)
			if !ok {
				return nil, errors.New("address-of operator requires identifier")
			}
			targetEnv, ok := env.resolve(ident.Name)
			if !ok {
				return nil, fmt.Errorf("undefined identifier %s", ident.Name)
			}
			return &Pointer{env: targetEnv, name: ident.Name}, nil
		}
		if node.Operator == "*" {
			ptrVal, err := evalExpression(node.Right, env)
			if err != nil {
				return nil, err
			}
			pointer, ok := ptrVal.(*Pointer)
			if !ok {
				return nil, fmt.Errorf("cannot dereference %s", ptrVal.Type())
			}
			return pointer.get()
		}
		right, err := evalExpression(node.Right, env)
		if err != nil {
			return nil, err
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left, err := evalExpression(node.Left, env)
		if err != nil {
			return nil, err
		}
		right, err := evalExpression(node.Right, env)
		if err != nil {
			return nil, err
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.AssignmentExpression:
		switch target := node.Target.(type) {
		case *ast.Identifier:
			right, err := evalExpression(node.Value, env)
			if err != nil {
				return nil, err
			}
			var result Value
			if node.Operator == token.ASSIGN {
				result = right
			} else {
				current, ok := env.Get(target.Name)
				if !ok {
					return nil, fmt.Errorf("undefined variable %s", target.Name)
				}
				result, err = applyAugmentedAssignment(node.Operator, current, right)
				if err != nil {
					return nil, err
				}
			}
			if _, err := env.Assign(target.Name, result); err != nil {
				return nil, err
			}
			return result, nil
		case *ast.PrefixExpression:
			if target.Operator != "*" {
				return nil, errors.New("unsupported assignment target")
			}
			ptrRef, err := evalExpression(target.Right, env)
			if err != nil {
				return nil, err
			}
			pointer, ok := ptrRef.(*Pointer)
			if !ok {
				return nil, errors.New("assignment target is not a pointer")
			}
			value, err := evalExpression(node.Value, env)
			if err != nil {
				return nil, err
			}
			result := value
			if node.Operator != token.ASSIGN {
				current, err := pointer.get()
				if err != nil {
					return nil, err
				}
				result, err = applyAugmentedAssignment(node.Operator, current, value)
				if err != nil {
					return nil, err
				}
			}
			if err := pointer.set(result); err != nil {
				return nil, err
			}
			return result, nil
		default:
			return nil, errors.New("unsupported assignment target")
		}
	case *ast.CallExpression:
		callee, err := evalExpression(node.Callee, env)
		if err != nil {
			return nil, err
		}
		args := make([]Value, 0, len(node.Arguments))
		for _, arg := range node.Arguments {
			val, err := evalExpression(arg, env)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}
		return applyFunction(callee, args)
	case *ast.ArrayLiteral:
		elements := make([]Value, 0, len(node.Elements))
		for _, el := range node.Elements {
			val, err := evalExpression(el, env)
			if err != nil {
				return nil, err
			}
			elements = append(elements, val)
		}
		return &Array{Elements: elements}, nil
	case *ast.ObjectLiteral:
		props := make(map[string]Value)
		for _, pair := range node.Pairs {
			val, err := evalExpression(pair.Value, env)
			if err != nil {
				return nil, err
			}
			props[pair.Key] = val
		}
		return &Object{Properties: props}, nil
	case *ast.IndexExpression:
		collection, err := evalExpression(node.Collection, env)
		if err != nil {
			return nil, err
		}
		indexVal, err := evalExpression(node.Index, env)
		if err != nil {
			return nil, err
		}
		return evalIndexExpression(collection, indexVal)
	case *ast.MemberExpression:
		objectVal, err := evalExpression(node.Object, env)
		if err != nil {
			return nil, err
		}
		return evalMemberExpression(objectVal, node.Property, node.Optional)
	case *ast.ElvisExpression:
		left, err := evalExpression(node.Left, env)
		if err != nil {
			return nil, err
		}
		if isTruthy(left) {
			return left, nil
		}
		return evalExpression(node.Right, env)
	case *ast.NonNullAssertion:
		val, err := evalExpression(node.Expression, env)
		if err != nil {
			return nil, err
		}
		if _, isNull := val.(*Null); isNull {
			return nil, errors.New("encountered null in non-null assertion")
		}
		return val, nil
	default:
		return nil, fmt.Errorf("runtime does not support expression %T", expr)
	}
}

func evalPrefixExpression(operator string, right Value) (Value, error) {
	switch operator {
	case "!":
		return NewBoolean(!isTruthy(right)), nil
	case "-":
		number, ok := right.(*Number)
		if !ok {
			return nil, fmt.Errorf("cannot negate %s", right.Type())
		}
		return NewNumber(-number.Value), nil
	case "+":
		if number, ok := right.(*Number); ok {
			return NewNumber(+number.Value), nil
		}
		return nil, fmt.Errorf("cannot apply unary + to %s", right.Type())
	default:
		return nil, fmt.Errorf("unknown prefix operator %s", operator)
	}
}

func evalInfixExpression(operator string, left, right Value) (Value, error) {
	switch operator {
	case "+":
		switch l := left.(type) {
		case *Number:
			r, ok := right.(*Number)
			if !ok {
				return nil, fmt.Errorf("cannot add %s to Number", right.Type())
			}
			return NewNumber(l.Value + r.Value), nil
		case *String:
			return NewString(l.Value + toString(right)), nil
		default:
			if _, ok := right.(*String); ok {
				return NewString(toString(left) + toString(right)), nil
			}
			return nil, fmt.Errorf("operator + not supported for %s", left.Type())
		}
	case "-":
		l, ok := left.(*Number)
		if !ok {
			return nil, fmt.Errorf("operator - not supported for %s", left.Type())
		}
		r, ok := right.(*Number)
		if !ok {
			return nil, fmt.Errorf("operator - requires Number on right, got %s", right.Type())
		}
		return NewNumber(l.Value - r.Value), nil
	case "*":
		l, ok := left.(*Number)
		if !ok {
			return nil, fmt.Errorf("operator * not supported for %s", left.Type())
		}
		r, ok := right.(*Number)
		if !ok {
			return nil, fmt.Errorf("operator * requires Number on right, got %s", right.Type())
		}
		return NewNumber(l.Value * r.Value), nil
	case "/":
		l, ok := left.(*Number)
		if !ok {
			return nil, fmt.Errorf("operator / not supported for %s", left.Type())
		}
		r, ok := right.(*Number)
		if !ok {
			return nil, fmt.Errorf("operator / requires Number on right, got %s", right.Type())
		}
		if r.Value == 0 {
			return nil, errors.New("division by zero")
		}
		return NewNumber(l.Value / r.Value), nil
	case "%":
		l, ok := left.(*Number)
		if !ok {
			return nil, fmt.Errorf("operator %% not supported for %s", left.Type())
		}
		r, ok := right.(*Number)
		if !ok {
			return nil, fmt.Errorf("operator %% requires Number on right, got %s", right.Type())
		}
		if r.Value == 0 {
			return nil, errors.New("modulo by zero")
		}
		return NewNumber(math.Mod(l.Value, r.Value)), nil
	case "==":
		return NewBoolean(equals(left, right)), nil
	case "!=":
		return NewBoolean(!equals(left, right)), nil
	case "<":
		return compareNumbers(operator, left, right)
	case "<=":
		return compareNumbers(operator, left, right)
	case ">":
		return compareNumbers(operator, left, right)
	case ">=":
		return compareNumbers(operator, left, right)
	case "&&":
		return NewBoolean(isTruthy(left) && isTruthy(right)), nil
	case "||":
		return NewBoolean(isTruthy(left) || isTruthy(right)), nil
	case "is":
		return evalIsOperator(left, right)
	case "!is":
		result, err := evalIsOperator(left, right)
		if err != nil {
			return nil, err
		}
		boolVal, ok := result.(*Boolean)
		if !ok {
			return nil, errors.New("operator is must return Boolean")
		}
		return NewBoolean(!boolVal.Value), nil
	default:
		return nil, fmt.Errorf("unknown infix operator %s", operator)
	}
}

func applyAugmentedAssignment(op token.Type, current, update Value) (Value, error) {
	switch op {
	case token.PLUS_ASSIGN:
		return evalInfixExpression("+", current, update)
	case token.MINUS_ASSIGN:
		return evalInfixExpression("-", current, update)
	case token.STAR_ASSIGN:
		return evalInfixExpression("*", current, update)
	case token.SLASH_ASSIGN:
		return evalInfixExpression("/", current, update)
	case token.PERCENT_ASSIGN:
		return evalInfixExpression("%", current, update)
	default:
		return nil, fmt.Errorf("unsupported assignment operator %s", op)
	}
}

func compareNumbers(operator string, left, right Value) (Value, error) {
	l, ok := left.(*Number)
	if !ok {
		return nil, fmt.Errorf("operator %s requires Number operands", operator)
	}
	r, ok := right.(*Number)
	if !ok {
		return nil, fmt.Errorf("operator %s requires Number operands", operator)
	}

	switch operator {
	case "<":
		return NewBoolean(l.Value < r.Value), nil
	case "<=":
		return NewBoolean(l.Value <= r.Value), nil
	case ">":
		return NewBoolean(l.Value > r.Value), nil
	case ">=":
		return NewBoolean(l.Value >= r.Value), nil
	default:
		return nil, fmt.Errorf("unknown comparison operator %s", operator)
	}
}

func equals(left, right Value) bool {
	switch l := left.(type) {
	case *Null:
		_, isNull := right.(*Null)
		return isNull
	case *Boolean:
		r, ok := right.(*Boolean)
		return ok && l.Value == r.Value
	case *Number:
		r, ok := right.(*Number)
		return ok && l.Value == r.Value
	case *String:
		r, ok := right.(*String)
		return ok && l.Value == r.Value
	default:
		return left == right
	}
}

func toString(val Value) string {
	return val.Inspect()
}

func isTruthy(val Value) bool {
	switch v := val.(type) {
	case *Null:
		return false
	case *Boolean:
		return v.Value
	case *Number:
		return v.Value != 0
	case *String:
		return v.Value != ""
	case *Array:
		return len(v.Elements) > 0
	case *Object:
		return len(v.Properties) > 0
	default:
		return val != nil
	}
}

func evalIndexExpression(collection, index Value) (Value, error) {
	switch col := collection.(type) {
	case *Array:
		num, ok := index.(*Number)
		if !ok {
			return nil, fmt.Errorf("array index must be Number, got %s", index.Type())
		}
		idx := int(num.Value)
		if float64(idx) != num.Value {
			return nil, errors.New("array index must be integer")
		}
		if idx < 0 || idx >= len(col.Elements) {
			return nil, fmt.Errorf("array index %d out of range", idx)
		}
		return col.Elements[idx], nil
	case *String:
		num, ok := index.(*Number)
		if !ok {
			return nil, fmt.Errorf("string index must be Number, got %s", index.Type())
		}
		idx := int(num.Value)
		if float64(idx) != num.Value {
			return nil, errors.New("string index must be integer")
		}
		runes := []rune(col.Value)
		if idx < 0 || idx >= len(runes) {
			return nil, fmt.Errorf("string index %d out of range", idx)
		}
		return NewString(string(runes[idx])), nil
	default:
		return nil, fmt.Errorf("cannot index into %s", collection.Type())
	}
}

func evalMemberExpression(object Value, property string, optional bool) (Value, error) {
	val, ok, err := getProperty(object, property)
	if err != nil {
		if optional {
			return NullValue, nil
		}
		return nil, err
	}
	if !ok {
		if optional {
			return NullValue, nil
		}
		return nil, fmt.Errorf("%s has no property %s", object.Type(), property)
	}
	return val, nil
}

func evalMatchStatement(match *ast.MatchStatement, env *Environment) (Value, error) {
	target, err := evalExpression(match.Value, env)
	if err != nil {
		return nil, err
	}

	for _, clause := range match.Cases {
		caseEnv := NewEnclosedEnvironment(env)
		matched, err := matchPattern(clause.Pattern, target, caseEnv)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		if clause.Body == nil {
			return NullValue, nil
		}
		return evalStatement(clause.Body, caseEnv)
	}

	return NullValue, nil
}

func matchPattern(pattern ast.Pattern, value Value, env *Environment) (bool, error) {
	switch p := pattern.(type) {
	case *ast.IdentifierPattern:
		if p.Identifier == nil {
			return false, nil
		}
		env.Set(p.Identifier.Name, value)
		return true, nil
	case *ast.LiteralPattern:
		expected, err := evalExpression(p.Value, env)
		if err != nil {
			return false, err
		}
		return equals(expected, value), nil
	case *ast.ObjectPattern:
		obj, ok := value.(*Object)
		if !ok {
			return false, nil
		}
		return matchObjectPattern(p, obj, env)
	case *ast.StructPattern:
		return matchStructPatternRuntime(p, value, env)
	default:
		return false, fmt.Errorf("unsupported pattern %T", pattern)
	}
}

func matchObjectPattern(pattern *ast.ObjectPattern, value *Object, env *Environment) (bool, error) {
	for _, pair := range pattern.Pairs {
		propName := pair.Key
		if pair.KeyIsIdentifier {
			propName = pair.Key
		}
		propVal, ok := value.Properties[propName]
		if !ok {
			return false, nil
		}

		matched, err := matchPattern(pair.Value, propVal, env)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}
	return true, nil
}

func matchStructPatternRuntime(pattern *ast.StructPattern, value Value, env *Environment) (bool, error) {
	if pattern.Name == nil {
		return false, nil
	}
	switch v := value.(type) {
	case *StructInstance:
		if v.Definition == nil || v.Definition.Name != pattern.Name.Name {
			return false, nil
		}
		return matchStructFields(pattern.Fields, v.Definition.Fields, v.Fields, env)
	case *ClassInstance:
		if v.Definition == nil || v.Definition.Name != pattern.Name.Name {
			return false, nil
		}
		return matchStructFields(pattern.Fields, v.Definition.Fields, v.Fields, env)
	case *EnumInstance:
		if v.Case != pattern.Name.Name {
			return false, nil
		}
		return matchEnumFields(pattern.Fields, v, env)
	default:
		return false, nil
	}
}

func matchStructFields(patterns []ast.Pattern, names []string, fields map[string]Value, env *Environment) (bool, error) {
	if len(patterns) != len(names) {
		return false, nil
	}
	for i, pat := range patterns {
		name := names[i]
		val, ok := fields[name]
		if !ok {
			return false, nil
		}
		matched, err := matchPattern(pat, val, env)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}
	return true, nil
}

func matchEnumFields(patterns []ast.Pattern, instance *EnumInstance, env *Environment) (bool, error) {
	if len(patterns) != len(instance.Order) {
		return false, nil
	}
	for i, pat := range patterns {
		name := instance.Order[i]
		val := instance.Fields[name]
		matched, err := matchPattern(pat, val, env)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}
	return true, nil
}

func getProperty(object Value, property string) (Value, bool, error) {
	switch obj := object.(type) {
	case *Object:
		val, ok := obj.Properties[property]
		return val, ok, nil
	case *Module:
		val, ok := obj.Exports[property]
		return val, ok, nil
	case *Contract:
		val, ok := obj.Exports[property]
		return val, ok, nil
	case *Array:
		if property == "length" {
			return NewNumber(float64(len(obj.Elements))), true, nil
		}
		if fn, ok := lookupExtension(obj.Type(), property); ok {
			return bindMethod(fn, obj), true, nil
		}
		return nil, false, fmt.Errorf("unknown array property %s", property)
	case *String:
		if property == "length" {
			return NewNumber(float64(len(obj.Value))), true, nil
		}
		if fn, ok := lookupExtension(obj.Type(), property); ok {
			return bindMethod(fn, obj), true, nil
		}
		return nil, false, fmt.Errorf("unknown string property %s", property)
	case *StructInstance:
		if val, ok := obj.Fields[property]; ok {
			return val, true, nil
		}
		if obj.Definition != nil {
			if method, ok := obj.Definition.Methods[property]; ok {
				return bindMethod(method, obj), true, nil
			}
			if val, ok := obj.Definition.Static[property]; ok {
				return val, true, nil
			}
		}
		return nil, false, nil
	case *StructType:
		if val, ok := obj.Static[property]; ok {
			return val, true, nil
		}
		if method, ok := obj.Methods[property]; ok {
			return method, true, nil
		}
		return nil, false, nil
	case *ClassInstance:
		if val, ok := obj.Fields[property]; ok {
			return val, true, nil
		}
		if method, ok := obj.Definition.lookupMethod(property); ok {
			return bindMethod(method, obj), true, nil
		}
		if val, ok := obj.Definition.lookupStatic(property); ok {
			return val, true, nil
		}
		return nil, false, nil
	case *ClassType:
		if val, ok := obj.lookupStatic(property); ok {
			return val, true, nil
		}
		if method, ok := obj.lookupMethod(property); ok {
			return method, true, nil
		}
		return nil, false, nil
	case *EnumType:
		val, ok := obj.Constructors[property]
		return val, ok, nil
	case *EnumInstance:
		if property == "case" {
			return NewString(obj.Case), true, nil
		}
		val, ok := obj.Fields[property]
		return val, ok, nil
	case *ChannelValue:
		switch property {
		case "send":
			return NewBuiltin("send", func(args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, errors.New("send expects a single argument")
				}
				if err := obj.send(args[0]); err != nil {
					return nil, err
				}
				return NullValue, nil
			}), true, nil
		case "recv":
			return NewBuiltin("recv", func(args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, errors.New("recv takes no arguments")
				}
				return obj.recv()
			}), true, nil
		case "close":
			return NewBuiltin("close", func(args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, errors.New("close takes no arguments")
				}
				if err := obj.close(); err != nil {
					return nil, err
				}
				return NullValue, nil
			}), true, nil
		default:
			return nil, false, fmt.Errorf("unknown channel property %s", property)
		}
	case *Task:
		if property == "join" {
			return NewBuiltin("join", func(args []Value) (Value, error) {
				if len(args) != 0 {
					return nil, errors.New("join takes no arguments")
				}
				return obj.Join()
			}), true, nil
		}
		return nil, false, fmt.Errorf("unknown task property %s", property)
	default:
		if fn, ok := lookupExtension(object.Type(), property); ok {
			return bindMethod(fn, object), true, nil
		}
		return nil, false, fmt.Errorf("cannot access property %s on %s", property, object.Type())
	}
}

func bindMethod(fn *Function, self Value) Value {
	if fn == nil {
		return NullValue
	}
	if fn.Builtin != nil {
		return NewBuiltin(fn.Name, func(args []Value) (Value, error) {
			allArgs := append([]Value{self}, args...)
			return fn.Builtin(allArgs)
		})
	}
	boundEnv := NewEnclosedEnvironment(fn.Env)
	boundEnv.Set("self", self)
	boundEnv.Set("this", self)
	return &Function{Declaration: fn.Declaration, Env: boundEnv, Name: fn.Name}
}

func cloneStore(store map[string]Value) map[string]Value {
	out := make(map[string]Value, len(store))
	for k, v := range store {
		out[k] = v
	}
	return out
}

func evalModuleDeclaration(module *ast.ModuleDeclaration, env *Environment) (Value, error) {
	if module.Name == nil {
		return nil, errors.New("module declaration requires a name")
	}
	moduleEnv := NewEnclosedEnvironment(env)
	if module.Body != nil {
		if _, err := evalBlock(module.Body, moduleEnv); err != nil {
			switch err.(type) {
			case *returnSignal:
				return nil, errors.New("return not allowed in module body")
			case *breakSignal, *continueSignal:
				return nil, errors.New("loop control not allowed in module body")
			default:
				return nil, err
			}
		}
	}
	exports := cloneStore(moduleEnv.store)
	moduleVal := &Module{Name: module.Name.Name, Exports: exports}
	env.Set(module.Name.Name, moduleVal)
	return moduleVal, nil
}

// ExecuteModuleDeclaration exposes module evaluation for external consumers like the JIT
// compiler while preserving the interpreter's semantics.
func ExecuteModuleDeclaration(module *ast.ModuleDeclaration, env *Environment) (Value, error) {
	return evalModuleDeclaration(module, env)
}

func evalImportDeclaration(imp *ast.ImportDeclaration, env *Environment) (Value, error) {
	if len(imp.Path) == 0 {
		return nil, errors.New("import path cannot be empty")
	}
	val, err := resolveImportPath(imp.Path, env)
	if err != nil {
		return nil, err
	}
	name := imp.Path[len(imp.Path)-1].Name
	if imp.Alias != nil {
		name = imp.Alias.Name
	}
	env.Set(name, val)
	return val, nil
}

func resolveImportPath(path []*ast.Identifier, env *Environment) (Value, error) {
	var current Value
	var ok bool
	for i, segment := range path {
		if i == 0 {
			current, ok = env.Get(segment.Name)
			if !ok {
				return nil, fmt.Errorf("unknown import %s", segment.Name)
			}
			continue
		}
		val, found, err := getProperty(current, segment.Name)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, fmt.Errorf("%s has no property %s", current.Type(), segment.Name)
		}
		current = val
	}
	return current, nil
}

func evalStructDeclaration(decl *ast.StructDeclaration, env *Environment) (Value, error) {
	if decl.Name == nil {
		return nil, errors.New("struct declaration requires a name")
	}
	structType := &StructType{
		Name:    decl.Name.Name,
		Fields:  make([]string, 0, len(decl.Params)),
		Methods: make(map[string]*Function),
		Static:  make(map[string]Value),
	}
	for _, param := range decl.Params {
		if param.Name != nil {
			structType.Fields = append(structType.Fields, param.Name.Name)
		}
	}
	bodyEnv := NewEnclosedEnvironment(env)
	if decl.Body != nil {
		if _, err := evalBlock(decl.Body, bodyEnv); err != nil {
			switch err.(type) {
			case *returnSignal:
				return nil, errors.New("return not allowed in struct body")
			case *breakSignal, *continueSignal:
				return nil, errors.New("loop control not allowed in struct body")
			default:
				return nil, err
			}
		}
	}
	for name, val := range bodyEnv.store {
		switch v := val.(type) {
		case *Function:
			structType.Methods[name] = v
		default:
			structType.Static[name] = v
		}
	}
	env.Set(structType.Name, structType)
	return structType, nil
}

func evalClassDeclaration(decl *ast.ClassDeclaration, env *Environment) (Value, error) {
	if decl.Name == nil {
		return nil, errors.New("class declaration requires a name")
	}
	classType := &ClassType{
		Name:    decl.Name.Name,
		Fields:  make([]string, 0, len(decl.Params)),
		Methods: make(map[string]*Function),
		Static:  make(map[string]Value),
	}
	for _, param := range decl.Params {
		if param.Name != nil {
			classType.Fields = append(classType.Fields, param.Name.Name)
		}
	}
	if decl.SuperClass != nil {
		superVal, ok := env.Get(decl.SuperClass.Name)
		if !ok {
			return nil, fmt.Errorf("unknown superclass %s", decl.SuperClass.Name)
		}
		superType, ok := superVal.(*ClassType)
		if !ok {
			return nil, fmt.Errorf("%s is not a class", decl.SuperClass.Name)
		}
		classType.Super = superType
		for name, method := range superType.Methods {
			classType.Methods[name] = method
		}
		for name, val := range superType.Static {
			classType.Static[name] = val
		}
	}

	bodyEnv := NewEnclosedEnvironment(env)
	if decl.Body != nil {
		if _, err := evalBlock(decl.Body, bodyEnv); err != nil {
			switch err.(type) {
			case *returnSignal:
				return nil, errors.New("return not allowed in class body")
			case *breakSignal, *continueSignal:
				return nil, errors.New("loop control not allowed in class body")
			default:
				return nil, err
			}
		}
	}
	for name, val := range bodyEnv.store {
		switch v := val.(type) {
		case *Function:
			classType.Methods[name] = v
		default:
			classType.Static[name] = v
		}
	}
	env.Set(classType.Name, classType)
	return classType, nil
}

func evalEnumDeclaration(decl *ast.EnumDeclaration, env *Environment) (Value, error) {
	if decl.Name == nil {
		return nil, errors.New("enum declaration requires a name")
	}
	enumType := &EnumType{
		Name:         decl.Name.Name,
		Cases:        make(map[string][]string),
		Constructors: make(map[string]Value),
	}
	for _, c := range decl.Cases {
		if c.Name == nil {
			continue
		}
		params := make([]string, 0, len(c.Params))
		for _, p := range c.Params {
			if p.Name != nil {
				params = append(params, p.Name.Name)
			}
		}
		enumType.Cases[c.Name.Name] = params
		enumType.Constructors[c.Name.Name] = newEnumConstructor(enumType, c.Name.Name, params)
	}
	env.Set(enumType.Name, enumType)
	return enumType, nil
}

func evalContractDeclaration(decl *ast.ContractDeclaration, env *Environment) (Value, error) {
	if decl.Name == nil {
		return nil, errors.New("contract declaration requires a name")
	}
	bodyEnv := NewEnclosedEnvironment(env)
	if decl.Body != nil {
		if _, err := evalBlock(decl.Body, bodyEnv); err != nil {
			switch err.(type) {
			case *returnSignal:
				return nil, errors.New("return not allowed in contract body")
			case *breakSignal, *continueSignal:
				return nil, errors.New("loop control not allowed in contract body")
			default:
				return nil, err
			}
		}
	}
	contract := &Contract{Name: decl.Name.Name, Exports: cloneStore(bodyEnv.store)}
	env.Set(contract.Name, contract)
	return contract, nil
}

func instantiateStruct(structType *StructType, args []Value) (Value, error) {
	if len(args) != len(structType.Fields) {
		return nil, fmt.Errorf("expected %d arguments to %s, got %d", len(structType.Fields), structType.Name, len(args))
	}
	fields := make(map[string]Value, len(structType.Fields))
	for i, name := range structType.Fields {
		fields[name] = args[i]
	}
	instance := &StructInstance{Definition: structType, Fields: fields}
	if init, ok := structType.Methods["init"]; ok {
		if err := callInitializer(init, instance); err != nil {
			return nil, err
		}
	}
	return instance, nil
}

func instantiateClass(classType *ClassType, args []Value) (Value, error) {
	if len(args) != len(classType.Fields) {
		return nil, fmt.Errorf("expected %d arguments to %s, got %d", len(classType.Fields), classType.Name, len(args))
	}
	fields := make(map[string]Value, len(classType.Fields))
	for i, name := range classType.Fields {
		fields[name] = args[i]
	}
	instance := &ClassInstance{Definition: classType, Fields: fields}
	if init, ok := classType.lookupMethod("init"); ok {
		if err := callInitializer(init, instance); err != nil {
			return nil, err
		}
	}
	return instance, nil
}

func callInitializer(fn *Function, self Value) error {
	bound := bindMethod(fn, self)
	_, err := applyFunction(bound, []Value{})
	return err
}

func newEnumConstructor(enum *EnumType, caseName string, params []string) Value {
	return NewBuiltin(caseName, func(args []Value) (Value, error) {
		if len(args) != len(params) {
			return nil, fmt.Errorf("expected %d arguments to %s.%s, got %d", len(params), enum.Name, caseName, len(args))
		}
		fields := make(map[string]Value, len(params))
		order := make([]string, len(params))
		for i, name := range params {
			fields[name] = args[i]
			order[i] = name
		}
		return &EnumInstance{Enum: enum, Case: caseName, Fields: fields, Order: order}, nil
	})
}

func enforceContract(block *ast.ContractBlock, env *Environment, result Value, fnName string) error {
	if block == nil {
		return nil
	}
	for _, clause := range block.Clauses {
		clauseEnv := NewEnclosedEnvironment(env)
		clauseEnv.Set("result", result)
		guardSatisfied := true
		if clause.Guard != nil {
			guard, err := evalExpression(clause.Guard, clauseEnv)
			if err != nil {
				return err
			}
			guardSatisfied = isTruthy(guard)
		}
		if !guardSatisfied {
			continue
		}
		condition, err := evalExpression(clause.Condition, clauseEnv)
		if err != nil {
			return err
		}
		if !isTruthy(condition) {
			if fnName != "" {
				return fmt.Errorf("contract violation in %s", fnName)
			}
			return errors.New("contract violation")
		}
	}
	return nil
}

func applyFunction(fn Value, args []Value) (Value, error) {
	switch callable := fn.(type) {
	case *Function:
		if callable.Builtin != nil {
			return callable.Builtin(args)
		}
		if callable.Declaration == nil {
			return nil, errors.New("function has no body")
		}
		if len(args) != len(callable.Declaration.Params) {
			return nil, fmt.Errorf("expected %d arguments, got %d", len(callable.Declaration.Params), len(args))
		}

		callEnv := NewEnclosedEnvironment(callable.Env)
		for i, param := range callable.Declaration.Params {
			callEnv.Set(param.Name.Name, args[i])
		}

		var result Value = NullValue
		var err error
		if callable.Declaration.IsExprBody {
			if callable.Declaration.BodyExpr != nil {
				result, err = evalExpression(callable.Declaration.BodyExpr, callEnv)
			}
		} else if callable.Declaration.Body != nil {
			result, err = evalBlock(callable.Declaration.Body, callEnv)
			if err != nil {
				switch sig := err.(type) {
				case *returnSignal:
					result = sig.value
					err = nil
				case *breakSignal:
					err = errors.New("break outside of loop")
				case *continueSignal:
					err = errors.New("continue outside of loop")
				}
			}
		}
		if err != nil {
			return nil, err
		}
		if callable.Declaration.Contract != nil {
			if err := enforceContract(callable.Declaration.Contract, callEnv, result, callable.Name); err != nil {
				return nil, err
			}
		}
		return result, nil
	case *StructType:
		return instantiateStruct(callable, args)
	case *ClassType:
		return instantiateClass(callable, args)
	default:
		return nil, fmt.Errorf("%s is not callable", fn.Type())
	}
}

func builtinPrint(args []Value) (Value, error) {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = arg.Inspect()
	}
	fmt.Println(strings.Join(parts, " "))
	return NullValue, nil
}

func builtinFormat(args []Value) (Value, error) {
	if len(args) == 0 {
		return nil, errors.New("format requires a template string")
	}
	tmpl, ok := args[0].(*String)
	if !ok {
		return nil, errors.New("format template must be a string")
	}
	var builder strings.Builder
	argIndex := 1
	for i := 0; i < len(tmpl.Value); i++ {
		ch := tmpl.Value[i]
		if ch == '{' {
			if i+1 < len(tmpl.Value) && tmpl.Value[i+1] == '{' {
				builder.WriteByte('{')
				i++
				continue
			}
			if i+1 < len(tmpl.Value) && tmpl.Value[i+1] == '}' {
				if argIndex >= len(args) {
					return nil, errors.New("not enough arguments for format string")
				}
				builder.WriteString(toString(args[argIndex]))
				argIndex++
				i++
				continue
			}
		}
		if ch == '}' && i+1 < len(tmpl.Value) && tmpl.Value[i+1] == '}' {
			builder.WriteByte('}')
			i++
			continue
		}
		builder.WriteByte(ch)
	}
	return NewString(builder.String()), nil
}

func builtinSpawn(args []Value) (Value, error) {
	if len(args) == 0 {
		return nil, errors.New("spawn requires a function")
	}
	fn := args[0]
	callArgs := args[1:]
	task := NewTask()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				task.deliver(NullValue, fmt.Errorf("panic: %v", r))
			}
		}()
		result, err := applyFunction(fn, callArgs)
		task.deliver(result, err)
	}()
	return task, nil
}

func builtinChannel(args []Value) (Value, error) {
	capacity := 0
	if len(args) > 0 {
		num, ok := args[0].(*Number)
		if !ok {
			return nil, errors.New("channel capacity must be a number")
		}
		if num.Value < 0 {
			return nil, errors.New("channel capacity must be non-negative")
		}
		if float64(int(num.Value)) != num.Value {
			return nil, errors.New("channel capacity must be an integer")
		}
		capacity = int(num.Value)
	}
	return &ChannelValue{ch: make(chan Value, capacity), capacity: capacity}, nil
}

func renderStringLiteral(lit *ast.StringLiteral, env *Environment) (Value, error) {
	raw := lit.Value
	var builder strings.Builder
	last := 0
	for i := 0; i < len(raw); {
		if raw[i] == '\\' {
			i += 2
			continue
		}
		if raw[i] == '$' && i+1 < len(raw) && raw[i+1] == '{' {
			chunk := raw[last:i]
			text, err := processStringChunk(chunk, lit.Raw)
			if err != nil {
				return nil, err
			}
			builder.WriteString(text)
			end, placeholder, err := extractPlaceholder(raw, i+2)
			if err != nil {
				return nil, err
			}
			exprText, formatSpec := splitFormatPlaceholder(placeholder)
			exprText = strings.TrimSpace(exprText)
			formatSpec = strings.TrimSpace(formatSpec)
			if exprText == "" {
				return nil, errors.New("empty interpolation expression")
			}
			value, err := evaluateInlineExpression(exprText, env)
			if err != nil {
				return nil, err
			}
			segment, err := applyFormatSpec(value, formatSpec, lit.Format)
			if err != nil {
				return nil, err
			}
			builder.WriteString(segment)
			last = end
			i = end
			continue
		}
		i++
	}
	if last < len(raw) {
		chunk := raw[last:]
		text, err := processStringChunk(chunk, lit.Raw)
		if err != nil {
			return nil, err
		}
		builder.WriteString(text)
	}
	return NewString(builder.String()), nil
}

func processStringChunk(chunk string, raw bool) (string, error) {
	if chunk == "" {
		return "", nil
	}
	if raw {
		return strings.ReplaceAll(chunk, "\\$", "$"), nil
	}
	return decodeEscapes(chunk)
}

func decodeEscapes(input string) (string, error) {
	var builder strings.Builder
	for i := 0; i < len(input); i++ {
		ch := input[i]
		if ch != '\\' {
			builder.WriteByte(ch)
			continue
		}
		i++
		if i >= len(input) {
			return "", errors.New("unterminated escape sequence")
		}
		switch input[i] {
		case 'n':
			builder.WriteByte('\n')
		case 'r':
			builder.WriteByte('\r')
		case 't':
			builder.WriteByte('\t')
		case '"':
			builder.WriteByte('"')
		case '\\':
			builder.WriteByte('\\')
		case '$':
			builder.WriteByte('$')
		default:
			builder.WriteByte(input[i])
		}
	}
	return builder.String(), nil
}

func extractPlaceholder(source string, start int) (int, string, error) {
	depth := 1
	for i := start; i < len(source); i++ {
		ch := source[i]
		if ch == '\\' {
			i++
			continue
		}
		if ch == '{' {
			depth++
			continue
		}
		if ch == '}' {
			depth--
			if depth == 0 {
				return i + 1, source[start:i], nil
			}
		}
	}
	return 0, "", errors.New("unterminated interpolation expression")
}

func splitFormatPlaceholder(input string) (string, string) {
	depth := 0
	quote := byte(0)
	for i := 0; i < len(input); i++ {
		ch := input[i]
		if ch == '\\' {
			i++
			continue
		}
		if ch == '"' || ch == '\'' {
			if quote == 0 {
				quote = ch
			} else if quote == ch {
				quote = 0
			}
			continue
		}
		if quote != 0 {
			continue
		}
		switch ch {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			if depth > 0 {
				depth--
			}
		case ':':
			if depth == 0 {
				return input[:i], input[i+1:]
			}
		}
	}
	return input, ""
}

func evaluateInlineExpression(src string, env *Environment) (Value, error) {
	lex := lexer.New(src)
	p := parser.New(lex)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("failed to parse interpolation: %s", strings.Join(p.Errors(), "; "))
	}
	result := NullValue
	for _, item := range program.Items {
		stmt, ok := item.(ast.Statement)
		if !ok {
			return nil, fmt.Errorf("unsupported interpolation item %T", item)
		}
		val, err := evalStatement(stmt, env)
		if err != nil {
			return nil, err
		}
		result = val
	}
	return result, nil
}

func applyFormatSpec(val Value, spec string, allow bool) (string, error) {
	if spec == "" {
		return toString(val), nil
	}
	if !allow {
		return "", errors.New("format specifiers require a format string literal")
	}
	switch spec {
	case "upper":
		return strings.ToUpper(toString(val)), nil
	case "lower":
		return strings.ToLower(toString(val)), nil
	case "title":
		return strings.Title(toString(val)), nil
	case "trim":
		return strings.TrimSpace(toString(val)), nil
	default:
		if strings.HasPrefix(spec, "%") {
			return fmt.Sprintf(spec, valueToInterface(val)), nil
		}
		return "", fmt.Errorf("unknown format specifier %q", spec)
	}
}

func valueToInterface(val Value) interface{} {
	switch v := val.(type) {
	case *Number:
		return v.Value
	case *String:
		return v.Value
	case *Boolean:
		return v.Value
	case *Null:
		return nil
	default:
		return v.Inspect()
	}
}

func implementsInterface(val Value, iface *InterfaceType) bool {
	for name, arity := range iface.Methods {
		prop, ok, err := getProperty(val, name)
		if err != nil || !ok {
			return false
		}
		fn, ok := prop.(*Function)
		if !ok {
			return false
		}
		if fn.Builtin == nil && fn.Declaration != nil {
			if len(fn.Declaration.Params) != arity {
				return false
			}
		}
	}
	return true
}

func evalIsOperator(left, right Value) (Value, error) {
	switch typeVal := right.(type) {
	case *InterfaceType:
		return NewBoolean(implementsInterface(left, typeVal)), nil
	case *ClassType:
		return NewBoolean(isInstanceOfClass(left, typeVal)), nil
	case *StructType:
		inst, ok := left.(*StructInstance)
		if !ok {
			return NewBoolean(false), nil
		}
		return NewBoolean(inst.Definition == typeVal), nil
	case *EnumType:
		inst, ok := left.(*EnumInstance)
		if !ok {
			return NewBoolean(false), nil
		}
		return NewBoolean(inst.Enum == typeVal), nil
	case *String:
		return NewBoolean(left.Type() == typeVal.Value), nil
	default:
		return nil, fmt.Errorf("unsupported right-hand side for is operator: %s", right.Type())
	}
}

func isInstanceOfClass(val Value, class *ClassType) bool {
	inst, ok := val.(*ClassInstance)
	if !ok {
		return false
	}
	current := inst.Definition
	for current != nil {
		if current == class {
			return true
		}
		current = current.Super
	}
	return false
}

func resolveCloser(val Value) (func() error, error) {
	if closable, ok := val.(interface{ Close() error }); ok {
		return func() error {
			if err := closable.Close(); err != nil {
				return wrapRuntimeError(err)
			}
			return nil
		}, nil
	}
	prop, ok, err := getProperty(val, "close")
	if err == nil && ok {
		if fn, ok := prop.(*Function); ok {
			return func() error {
				_, err := applyFunction(fn, nil)
				if err != nil {
					return err
				}
				return nil
			}, nil
		}
	}
	return nil, fmt.Errorf("value of type %s is not closable", val.Type())
}

func evalUsingStatement(stmt *ast.UsingStatement, env *Environment) (Value, error) {
	resource, err := evalExpression(stmt.Value, env)
	if err != nil {
		return nil, err
	}
	closer, err := resolveCloser(resource)
	if err != nil {
		return nil, err
	}
	blockEnv := NewEnclosedEnvironment(env)
	if stmt.Name != nil {
		blockEnv.Set(stmt.Name.Name, resource)
	}
	result, execErr := evalBlock(stmt.Body, blockEnv)
	if closeErr := closer(); closeErr != nil {
		if execErr == nil {
			execErr = closeErr
		}
	}
	if execErr != nil {
		return result, execErr
	}
	return result, nil
}

func evalTryStatement(stmt *ast.TryStatement, env *Environment) (Value, error) {
	tryEnv := NewEnclosedEnvironment(env)
	result, err := evalBlock(stmt.Body, tryEnv)

	switch err.(type) {
	case *returnSignal, *breakSignal, *continueSignal:
		if stmt.Finally != nil {
			finalEnv := NewEnclosedEnvironment(env)
			if finalResult, finalErr := evalBlock(stmt.Finally, finalEnv); finalErr != nil {
				return finalResult, finalErr
			}
		}
		return result, err
	case nil:
		// no error
	default:
		runtimeErr := wrapRuntimeError(err)
		if stmt.Catch != nil {
			catchEnv := NewEnclosedEnvironment(env)
			if stmt.Catch.Identifier != nil {
				catchEnv.Set(stmt.Catch.Identifier.Name, runtimeErr.value)
			}
			if stmt.Catch.Body != nil {
				result, err = evalBlock(stmt.Catch.Body, catchEnv)
			} else {
				err = nil
			}
		} else {
			err = runtimeErr
		}
	}

	if stmt.Finally != nil {
		finalEnv := NewEnclosedEnvironment(env)
		finalResult, finalErr := evalBlock(stmt.Finally, finalEnv)
		if finalErr != nil {
			return finalResult, finalErr
		}
		if err == nil && finalResult != nil {
			result = finalResult
		}
	}
	return result, err
}

func evalConditionStatement(stmt *ast.ConditionStatement, env *Environment) (Value, error) {
	for _, clause := range stmt.Clauses {
		cond, err := evalExpression(clause.Test, env)
		if err != nil {
			return nil, err
		}
		if isTruthy(cond) {
			clauseEnv := NewEnclosedEnvironment(env)
			return evalStatement(clause.Body, clauseEnv)
		}
	}
	if stmt.Else != nil {
		elseEnv := NewEnclosedEnvironment(env)
		return evalStatement(stmt.Else, elseEnv)
	}
	return NullValue, nil
}
