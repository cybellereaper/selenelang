// Package ast defines the abstract syntax tree that represents Selene
// source code as produced by the parser and consumed by later compilation
// stages.
package ast

import "selenelang/internal/token"

// Node is implemented by every AST element and reports its source span.
type Node interface {
	Pos() token.Position
	End() token.Position
}

// ProgramItem marks nodes that can appear at the top level of a module.
type ProgramItem interface {
	Node
	programItemNode()
}

// Program is the root node for a parsed Selene module.
type Program struct {
	Items  []ProgramItem
	Start  token.Position
	Finish token.Position
}

// Pos returns the position where the program begins.
func (p *Program) Pos() token.Position { return p.Start }

// End returns the position immediately following the program.
func (p *Program) End() token.Position { return p.Finish }

// Statement represents a top-level executable element.
type Statement interface {
	ProgramItem
	statementNode()
}

// Expression represents a value-producing AST node.
type Expression interface {
	Node
	expressionNode()
}

// Pattern is implemented by nodes used in pattern matching.
type Pattern interface {
	Node
	patternNode()
}

// Identifier names a value and records its lexical span.
type Identifier struct {
	Name   string
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the identifier begins.
func (i *Identifier) Pos() token.Position { return i.Start }

// End returns the location immediately after the identifier.
func (i *Identifier) End() token.Position { return i.Finish }
func (i *Identifier) expressionNode()     {}
func (i *Identifier) patternNode()        {}

// Literals

// NumberLiteral represents a numeric literal.
type NumberLiteral struct {
	Value  string
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the number literal begins.
func (n *NumberLiteral) Pos() token.Position { return n.Start }

// End returns the location immediately after the number literal.
func (n *NumberLiteral) End() token.Position { return n.Finish }
func (n *NumberLiteral) expressionNode()     {}
func (n *NumberLiteral) patternNode()        {}

// StringLiteral captures an optionally raw or formatted string literal.
type StringLiteral struct {
	Value  string
	Raw    bool
	Format bool
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the string literal begins.
func (s *StringLiteral) Pos() token.Position { return s.Start }

// End returns the location immediately after the string literal.
func (s *StringLiteral) End() token.Position { return s.Finish }
func (s *StringLiteral) expressionNode()     {}
func (s *StringLiteral) patternNode()        {}

// BooleanLiteral represents a true/false literal.
type BooleanLiteral struct {
	Value  bool
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the boolean literal begins.
func (b *BooleanLiteral) Pos() token.Position { return b.Start }

// End returns the location immediately after the boolean literal.
func (b *BooleanLiteral) End() token.Position { return b.Finish }
func (b *BooleanLiteral) expressionNode()     {}
func (b *BooleanLiteral) patternNode()        {}

// NullLiteral represents the `null` literal value.
type NullLiteral struct {
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the null literal begins.
func (n *NullLiteral) Pos() token.Position { return n.Start }

// End returns the location immediately after the null literal.
func (n *NullLiteral) End() token.Position { return n.Finish }
func (n *NullLiteral) expressionNode()     {}
func (n *NullLiteral) patternNode()        {}

// Composite literals

// ArrayLiteral records an array literal and its elements.
type ArrayLiteral struct {
	Elements []Expression
	Start    token.Position
	Finish   token.Position
}

// Pos returns the location where the array literal begins.
func (a *ArrayLiteral) Pos() token.Position { return a.Start }

// End returns the location immediately after the array literal.
func (a *ArrayLiteral) End() token.Position { return a.Finish }
func (a *ArrayLiteral) expressionNode()     {}

// ObjectPair links a literal object's key with its value expression.
type ObjectPair struct {
	Key             string
	KeyIsIdentifier bool
	Value           Expression
}

// ObjectLiteral represents a map-style literal.
type ObjectLiteral struct {
	Pairs  []ObjectPair
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the object literal begins.
func (o *ObjectLiteral) Pos() token.Position { return o.Start }

// End returns the location immediately after the object literal.
func (o *ObjectLiteral) End() token.Position { return o.Finish }
func (o *ObjectLiteral) expressionNode()     {}

// Expressions

// AwaitExpression models awaiting the value of a promise-like expression.
type AwaitExpression struct {
	Expression Expression
	Start      token.Position
	Finish     token.Position
}

// Pos returns the location where the await expression begins.
func (a *AwaitExpression) Pos() token.Position { return a.Start }

// End returns the location immediately after the await expression.
func (a *AwaitExpression) End() token.Position { return a.Finish }
func (a *AwaitExpression) expressionNode()     {}

// PrefixExpression represents a unary operator applied to a right-hand expression.
type PrefixExpression struct {
	Operator string
	Right    Expression
	Start    token.Position
	Finish   token.Position
}

// Pos returns the location where the prefix expression begins.
func (p *PrefixExpression) Pos() token.Position { return p.Start }

// End returns the location immediately after the prefix expression.
func (p *PrefixExpression) End() token.Position { return p.Finish }
func (p *PrefixExpression) expressionNode()     {}

// InfixExpression represents a binary operation joining two expressions.
type InfixExpression struct {
	Left     Expression
	Operator string
	Right    Expression
	Start    token.Position
	Finish   token.Position
}

// Pos returns the location where the infix expression begins.
func (i *InfixExpression) Pos() token.Position { return i.Start }

// End returns the location immediately after the infix expression.
func (i *InfixExpression) End() token.Position { return i.Finish }
func (i *InfixExpression) expressionNode()     {}

// AssignmentExpression captures an assignment, including compound operators.
type AssignmentExpression struct {
	Target   Expression
	Value    Expression
	Operator token.Type
	Start    token.Position
	Finish   token.Position
}

// Pos returns the location where the assignment expression begins.
func (a *AssignmentExpression) Pos() token.Position { return a.Start }

// End returns the location immediately after the assignment expression.
func (a *AssignmentExpression) End() token.Position { return a.Finish }
func (a *AssignmentExpression) expressionNode()     {}

// ElvisExpression represents the `?:` operator.
type ElvisExpression struct {
	Left   Expression
	Right  Expression
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the Elvis expression begins.
func (e *ElvisExpression) Pos() token.Position { return e.Start }

// End returns the location immediately after the Elvis expression.
func (e *ElvisExpression) End() token.Position { return e.Finish }
func (e *ElvisExpression) expressionNode()     {}

// CallExpression represents invoking a callee with arguments.
type CallExpression struct {
	Callee    Expression
	Arguments []Expression
	Start     token.Position
	Finish    token.Position
}

// Pos returns the location where the call expression begins.
func (c *CallExpression) Pos() token.Position { return c.Start }

// End returns the location immediately after the call expression.
func (c *CallExpression) End() token.Position { return c.Finish }
func (c *CallExpression) expressionNode()     {}

// IndexExpression models bracket indexing into a collection.
type IndexExpression struct {
	Collection Expression
	Index      Expression
	Start      token.Position
	Finish     token.Position
}

// Pos returns the location where the index expression begins.
func (i *IndexExpression) Pos() token.Position { return i.Start }

// End returns the location immediately after the index expression.
func (i *IndexExpression) End() token.Position { return i.Finish }
func (i *IndexExpression) expressionNode()     {}

// MemberExpression represents property access on an object.
type MemberExpression struct {
	Object   Expression
	Property string
	Optional bool
	Start    token.Position
	Finish   token.Position
}

// Pos returns the location where the member expression begins.
func (m *MemberExpression) Pos() token.Position { return m.Start }

// End returns the location immediately after the member expression.
func (m *MemberExpression) End() token.Position { return m.Finish }
func (m *MemberExpression) expressionNode()     {}

// NonNullAssertion represents the `!!` operator.
type NonNullAssertion struct {
	Expression Expression
	Start      token.Position
	Finish     token.Position
}

// Pos returns the location where the non-null assertion begins.
func (n *NonNullAssertion) Pos() token.Position { return n.Start }

// End returns the location immediately after the non-null assertion.
func (n *NonNullAssertion) End() token.Position { return n.Finish }
func (n *NonNullAssertion) expressionNode()     {}

// Statements

// BlockStatement groups a series of statements within braces.
type BlockStatement struct {
	Statements []Statement
	Start      token.Position
	Finish     token.Position
}

// Pos returns the location where the block starts.
func (b *BlockStatement) Pos() token.Position { return b.Start }

// End returns the location immediately after the block.
func (b *BlockStatement) End() token.Position { return b.Finish }
func (b *BlockStatement) statementNode()      {}
func (b *BlockStatement) programItemNode()    {}

// ExpressionStatement wraps an expression as a statement.
type ExpressionStatement struct {
	Expression Expression
	Start      token.Position
	Finish     token.Position
}

// Pos returns the location where the expression statement begins.
func (e *ExpressionStatement) Pos() token.Position { return e.Start }

// End returns the location immediately after the expression statement.
func (e *ExpressionStatement) End() token.Position { return e.Finish }
func (e *ExpressionStatement) statementNode()      {}
func (e *ExpressionStatement) programItemNode()    {}

// IfStatement models conditional branching.
type IfStatement struct {
	Condition   Expression
	Consequence Statement
	Alternative Statement
	Start       token.Position
	Finish      token.Position
}

// Pos returns the location where the if statement begins.
func (i *IfStatement) Pos() token.Position { return i.Start }

// End returns the location immediately after the if statement.
func (i *IfStatement) End() token.Position { return i.Finish }
func (i *IfStatement) statementNode()      {}
func (i *IfStatement) programItemNode()    {}

// WhileStatement represents a looping construct evaluated before the body.
type WhileStatement struct {
	Condition Expression
	Body      Statement
	Start     token.Position
	Finish    token.Position
}

// Pos returns the location where the while statement begins.
func (w *WhileStatement) Pos() token.Position { return w.Start }

// End returns the location immediately after the while statement.
func (w *WhileStatement) End() token.Position { return w.Finish }
func (w *WhileStatement) statementNode()      {}
func (w *WhileStatement) programItemNode()    {}

// ForStatement encodes a three-part loop.
type ForStatement struct {
	Init      Statement
	Condition Expression
	Post      Expression
	Body      Statement
	Start     token.Position
	Finish    token.Position
}

// Pos returns the location where the for statement begins.
func (f *ForStatement) Pos() token.Position { return f.Start }

// End returns the location immediately after the for statement.
func (f *ForStatement) End() token.Position { return f.Finish }
func (f *ForStatement) statementNode()      {}
func (f *ForStatement) programItemNode()    {}

// ReturnStatement returns control to the caller with an optional value.
type ReturnStatement struct {
	Value  Expression
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the return statement begins.
func (r *ReturnStatement) Pos() token.Position { return r.Start }

// End returns the location immediately after the return statement.
func (r *ReturnStatement) End() token.Position { return r.Finish }
func (r *ReturnStatement) statementNode()      {}
func (r *ReturnStatement) programItemNode()    {}

// BreakStatement exits the nearest enclosing loop.
type BreakStatement struct {
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the break statement begins.
func (b *BreakStatement) Pos() token.Position { return b.Start }

// End returns the location immediately after the break statement.
func (b *BreakStatement) End() token.Position { return b.Finish }
func (b *BreakStatement) statementNode()      {}
func (b *BreakStatement) programItemNode()    {}

// ContinueStatement skips to the next loop iteration.
type ContinueStatement struct {
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the continue statement begins.
func (c *ContinueStatement) Pos() token.Position { return c.Start }

// End returns the location immediately after the continue statement.
func (c *ContinueStatement) End() token.Position { return c.Finish }
func (c *ContinueStatement) statementNode()      {}
func (c *ContinueStatement) programItemNode()    {}

// ThrowStatement raises an exception-like value.
type ThrowStatement struct {
	Value  Expression
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the throw statement begins.
func (t *ThrowStatement) Pos() token.Position { return t.Start }

// End returns the location immediately after the throw statement.
func (t *ThrowStatement) End() token.Position { return t.Finish }
func (t *ThrowStatement) statementNode()      {}
func (t *ThrowStatement) programItemNode()    {}

// UsingStatement introduces a scoped resource binding.
type UsingStatement struct {
	Name   *Identifier
	Value  Expression
	Body   *BlockStatement
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the using statement begins.
func (u *UsingStatement) Pos() token.Position { return u.Start }

// End returns the location immediately after the using statement.
func (u *UsingStatement) End() token.Position { return u.Finish }
func (u *UsingStatement) statementNode()      {}
func (u *UsingStatement) programItemNode()    {}

// TryStatement models structured exception handling.
type TryStatement struct {
	Body    *BlockStatement
	Catch   *CatchClause
	Finally *BlockStatement
	Start   token.Position
	Finish  token.Position
}

// Pos returns the location where the try statement begins.
func (t *TryStatement) Pos() token.Position { return t.Start }

// End returns the location immediately after the try statement.
func (t *TryStatement) End() token.Position { return t.Finish }
func (t *TryStatement) statementNode()      {}
func (t *TryStatement) programItemNode()    {}

// CatchClause handles values thrown in a try block.
type CatchClause struct {
	Identifier *Identifier
	Body       *BlockStatement
	Start      token.Position
	Finish     token.Position
}

// Pos returns the location where the catch clause begins.
func (c *CatchClause) Pos() token.Position { return c.Start }

// End returns the location immediately after the catch clause.
func (c *CatchClause) End() token.Position { return c.Finish }

// ConditionClause pairs a boolean test with an associated body.
type ConditionClause struct {
	Test   Expression
	Body   Statement
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the condition clause begins.
func (c *ConditionClause) Pos() token.Position { return c.Start }

// End returns the location immediately after the condition clause.
func (c *ConditionClause) End() token.Position { return c.Finish }

// ConditionStatement chains conditional clauses with an optional fallback.
type ConditionStatement struct {
	Clauses []ConditionClause
	Else    Statement
	Start   token.Position
	Finish  token.Position
}

// Pos returns the location where the condition statement begins.
func (c *ConditionStatement) Pos() token.Position { return c.Start }

// End returns the location immediately after the condition statement.
func (c *ConditionStatement) End() token.Position { return c.Finish }
func (c *ConditionStatement) statementNode()      {}
func (c *ConditionStatement) programItemNode()    {}

// VariableDeclaration introduces a new binding.
type VariableDeclaration struct {
	Mutable bool
	Name    *Identifier
	Type    *TypeAnnotation
	Value   Expression
	Start   token.Position
	Finish  token.Position
}

// Pos returns the location where the variable declaration begins.
func (v *VariableDeclaration) Pos() token.Position { return v.Start }

// End returns the location immediately after the variable declaration.
func (v *VariableDeclaration) End() token.Position { return v.Finish }
func (v *VariableDeclaration) statementNode()      {}
func (v *VariableDeclaration) programItemNode()    {}

// Parameter describes a function parameter.
type Parameter struct {
	Name *Identifier
	Type *TypeAnnotation
}

// TypeAnnotation records the declared type of an expression.
type TypeAnnotation struct {
	Name     *Identifier
	TypeArgs []*TypeAnnotation
	Nullable bool
	Start    token.Position
	Finish   token.Position
}

// Pos returns the location where the type annotation begins.
func (t *TypeAnnotation) Pos() token.Position { return t.Start }

// End returns the location immediately after the type annotation.
func (t *TypeAnnotation) End() token.Position { return t.Finish }

// Functions and contracts

// FunctionDeclaration declares a function or method.
type FunctionDeclaration struct {
	Name        *Identifier
	Receiver    *TypeAnnotation
	TypeParams  []*Identifier
	Params      []Parameter
	ReturnType  *TypeAnnotation
	Async       bool
	Contract    *ContractBlock
	Body        *BlockStatement
	BodyExpr    Expression
	IsExprBody  bool
	IsExtension bool
	Start       token.Position
	Finish      token.Position
}

// Pos returns the location where the function declaration begins.
func (f *FunctionDeclaration) Pos() token.Position { return f.Start }

// End returns the location immediately after the function declaration.
func (f *FunctionDeclaration) End() token.Position { return f.Finish }
func (f *FunctionDeclaration) statementNode()      {}
func (f *FunctionDeclaration) programItemNode()    {}

// ContractBlock collects design-by-contract clauses.
type ContractBlock struct {
	Clauses []ContractClause
	Start   token.Position
	Finish  token.Position
}

// Pos returns the location where the contract block begins.
func (c *ContractBlock) Pos() token.Position { return c.Start }

// End returns the location immediately after the contract block.
func (c *ContractBlock) End() token.Position { return c.Finish }

// ContractClause defines a pre- or post-condition for a function.
type ContractClause struct {
	Guard     Expression
	Condition Expression
	Start     token.Position
	Finish    token.Position
}

// Pos returns the location where the contract clause begins.
func (c *ContractClause) Pos() token.Position { return c.Start }

// End returns the location immediately after the contract clause.
func (c *ContractClause) End() token.Position { return c.Finish }

// ClassDeclaration defines a class with optional inheritance.
type ClassDeclaration struct {
	Name       *Identifier
	Params     []Parameter
	SuperClass *Identifier
	Body       *BlockStatement
	Start      token.Position
	Finish     token.Position
}

// Pos returns the location where the class declaration begins.
func (c *ClassDeclaration) Pos() token.Position { return c.Start }

// End returns the location immediately after the class declaration.
func (c *ClassDeclaration) End() token.Position { return c.Finish }
func (c *ClassDeclaration) statementNode()      {}
func (c *ClassDeclaration) programItemNode()    {}

// InterfaceDeclaration introduces an interface type.
type InterfaceDeclaration struct {
	Name    *Identifier
	Methods []InterfaceMethod
	Start   token.Position
	Finish  token.Position
}

// Pos returns the location where the interface declaration begins.
func (i *InterfaceDeclaration) Pos() token.Position { return i.Start }

// End returns the location immediately after the interface declaration.
func (i *InterfaceDeclaration) End() token.Position { return i.Finish }
func (i *InterfaceDeclaration) statementNode()      {}
func (i *InterfaceDeclaration) programItemNode()    {}

// InterfaceMethod describes a required method on an interface.
type InterfaceMethod struct {
	Name       *Identifier
	Params     []Parameter
	ReturnType *TypeAnnotation
	Start      token.Position
	Finish     token.Position
}

// Pos returns the location where the interface method begins.
func (m *InterfaceMethod) Pos() token.Position { return m.Start }

// End returns the location immediately after the interface method.
func (m *InterfaceMethod) End() token.Position { return m.Finish }

// StructDeclaration defines a struct type.
type StructDeclaration struct {
	Name   *Identifier
	Params []Parameter
	Body   *BlockStatement
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the struct declaration begins.
func (s *StructDeclaration) Pos() token.Position { return s.Start }

// End returns the location immediately after the struct declaration.
func (s *StructDeclaration) End() token.Position { return s.Finish }
func (s *StructDeclaration) statementNode()      {}
func (s *StructDeclaration) programItemNode()    {}

// EnumDeclaration defines an enumeration type.
type EnumDeclaration struct {
	Name       *Identifier
	TypeParams []*Identifier
	Cases      []EnumCase
	Start      token.Position
	Finish     token.Position
}

// Pos returns the location where the enum declaration begins.
func (e *EnumDeclaration) Pos() token.Position { return e.Start }

// End returns the location immediately after the enum declaration.
func (e *EnumDeclaration) End() token.Position { return e.Finish }
func (e *EnumDeclaration) statementNode()      {}
func (e *EnumDeclaration) programItemNode()    {}

// EnumCase describes a single enumeration variant.
type EnumCase struct {
	Name   *Identifier
	Params []Parameter
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the enum case begins.
func (e *EnumCase) Pos() token.Position { return e.Start }

// End returns the location immediately after the enum case.
func (e *EnumCase) End() token.Position { return e.Finish }

// ContractDeclaration defines a contract type.
type ContractDeclaration struct {
	Name   *Identifier
	Body   *BlockStatement
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the contract declaration begins.
func (c *ContractDeclaration) Pos() token.Position { return c.Start }

// End returns the location immediately after the contract declaration.
func (c *ContractDeclaration) End() token.Position { return c.Finish }
func (c *ContractDeclaration) statementNode()      {}
func (c *ContractDeclaration) programItemNode()    {}

// ImportDeclaration brings a module or package into scope.
type ImportDeclaration struct {
	Path        []*Identifier
	PathLiteral string
	Alias       *Identifier
	Start       token.Position
	Finish      token.Position
}

// Pos returns the location where the import declaration begins.
func (i *ImportDeclaration) Pos() token.Position { return i.Start }

// End returns the location immediately after the import declaration.
func (i *ImportDeclaration) End() token.Position { return i.Finish }
func (i *ImportDeclaration) statementNode()      {}
func (i *ImportDeclaration) programItemNode()    {}

// PackageDeclaration declares the package name for a file.
type PackageDeclaration struct {
	Name   *Identifier
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the package declaration begins.
func (p *PackageDeclaration) Pos() token.Position { return p.Start }

// End returns the location immediately after the package declaration.
func (p *PackageDeclaration) End() token.Position { return p.Finish }
func (p *PackageDeclaration) programItemNode()    {}

// ModuleDeclaration defines a module and its body.
type ModuleDeclaration struct {
	Name   *Identifier
	Body   *BlockStatement
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the module declaration begins.
func (m *ModuleDeclaration) Pos() token.Position { return m.Start }

// End returns the location immediately after the module declaration.
func (m *ModuleDeclaration) End() token.Position { return m.Finish }
func (m *ModuleDeclaration) programItemNode()    {}

// MatchStatement performs structural pattern matching.
type MatchStatement struct {
	Value  Expression
	Cases  []MatchCase
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the match statement begins.
func (m *MatchStatement) Pos() token.Position { return m.Start }

// End returns the location immediately after the match statement.
func (m *MatchStatement) End() token.Position { return m.Finish }
func (m *MatchStatement) statementNode()      {}
func (m *MatchStatement) programItemNode()    {}

// MatchCase pairs a pattern with a body to execute.
type MatchCase struct {
	Pattern Pattern
	Body    Statement
	Start   token.Position
	Finish  token.Position
}

// Pos returns the location where the match case begins.
func (m *MatchCase) Pos() token.Position { return m.Start }

// End returns the location immediately after the match case.
func (m *MatchCase) End() token.Position { return m.Finish }

// PatternPair associates a key with a nested pattern.
type PatternPair struct {
	Key             string
	KeyIsIdentifier bool
	Value           Pattern
}

// ObjectPattern matches object literals by key.
type ObjectPattern struct {
	Pairs  []PatternPair
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the object pattern begins.
func (o *ObjectPattern) Pos() token.Position { return o.Start }

// End returns the location immediately after the object pattern.
func (o *ObjectPattern) End() token.Position { return o.Finish }
func (o *ObjectPattern) patternNode()        {}

// StructPattern matches struct fields in order.
type StructPattern struct {
	Name   *Identifier
	Fields []Pattern
	Start  token.Position
	Finish token.Position
}

// Pos returns the location where the struct pattern begins.
func (s *StructPattern) Pos() token.Position { return s.Start }

// End returns the location immediately after the struct pattern.
func (s *StructPattern) End() token.Position { return s.Finish }
func (s *StructPattern) patternNode()        {}

// IdentifierPattern binds a matched value to a name.
type IdentifierPattern struct {
	Identifier *Identifier
}

// Pos returns the location where the identifier pattern begins.
func (i *IdentifierPattern) Pos() token.Position { return i.Identifier.Pos() }

// End returns the location immediately after the identifier pattern.
func (i *IdentifierPattern) End() token.Position { return i.Identifier.End() }
func (i *IdentifierPattern) patternNode()        {}

// LiteralPattern matches against a literal expression.
type LiteralPattern struct {
	Value Expression
}

// Pos returns the location where the literal pattern begins.
func (l *LiteralPattern) Pos() token.Position { return l.Value.Pos() }

// End returns the location immediately after the literal pattern.
func (l *LiteralPattern) End() token.Position { return l.Value.End() }
func (l *LiteralPattern) patternNode()        {}
