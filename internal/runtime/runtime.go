package runtime

import (
	"errors"
	"fmt"
	"math"
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
	case *ast.BlockStatement:
		blockEnv := NewEnclosedEnvironment(env)
		return evalBlock(node, blockEnv)
	default:
		return nil, fmt.Errorf("runtime does not support statement %T", stmt)
	}
}

func evalBlock(block *ast.BlockStatement, env *Environment) (Value, error) {
	result := NullValue
	for _, stmt := range block.Statements {
		val, err := evalStatement(stmt, env)
		if err != nil {
			return nil, err
		}
		result = val
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
	switch obj := object.(type) {
	case *Object:
		val, ok := obj.Properties[property]
		if !ok {
			if optional {
				return NullValue, nil
			}
			return nil, fmt.Errorf("object has no property %s", property)
		}
		return val, nil
	default:
		if optional {
			return NullValue, nil
		}
		return nil, fmt.Errorf("cannot access property %s on %s", property, object.Type())
	}
}

func applyFunction(fn Value, args []Value) (Value, error) {
	function, ok := fn.(*Function)
	if !ok {
		return nil, fmt.Errorf("%s is not callable", fn.Type())
	}
	if function.Builtin != nil {
		return function.Builtin(args)
	}
	if function.Declaration == nil {
		return nil, errors.New("function has no body")
	}
	if len(args) != len(function.Declaration.Params) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(function.Declaration.Params), len(args))
	}

	callEnv := NewEnclosedEnvironment(function.Env)
	for i, param := range function.Declaration.Params {
		callEnv.Set(param.Name.Name, args[i])
	}

	if function.Declaration.IsExprBody {
		if function.Declaration.BodyExpr == nil {
			return NullValue, nil
		}
		return evalExpression(function.Declaration.BodyExpr, callEnv)
	}
	if function.Declaration.Body == nil {
		return NullValue, nil
	}
	return evalBlock(function.Declaration.Body, callEnv)
}

func builtinPrint(args []Value) (Value, error) {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = arg.Inspect()
	}
	fmt.Println(strings.Join(parts, " "))
	return NullValue, nil
}
