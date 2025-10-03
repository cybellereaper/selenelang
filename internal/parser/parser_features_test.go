package parser

import (
	"testing"

	"selenelang/internal/ast"
	"selenelang/internal/lexer"
)

func parseProgram(t *testing.T, src string) *ast.Program {
	t.Helper()
	l := lexer.New(src)
	p := New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}
	return program
}

func TestParserParsesVariableDeclarations(t *testing.T) {
	source := `
let answer = 42;
var counter: Number = answer;
`

	program := parseProgram(t, source)
	if len(program.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(program.Items))
	}

	letDecl, ok := program.Items[0].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("expected first item to be variable declaration, got %T", program.Items[0])
	}
	if letDecl.Mutable {
		t.Fatalf("expected let declaration to be immutable")
	}
	if letDecl.Name.Name != "answer" {
		t.Fatalf("expected identifier 'answer', got %s", letDecl.Name.Name)
	}
	if lit, ok := letDecl.Value.(*ast.NumberLiteral); !ok || lit.Value != "42" {
		t.Fatalf("expected number literal 42, got %T", letDecl.Value)
	}

	varDecl, ok := program.Items[1].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("expected second item to be variable declaration, got %T", program.Items[1])
	}
	if !varDecl.Mutable {
		t.Fatalf("expected var declaration to be mutable")
	}
	if varDecl.Type == nil || varDecl.Type.Name.Name != "Number" {
		t.Fatalf("expected type annotation Number")
	}
}

func TestParserParsesFunctionAndExtensionFeatures(t *testing.T) {
	source := `
fn compute<T>(value: Number): Number async contract {
    returns(value > 0) => value;
} => value ?: 0;

ext fn Result<T>.success(value: T): Result<T> async => value;
`

	program := parseProgram(t, source)
	if len(program.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(program.Items))
	}

	fnDecl, ok := program.Items[0].(*ast.FunctionDeclaration)
	if !ok {
		t.Fatalf("expected first item to be function declaration, got %T", program.Items[0])
	}
	if len(fnDecl.TypeParams) != 1 || fnDecl.TypeParams[0].Name != "T" {
		t.Fatalf("expected single type parameter T")
	}
	if len(fnDecl.Params) != 1 || fnDecl.Params[0].Name.Name != "value" {
		t.Fatalf("expected parameter named value")
	}
	if fnDecl.ReturnType == nil || fnDecl.ReturnType.Name.Name != "Number" {
		t.Fatalf("expected return type Number")
	}
	if !fnDecl.Async {
		t.Fatalf("expected function to be async")
	}
	if fnDecl.Contract == nil || len(fnDecl.Contract.Clauses) != 1 {
		t.Fatalf("expected single contract clause")
	}
	if !fnDecl.IsExprBody {
		t.Fatalf("expected function to use expression body")
	}
	if _, ok := fnDecl.BodyExpr.(*ast.ElvisExpression); !ok {
		t.Fatalf("expected elvis expression body, got %T", fnDecl.BodyExpr)
	}

	extDecl, ok := program.Items[1].(*ast.FunctionDeclaration)
	if !ok || !extDecl.IsExtension {
		t.Fatalf("expected extension function, got %T", program.Items[1])
	}
	if extDecl.Receiver == nil || extDecl.Receiver.Name.Name != "Result" {
		t.Fatalf("expected receiver type Result")
	}
	if !extDecl.Async {
		t.Fatalf("expected extension to be async")
	}
	if !extDecl.IsExprBody {
		t.Fatalf("expected extension to use expression body")
	}
}

func TestParserParsesTypeDeclarations(t *testing.T) {
	source := `
class Point(x: Number, y: Number): Shape {}
struct Pair(left: String, right: String) {}
enum Status {
    Ready;
    Error(reason: String);
}
interface Runnable {
    fn run(): Void;
}
contract Loggable {
    fn log(): Void {
        return;
    }
}
`

	program := parseProgram(t, source)
	if len(program.Items) != 5 {
		t.Fatalf("expected 5 declarations, got %d", len(program.Items))
	}

	if classDecl, ok := program.Items[0].(*ast.ClassDeclaration); !ok || classDecl.SuperClass.Name != "Shape" {
		t.Fatalf("expected class declaration with superclass Shape, got %T", program.Items[0])
	}

	if _, ok := program.Items[1].(*ast.StructDeclaration); !ok {
		t.Fatalf("expected struct declaration, got %T", program.Items[1])
	}

	if enumDecl, ok := program.Items[2].(*ast.EnumDeclaration); !ok || len(enumDecl.Cases) != 2 {
		t.Fatalf("expected enum with two cases, got %T", program.Items[2])
	}

	if ifaceDecl, ok := program.Items[3].(*ast.InterfaceDeclaration); !ok || len(ifaceDecl.Methods) != 1 {
		t.Fatalf("expected interface with one method, got %T", program.Items[3])
	}

	if _, ok := program.Items[4].(*ast.ContractDeclaration); !ok {
		t.Fatalf("expected contract declaration, got %T", program.Items[4])
	}
}

