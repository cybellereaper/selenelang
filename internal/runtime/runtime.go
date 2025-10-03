package runtime

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"selenelang/internal/ast"
)

type Value interface {
	Type() string
	Inspect() string
}

type Number struct {
	Value float64
}

func (n *Number) Type() string    { return "Number" }
func (n *Number) Inspect() string { return strconv.FormatFloat(n.Value, 'f', -1, 64) }

type String struct {
	Value string
}

func (s *String) Type() string    { return "String" }
func (s *String) Inspect() string { return s.Value }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() string { return "Boolean" }
func (b *Boolean) Inspect() string {
	if b.Value {
		return "true"
	}
	return "false"
}

type Null struct{}

func (n *Null) Type() string    { return "Null" }
func (n *Null) Inspect() string { return "null" }

var NullValue Value = &Null{}
var TrueValue Value = &Boolean{Value: true}
var FalseValue Value = &Boolean{Value: false}

type Array struct {
	Elements []Value
}

func (a *Array) Type() string { return "Array" }
func (a *Array) Inspect() string {
	parts := make([]string, len(a.Elements))
	for i, el := range a.Elements {
		parts[i] = el.Inspect()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

type Object struct {
	Properties map[string]Value
}

func (o *Object) Type() string { return "Object" }
func (o *Object) Inspect() string {
	parts := make([]string, 0, len(o.Properties))
	for k, v := range o.Properties {
		parts = append(parts, fmt.Sprintf("%s: %s", k, v.Inspect()))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

type Module struct {
	Name    string
	Exports map[string]Value
}

func (m *Module) Type() string { return "Module" }
func (m *Module) Inspect() string {
	keys := make([]string, 0, len(m.Exports))
	for k := range m.Exports {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k
	}
	return fmt.Sprintf("<module %s [%s]>", m.Name, strings.Join(parts, ", "))
}

type StructType struct {
	Name    string
	Fields  []string
	Methods map[string]*Function
	Static  map[string]Value
}

func (s *StructType) Type() string { return "Struct" }
func (s *StructType) Inspect() string {
	return fmt.Sprintf("<struct %s>", s.Name)
}

type StructInstance struct {
	Definition *StructType
	Fields     map[string]Value
}

func (s *StructInstance) Type() string { return s.Definition.Name }
func (s *StructInstance) Inspect() string {
	keys := make([]string, 0, len(s.Fields))
	for k := range s.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%s: %s", k, s.Fields[k].Inspect())
	}
	return fmt.Sprintf("%s{%s}", s.Definition.Name, strings.Join(parts, ", "))
}

type ClassType struct {
	Name    string
	Fields  []string
	Methods map[string]*Function
	Static  map[string]Value
	Super   *ClassType
}

func (c *ClassType) Type() string { return "Class" }
func (c *ClassType) Inspect() string {
	return fmt.Sprintf("<class %s>", c.Name)
}

type ClassInstance struct {
	Definition *ClassType
	Fields     map[string]Value
}

func (c *ClassInstance) Type() string { return c.Definition.Name }
func (c *ClassInstance) Inspect() string {
	keys := make([]string, 0, len(c.Fields))
	for k := range c.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%s: %s", k, c.Fields[k].Inspect())
	}
	return fmt.Sprintf("%s{%s}", c.Definition.Name, strings.Join(parts, ", "))
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

type EnumType struct {
	Name         string
	Cases        map[string][]string
	Constructors map[string]Value
}

func (e *EnumType) Type() string { return "Enum" }
func (e *EnumType) Inspect() string {
	keys := make([]string, 0, len(e.Cases))
	for k := range e.Cases {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return fmt.Sprintf("<enum %s [%s]>", e.Name, strings.Join(keys, ", "))
}

type EnumInstance struct {
	Enum   *EnumType
	Case   string
	Fields map[string]Value
	Order  []string
}

func (e *EnumInstance) Type() string { return e.Enum.Name }
func (e *EnumInstance) Inspect() string {
	args := make([]string, len(e.Order))
	for i, name := range e.Order {
		args[i] = e.Fields[name].Inspect()
	}
	if len(args) == 0 {
		return fmt.Sprintf("%s.%s", e.Enum.Name, e.Case)
	}
	return fmt.Sprintf("%s.%s(%s)", e.Enum.Name, e.Case, strings.Join(args, ", "))
}

type Contract struct {
	Name    string
	Exports map[string]Value
}

func (c *Contract) Type() string { return "Contract" }
func (c *Contract) Inspect() string {
	keys := make([]string, 0, len(c.Exports))
	for k := range c.Exports {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return fmt.Sprintf("<contract %s [%s]>", c.Name, strings.Join(keys, ", "))
}

type BuiltinFunction func(args []Value) (Value, error)

type Function struct {
	Declaration *ast.FunctionDeclaration
	Env         *Environment
	Builtin     BuiltinFunction
	Name        string
}

func (f *Function) Type() string { return "Function" }
func (f *Function) Inspect() string {
	if f.Name != "" {
		return fmt.Sprintf("<fn %s>", f.Name)
	}
	return "<fn>"
}

func NewBoolean(v bool) Value {
	if v {
		return TrueValue
	}
	return FalseValue
}

func NewNumber(v float64) Value { return &Number{Value: v} }
func NewString(v string) Value  { return &String{Value: v} }

func NewBuiltin(name string, fn BuiltinFunction) Value {
	return &Function{Name: name, Builtin: fn}
}

type returnSignal struct {
	value Value
}

func (r *returnSignal) Error() string { return "return" }

type breakSignal struct{}

func (b *breakSignal) Error() string { return "break" }

type continueSignal struct{}

func (c *continueSignal) Error() string { return "continue" }

type Environment struct {
	store map[string]Value
	outer *Environment
}

func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]Value)}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Get(name string) (Value, bool) {
	if val, ok := e.store[name]; ok {
		return val, true
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil, false
}

func (e *Environment) Set(name string, val Value) Value {
	e.store[name] = val
	return val
}

func (e *Environment) Assign(name string, val Value) (Value, error) {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return val, nil
	}
	if e.outer != nil {
		return e.outer.Assign(name, val)
	}
	return nil, fmt.Errorf("undefined variable %s", name)
}

type Runtime struct {
	env *Environment
}

func New() *Runtime {
	env := NewEnvironment()
	env.Set("print", NewBuiltin("print", builtinPrint))
	return &Runtime{env: env}
}

func (r *Runtime) Environment() *Environment {
	return r.env
}

func (r *Runtime) Run(program *ast.Program) (Value, error) {
	return evalProgram(program, r.env)
}

func evalProgram(program *ast.Program, env *Environment) (Value, error) {
	result := NullValue
	for _, item := range program.Items {
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
			result = val
		case *ast.ModuleDeclaration:
			val, err := evalModuleDeclaration(node, env)
			if err != nil {
				return nil, err
			}
			result = val
		default:
			return nil, fmt.Errorf("runtime does not support top-level %T", item)
		}
	}
	return result, nil
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
		env.Set(node.Name.Name, fn)
		return fn, nil
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
		return NewString(node.Value), nil
	case *ast.BooleanLiteral:
		return NewBoolean(node.Value), nil
	case *ast.NullLiteral:
		return NullValue, nil
	case *ast.AwaitExpression:
		return evalExpression(node.Expression, env)
	case *ast.PrefixExpression:
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
		right, err := evalExpression(node.Value, env)
		if err != nil {
			return nil, err
		}
		ident, ok := node.Target.(*ast.Identifier)
		if !ok {
			return nil, errors.New("only simple identifiers can be assigned")
		}
		if _, err := env.Assign(ident.Name, right); err != nil {
			return nil, err
		}
		return right, nil
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
	default:
		return nil, fmt.Errorf("unknown infix operator %s", operator)
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
		return nil, false, fmt.Errorf("unknown array property %s", property)
	case *String:
		if property == "length" {
			return NewNumber(float64(len(obj.Value))), true, nil
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
	default:
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
