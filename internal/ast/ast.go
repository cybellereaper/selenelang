package ast

import "selenelang/internal/token"

type Node interface {
	Pos() token.Position
	End() token.Position
}

type ProgramItem interface {
	Node
	programItemNode()
}

type Program struct {
	Items  []ProgramItem
	Start  token.Position
	Finish token.Position
}

func (p *Program) Pos() token.Position { return p.Start }
func (p *Program) End() token.Position { return p.Finish }

type Statement interface {
	ProgramItem
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Pattern interface {
	Node
	patternNode()
}

type Identifier struct {
	Name   string
	Start  token.Position
	Finish token.Position
}

func (i *Identifier) Pos() token.Position { return i.Start }
func (i *Identifier) End() token.Position { return i.Finish }
func (i *Identifier) expressionNode()     {}
func (i *Identifier) patternNode()        {}

// Literals

type NumberLiteral struct {
	Value  string
	Start  token.Position
	Finish token.Position
}

func (n *NumberLiteral) Pos() token.Position { return n.Start }
func (n *NumberLiteral) End() token.Position { return n.Finish }
func (n *NumberLiteral) expressionNode()     {}
func (n *NumberLiteral) patternNode()        {}

type StringLiteral struct {
	Value  string
	Start  token.Position
	Finish token.Position
}

func (s *StringLiteral) Pos() token.Position { return s.Start }
func (s *StringLiteral) End() token.Position { return s.Finish }
func (s *StringLiteral) expressionNode()     {}
func (s *StringLiteral) patternNode()        {}

type BooleanLiteral struct {
	Value  bool
	Start  token.Position
	Finish token.Position
}

func (b *BooleanLiteral) Pos() token.Position { return b.Start }
func (b *BooleanLiteral) End() token.Position { return b.Finish }
func (b *BooleanLiteral) expressionNode()     {}
func (b *BooleanLiteral) patternNode()        {}

type NullLiteral struct {
	Start  token.Position
	Finish token.Position
}

func (n *NullLiteral) Pos() token.Position { return n.Start }
func (n *NullLiteral) End() token.Position { return n.Finish }
func (n *NullLiteral) expressionNode()     {}
func (n *NullLiteral) patternNode()        {}

// Composite literals

type ArrayLiteral struct {
	Elements []Expression
	Start    token.Position
	Finish   token.Position
}

func (a *ArrayLiteral) Pos() token.Position { return a.Start }
func (a *ArrayLiteral) End() token.Position { return a.Finish }
func (a *ArrayLiteral) expressionNode()     {}

type ObjectPair struct {
	Key             string
	KeyIsIdentifier bool
	Value           Expression
}

type ObjectLiteral struct {
	Pairs  []ObjectPair
	Start  token.Position
	Finish token.Position
}

func (o *ObjectLiteral) Pos() token.Position { return o.Start }
func (o *ObjectLiteral) End() token.Position { return o.Finish }
func (o *ObjectLiteral) expressionNode()     {}

// Expressions

type AwaitExpression struct {
	Expression Expression
	Start      token.Position
	Finish     token.Position
}

func (a *AwaitExpression) Pos() token.Position { return a.Start }
func (a *AwaitExpression) End() token.Position { return a.Finish }
func (a *AwaitExpression) expressionNode()     {}

type PrefixExpression struct {
	Operator string
	Right    Expression
	Start    token.Position
	Finish   token.Position
}

func (p *PrefixExpression) Pos() token.Position { return p.Start }
func (p *PrefixExpression) End() token.Position { return p.Finish }
func (p *PrefixExpression) expressionNode()     {}

type InfixExpression struct {
	Left     Expression
	Operator string
	Right    Expression
	Start    token.Position
	Finish   token.Position
}

func (i *InfixExpression) Pos() token.Position { return i.Start }
func (i *InfixExpression) End() token.Position { return i.Finish }
func (i *InfixExpression) expressionNode()     {}

type AssignmentExpression struct {
	Target Expression
	Value  Expression
	Start  token.Position
	Finish token.Position
}

func (a *AssignmentExpression) Pos() token.Position { return a.Start }
func (a *AssignmentExpression) End() token.Position { return a.Finish }
func (a *AssignmentExpression) expressionNode()     {}

type ElvisExpression struct {
	Left   Expression
	Right  Expression
	Start  token.Position
	Finish token.Position
}

func (e *ElvisExpression) Pos() token.Position { return e.Start }
func (e *ElvisExpression) End() token.Position { return e.Finish }
func (e *ElvisExpression) expressionNode()     {}

type CallExpression struct {
	Callee    Expression
	Arguments []Expression
	Start     token.Position
	Finish    token.Position
}

func (c *CallExpression) Pos() token.Position { return c.Start }
func (c *CallExpression) End() token.Position { return c.Finish }
func (c *CallExpression) expressionNode()     {}

type IndexExpression struct {
	Collection Expression
	Index      Expression
	Start      token.Position
	Finish     token.Position
}

func (i *IndexExpression) Pos() token.Position { return i.Start }
func (i *IndexExpression) End() token.Position { return i.Finish }
func (i *IndexExpression) expressionNode()     {}

type MemberExpression struct {
	Object   Expression
	Property string
	Optional bool
	Start    token.Position
	Finish   token.Position
}

func (m *MemberExpression) Pos() token.Position { return m.Start }
func (m *MemberExpression) End() token.Position { return m.Finish }
func (m *MemberExpression) expressionNode()     {}

type NonNullAssertion struct {
	Expression Expression
	Start      token.Position
	Finish     token.Position
}

func (n *NonNullAssertion) Pos() token.Position { return n.Start }
func (n *NonNullAssertion) End() token.Position { return n.Finish }
func (n *NonNullAssertion) expressionNode()     {}

// Statements

type BlockStatement struct {
	Statements []Statement
	Start      token.Position
	Finish     token.Position
}

func (b *BlockStatement) Pos() token.Position { return b.Start }
func (b *BlockStatement) End() token.Position { return b.Finish }
func (b *BlockStatement) statementNode()      {}
func (b *BlockStatement) programItemNode()    {}

type ExpressionStatement struct {
	Expression Expression
	Start      token.Position
	Finish     token.Position
}

func (e *ExpressionStatement) Pos() token.Position { return e.Start }
func (e *ExpressionStatement) End() token.Position { return e.Finish }
func (e *ExpressionStatement) statementNode()      {}
func (e *ExpressionStatement) programItemNode()    {}

type VariableDeclaration struct {
	Mutable bool
	Name    *Identifier
	Type    *TypeAnnotation
	Value   Expression
	Start   token.Position
	Finish  token.Position
}

func (v *VariableDeclaration) Pos() token.Position { return v.Start }
func (v *VariableDeclaration) End() token.Position { return v.Finish }
func (v *VariableDeclaration) statementNode()      {}
func (v *VariableDeclaration) programItemNode()    {}

type Parameter struct {
	Name *Identifier
	Type *TypeAnnotation
}

type TypeAnnotation struct {
	Name     *Identifier
	TypeArgs []*TypeAnnotation
	Nullable bool
	Start    token.Position
	Finish   token.Position
}

func (t *TypeAnnotation) Pos() token.Position { return t.Start }
func (t *TypeAnnotation) End() token.Position { return t.Finish }

// Functions and contracts

type FunctionDeclaration struct {
	Name       *Identifier
	TypeParams []*Identifier
	Params     []Parameter
	ReturnType *TypeAnnotation
	Async      bool
	Contract   *ContractBlock
	Body       *BlockStatement
	BodyExpr   Expression
	IsExprBody bool
	Start      token.Position
	Finish     token.Position
}

func (f *FunctionDeclaration) Pos() token.Position { return f.Start }
func (f *FunctionDeclaration) End() token.Position { return f.Finish }
func (f *FunctionDeclaration) statementNode()      {}
func (f *FunctionDeclaration) programItemNode()    {}

type ContractBlock struct {
	Clauses []ContractClause
	Start   token.Position
	Finish  token.Position
}

func (c *ContractBlock) Pos() token.Position { return c.Start }
func (c *ContractBlock) End() token.Position { return c.Finish }

type ContractClause struct {
	Guard     Expression
	Condition Expression
	Start     token.Position
	Finish    token.Position
}

func (c *ContractClause) Pos() token.Position { return c.Start }
func (c *ContractClause) End() token.Position { return c.Finish }

type ClassDeclaration struct {
	Name       *Identifier
	Params     []Parameter
	SuperClass *Identifier
	Body       *BlockStatement
	Start      token.Position
	Finish     token.Position
}

func (c *ClassDeclaration) Pos() token.Position { return c.Start }
func (c *ClassDeclaration) End() token.Position { return c.Finish }
func (c *ClassDeclaration) statementNode()      {}
func (c *ClassDeclaration) programItemNode()    {}

type StructDeclaration struct {
	Name   *Identifier
	Params []Parameter
	Body   *BlockStatement
	Start  token.Position
	Finish token.Position
}

func (s *StructDeclaration) Pos() token.Position { return s.Start }
func (s *StructDeclaration) End() token.Position { return s.Finish }
func (s *StructDeclaration) statementNode()      {}
func (s *StructDeclaration) programItemNode()    {}

type EnumDeclaration struct {
	Name       *Identifier
	TypeParams []*Identifier
	Cases      []EnumCase
	Start      token.Position
	Finish     token.Position
}

func (e *EnumDeclaration) Pos() token.Position { return e.Start }
func (e *EnumDeclaration) End() token.Position { return e.Finish }
func (e *EnumDeclaration) statementNode()      {}
func (e *EnumDeclaration) programItemNode()    {}

type EnumCase struct {
	Name   *Identifier
	Params []Parameter
	Start  token.Position
	Finish token.Position
}

func (e *EnumCase) Pos() token.Position { return e.Start }
func (e *EnumCase) End() token.Position { return e.Finish }

type ContractDeclaration struct {
	Name   *Identifier
	Body   *BlockStatement
	Start  token.Position
	Finish token.Position
}

func (c *ContractDeclaration) Pos() token.Position { return c.Start }
func (c *ContractDeclaration) End() token.Position { return c.Finish }
func (c *ContractDeclaration) statementNode()      {}
func (c *ContractDeclaration) programItemNode()    {}

type ImportDeclaration struct {
	Path   []*Identifier
	Alias  *Identifier
	Start  token.Position
	Finish token.Position
}

func (i *ImportDeclaration) Pos() token.Position { return i.Start }
func (i *ImportDeclaration) End() token.Position { return i.Finish }
func (i *ImportDeclaration) statementNode()      {}
func (i *ImportDeclaration) programItemNode()    {}

type ModuleDeclaration struct {
	Name   *Identifier
	Body   *BlockStatement
	Start  token.Position
	Finish token.Position
}

func (m *ModuleDeclaration) Pos() token.Position { return m.Start }
func (m *ModuleDeclaration) End() token.Position { return m.Finish }
func (m *ModuleDeclaration) programItemNode()    {}

type MatchStatement struct {
	Value  Expression
	Cases  []MatchCase
	Start  token.Position
	Finish token.Position
}

func (m *MatchStatement) Pos() token.Position { return m.Start }
func (m *MatchStatement) End() token.Position { return m.Finish }
func (m *MatchStatement) statementNode()      {}
func (m *MatchStatement) programItemNode()    {}

type MatchCase struct {
	Pattern Pattern
	Body    Statement
	Start   token.Position
	Finish  token.Position
}

func (m *MatchCase) Pos() token.Position { return m.Start }
func (m *MatchCase) End() token.Position { return m.Finish }

type PatternPair struct {
	Key             string
	KeyIsIdentifier bool
	Value           Pattern
}

type ObjectPattern struct {
	Pairs  []PatternPair
	Start  token.Position
	Finish token.Position
}

func (o *ObjectPattern) Pos() token.Position { return o.Start }
func (o *ObjectPattern) End() token.Position { return o.Finish }
func (o *ObjectPattern) patternNode()        {}

type StructPattern struct {
	Name   *Identifier
	Fields []Pattern
	Start  token.Position
	Finish token.Position
}

func (s *StructPattern) Pos() token.Position { return s.Start }
func (s *StructPattern) End() token.Position { return s.Finish }
func (s *StructPattern) patternNode()        {}

type IdentifierPattern struct {
	Identifier *Identifier
}

func (i *IdentifierPattern) Pos() token.Position { return i.Identifier.Pos() }
func (i *IdentifierPattern) End() token.Position { return i.Identifier.End() }
func (i *IdentifierPattern) patternNode()        {}

type LiteralPattern struct {
	Value Expression
}

func (l *LiteralPattern) Pos() token.Position { return l.Value.Pos() }
func (l *LiteralPattern) End() token.Position { return l.Value.End() }
func (l *LiteralPattern) patternNode()        {}