func TestParserParsesControlFlow(t *testing.T) {
	source := `
fn main() {
    if true { return; } else { return; }
    while ready { break; }
    for (let i = 0; i < 10; i += 1) { continue; }
    using resource = acquire() { return; }
    try { throw err; } catch (err) { return; } finally { return; }
    condition {
        when ready => { break; }
        else => { continue; }
    }
    match value {
        1 => return;
        { kind: kind } => return;
    }
}
`

	program := parseProgram(t, source)
	if len(program.Items) != 1 {
		t.Fatalf("expected single item, got %d", len(program.Items))
	}

	fnDecl, ok := program.Items[0].(*ast.FunctionDeclaration)
	if !ok || fnDecl.Body == nil {
		t.Fatalf("expected function with body, got %T", program.Items[0])
	}

	stmts := fnDecl.Body.Statements
	if len(stmts) != 7 {
		t.Fatalf("expected 7 statements, got %d", len(stmts))
	}

	if _, ok := stmts[0].(*ast.IfStatement); !ok {
		t.Fatalf("expected if statement, got %T", stmts[0])
	}
	if _, ok := stmts[1].(*ast.WhileStatement); !ok {
		t.Fatalf("expected while statement, got %T", stmts[1])
	}
	if _, ok := stmts[2].(*ast.ForStatement); !ok {
		t.Fatalf("expected for statement, got %T", stmts[2])
	}
	if usingStmt, ok := stmts[3].(*ast.UsingStatement); !ok || usingStmt.Name.Name != "resource" {
		t.Fatalf("expected using statement with resource binding, got %T", stmts[3])
	}
	if tryStmt, ok := stmts[4].(*ast.TryStatement); !ok || tryStmt.Catch == nil || tryStmt.Finally == nil {
		t.Fatalf("expected try statement with catch and finally, got %T", stmts[4])
	}
	if condStmt, ok := stmts[5].(*ast.ConditionStatement); !ok || len(condStmt.Clauses) != 1 || condStmt.Else == nil {
		t.Fatalf("expected condition statement with clause and else, got %T", stmts[5])
	}
	if matchStmt, ok := stmts[6].(*ast.MatchStatement); !ok || len(matchStmt.Cases) != 2 {
		t.Fatalf("expected match with two cases, got %T", stmts[6])
	}
}

func TestParserParsesExpressionFeatures(t *testing.T) {
	source := `
let result = await compute()?.value!! ?: 0;
let composite = { label: "ok", count: [1, 2, 3][0] };
`

	program := parseProgram(t, source)
	if len(program.Items) != 2 {
		t.Fatalf("expected two declarations, got %d", len(program.Items))
	}

	resultDecl := program.Items[0].(*ast.VariableDeclaration)
	elvis, ok := resultDecl.Value.(*ast.ElvisExpression)
	if !ok {
		t.Fatalf("expected elvis expression, got %T", resultDecl.Value)
	}
	awaitExpr, ok := elvis.Left.(*ast.AwaitExpression)
	if !ok {
		t.Fatalf("expected await expression on left, got %T", elvis.Left)
	}
	nonNull, ok := awaitExpr.Expression.(*ast.NonNullAssertion)
	if !ok {
		t.Fatalf("expected non-null assertion, got %T", awaitExpr.Expression)
	}
	member, ok := nonNull.Expression.(*ast.MemberExpression)
	if !ok || !member.Optional {
		t.Fatalf("expected optional member access, got %T", nonNull.Expression)
	}

	compositeDecl := program.Items[1].(*ast.VariableDeclaration)
	obj, ok := compositeDecl.Value.(*ast.ObjectLiteral)
	if !ok || len(obj.Pairs) != 2 {
		t.Fatalf("expected object literal with two pairs, got %T", compositeDecl.Value)
	}
	if _, ok := obj.Pairs[1].Value.(*ast.IndexExpression); !ok {
		t.Fatalf("expected index expression in object value, got %T", obj.Pairs[1].Value)
	}
}
