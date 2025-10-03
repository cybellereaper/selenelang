package parser

import (
	"fmt"

	"selenelang/internal/ast"
	"selenelang/internal/lexer"
	"selenelang/internal/token"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	errors []string

	prefixParseFns map[token.Type]prefixParseFn
	infixParseFns  map[token.Type]infixParseFn
}

const (
	_ int = iota
	LOWEST
	ASSIGNMENT
	ELVIS
	OR
	AND
	EQUALITY
	COMPARISON
	SUM
	PRODUCT
	PREFIX
	CALL
)

var precedences = map[token.Type]int{
	token.ASSIGN:   ASSIGNMENT,
	token.ELVIS:    ELVIS,
	token.OR:       OR,
	token.AND:      AND,
	token.EQ:       EQUALITY,
	token.NOT_EQ:   EQUALITY,
	token.LT:       COMPARISON,
	token.LTE:      COMPARISON,
	token.GT:       COMPARISON,
	token.GTE:      COMPARISON,
	token.IS:       COMPARISON,
	token.NOT_IS:   COMPARISON,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
	token.PERCENT:  PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: CALL,
	token.DOT:      CALL,
	token.SAFE_DOT: CALL,
	token.NON_NULL: CALL,
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:              l,
		prefixParseFns: make(map[token.Type]prefixParseFn),
		infixParseFns:  make(map[token.Type]infixParseFn),
	}

	// Read two tokens so cur and peek are set
	p.nextToken()
	p.nextToken()

	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.NUMBER, p.parseNumberLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.NULL, p.parseNullLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseObjectLiteral)
	p.registerPrefix(token.AWAIT, p.parseAwaitExpression)

	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.PERCENT, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.IS, p.parseInfixExpression)
	p.registerInfix(token.NOT_IS, p.parseInfixExpression)
	p.registerInfix(token.ELVIS, p.parseElvisExpression)
	p.registerInfix(token.ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.DOT, p.parseMemberExpression)
	p.registerInfix(token.SAFE_DOT, p.parseMemberExpression)
	p.registerInfix(token.NON_NULL, p.parseNonNullAssertion)

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) registerPrefix(t token.Type, fn prefixParseFn) {
	p.prefixParseFns[t] = fn
}

func (p *Parser) registerInfix(t token.Type, fn infixParseFn) {
	p.infixParseFns[t] = fn
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	if p.curToken.Type != token.EOF {
		program.Start = p.curToken.Pos
	}

	for p.curToken.Type != token.EOF {
		var item ast.ProgramItem
		if p.curToken.Type == token.MODULE {
			item = p.parseModuleDeclaration()
		} else {
			stmt := p.parseStatement()
			if stmt != nil {
				item = stmt
			}
		}
		if item != nil {
			program.Items = append(program.Items, item)
		}
		p.nextToken()
	}

	program.Finish = p.curToken.Pos
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET, token.VAR:
		return p.parseVariableDeclaration()
	case token.FN:
		return p.parseFunctionDeclaration()
	case token.CLASS:
		return p.parseClassDeclaration()
	case token.STRUCT:
		return p.parseStructDeclaration()
	case token.ENUM:
		return p.parseEnumDeclaration()
	case token.CONTRACT:
		return p.parseContractDeclaration()
	case token.IMPORT:
		return p.parseImportDeclaration()
	case token.MATCH:
		return p.parseMatchStatement()
	case token.LBRACE:
		return p.parseBlockStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseModuleDeclaration() ast.ProgramItem {
	module := &ast.ModuleDeclaration{Start: p.curToken.Pos}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	module.Name = p.currentIdentifier()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	block := p.parseBlockStatement()
	module.Body = block
	if block != nil {
		module.Finish = block.End()
	} else {
		module.Finish = p.curToken.End
	}
	return module
}

func (p *Parser) parseVariableDeclaration() ast.Statement {
	stmt := &ast.VariableDeclaration{Start: p.curToken.Pos, Mutable: p.curToken.Type == token.VAR}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = p.currentIdentifier()

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
		p.nextToken()
		stmt.Type = p.parseTypeAnnotation()
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if stmt.Value != nil {
		stmt.Finish = stmt.Value.End()
	}

	if !p.expectPeek(token.SEMICOLON) {
		return stmt
	}
	stmt.Finish = p.curToken.End
	return stmt
}

func (p *Parser) parseFunctionDeclaration() ast.Statement {
	fn := &ast.FunctionDeclaration{Start: p.curToken.Pos}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	fn.Name = p.currentIdentifier()

	if p.peekTokenIs(token.LT) {
		fn.TypeParams = p.parseTypeParameters()
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	fn.Params = p.parseParameterList(token.RPAREN)

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
		p.nextToken()
		fn.ReturnType = p.parseTypeAnnotation()
	}

	if p.peekTokenIs(token.ASYNC) {
		p.nextToken()
		fn.Async = true
	}

	if p.peekTokenIs(token.CONTRACT) {
		p.nextToken()
		fn.Contract = p.parseContractBlock()
	}

	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		fn.Body = p.parseBlockStatement()
		if fn.Body != nil {
			fn.Finish = fn.Body.End()
		}
	} else if p.peekTokenIs(token.ASSIGN) || p.peekTokenIs(token.ARROW) {
		p.nextToken()
		p.nextToken()
		fn.IsExprBody = true
		fn.BodyExpr = p.parseExpression(LOWEST)
		if fn.BodyExpr != nil {
			fn.Finish = fn.BodyExpr.End()
		}
		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
			fn.Finish = p.curToken.End
		}
	} else {
		fn.Finish = p.curToken.End
	}

	return fn
}

func (p *Parser) parseClassDeclaration() ast.Statement {
	class := &ast.ClassDeclaration{Start: p.curToken.Pos}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	class.Name = p.currentIdentifier()

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	class.Params = p.parseParameterList(token.RPAREN)

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
		if !p.expectPeek(token.IDENT) {
			return class
		}
		class.SuperClass = p.currentIdentifier()
	}

	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		class.Body = p.parseBlockStatement()
		if class.Body != nil {
			class.Finish = class.Body.End()
		}
	} else {
		class.Finish = p.curToken.End
	}
	return class
}

func (p *Parser) parseStructDeclaration() ast.Statement {
	st := &ast.StructDeclaration{Start: p.curToken.Pos}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	st.Name = p.currentIdentifier()

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	st.Params = p.parseParameterList(token.RPAREN)

	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		st.Body = p.parseBlockStatement()
		if st.Body != nil {
			st.Finish = st.Body.End()
		}
	} else {
		st.Finish = p.curToken.End
	}
	return st
}

func (p *Parser) parseEnumDeclaration() ast.Statement {
	enumNode := &ast.EnumDeclaration{Start: p.curToken.Pos}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	enumNode.Name = p.currentIdentifier()

	if p.peekTokenIs(token.LT) {
		enumNode.TypeParams = p.parseTypeParameters()
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		caseNode := p.parseEnumCase()
		if caseNode != nil {
			enumNode.Cases = append(enumNode.Cases, *caseNode)
			enumNode.Finish = caseNode.End()
		}
		p.nextToken()
	}
	if p.curTokenIs(token.RBRACE) {
		enumNode.Finish = p.curToken.End
	}
	return enumNode
}

func (p *Parser) parseEnumCase() *ast.EnumCase {
	caseNode := &ast.EnumCase{Start: p.curToken.Pos}
	if p.curToken.Type != token.IDENT {
		p.errors = append(p.errors, fmt.Sprintf("expected enum case identifier, got %s", p.curToken.Type))
		return nil
	}
	caseNode.Name = p.currentIdentifier()

	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		p.nextToken()
		caseNode.Params = p.parseParameterList(token.RPAREN)
	}

	if !p.expectPeek(token.SEMICOLON) {
		return caseNode
	}
	caseNode.Finish = p.curToken.End
	return caseNode
}

func (p *Parser) parseContractDeclaration() ast.Statement {
	contract := &ast.ContractDeclaration{Start: p.curToken.Pos}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	contract.Name = p.currentIdentifier()

	if !p.expectPeek(token.LBRACE) {
		return contract
	}
	contract.Body = p.parseBlockStatement()
	if contract.Body != nil {
		contract.Finish = contract.Body.End()
	}
	return contract
}

func (p *Parser) parseImportDeclaration() ast.Statement {
	imp := &ast.ImportDeclaration{Start: p.curToken.Pos}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	imp.Path = append(imp.Path, p.currentIdentifier())
	for p.peekTokenIs(token.DOT) {
		p.nextToken()
		if !p.expectPeek(token.IDENT) {
			return imp
		}
		imp.Path = append(imp.Path, p.currentIdentifier())
	}

	if p.peekTokenIs(token.AS) {
		p.nextToken()
		if !p.expectPeek(token.IDENT) {
			return imp
		}
		imp.Alias = p.currentIdentifier()
	}

	if !p.expectPeek(token.SEMICOLON) {
		return imp
	}
	imp.Finish = p.curToken.End
	return imp
}

func (p *Parser) parseMatchStatement() ast.Statement {
	match := &ast.MatchStatement{Start: p.curToken.Pos}
	p.nextToken()
	match.Value = p.parseExpression(LOWEST)
	if !p.expectPeek(token.LBRACE) {
		return match
	}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		caseNode := p.parseMatchCase()
		if caseNode != nil {
			match.Cases = append(match.Cases, *caseNode)
			match.Finish = caseNode.End()
		}
		p.nextToken()
	}
	if p.curTokenIs(token.RBRACE) {
		match.Finish = p.curToken.End
	}
	return match
}

func (p *Parser) parseMatchCase() *ast.MatchCase {
	caseNode := &ast.MatchCase{Start: p.curToken.Pos}
	pattern := p.parsePattern()
	caseNode.Pattern = pattern

	if !p.expectPeek(token.ARROW) {
		return caseNode
	}
	p.nextToken()
	body := p.parseStatement()
	caseNode.Body = body
	if body != nil {
		caseNode.Finish = body.End()
	}
	return caseNode
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Start: p.curToken.Pos}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	block.Finish = p.curToken.End
	return block
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Start: p.curToken.Pos}
	stmt.Expression = p.parseExpression(LOWEST)
	if stmt.Expression != nil {
		stmt.Finish = stmt.Expression.End()
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
		stmt.Finish = p.curToken.End
	}
	return stmt
}

func (p *Parser) parseTypeParameters() []*ast.Identifier {
	params := []*ast.Identifier{}
	if !p.expectPeek(token.LT) {
		return params
	}
	if !p.expectPeek(token.IDENT) {
		return params
	}
	params = append(params, p.currentIdentifier())
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if !p.expectPeek(token.IDENT) {
			return params
		}
		params = append(params, p.currentIdentifier())
	}
	if !p.expectPeek(token.GT) {
		return params
	}
	return params
}

func (p *Parser) parseParameterList(end token.Type) []ast.Parameter {
	params := []ast.Parameter{}
	if p.curTokenIs(end) {
		return params
	}
	param := p.parseParameter()
	params = append(params, param)
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		params = append(params, p.parseParameter())
	}
	if !p.expectPeek(end) {
		return params
	}
	return params
}

func (p *Parser) parseParameter() ast.Parameter {
	param := ast.Parameter{}
	if p.curToken.Type != token.IDENT {
		p.errors = append(p.errors, fmt.Sprintf("expected parameter name, got %s", p.curToken.Type))
		return param
	}
	param.Name = p.currentIdentifier()
	if !p.expectPeek(token.COLON) {
		return param
	}
	p.nextToken()
	param.Type = p.parseTypeAnnotation()
	return param
}

func (p *Parser) parseTypeAnnotation() *ast.TypeAnnotation {
	if p.curToken.Type != token.IDENT {
		p.errors = append(p.errors, fmt.Sprintf("expected type identifier, got %s", p.curToken.Type))
		return nil
	}
	typeNode := &ast.TypeAnnotation{Start: p.curToken.Pos}
	typeNode.Name = p.currentIdentifier()

	if p.peekTokenIs(token.LT) {
		p.nextToken()
		p.nextToken()
		if arg := p.parseTypeAnnotation(); arg != nil {
			typeNode.TypeArgs = append(typeNode.TypeArgs, arg)
		}
		for p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			if arg := p.parseTypeAnnotation(); arg != nil {
				typeNode.TypeArgs = append(typeNode.TypeArgs, arg)
			}
		}
		if !p.expectPeek(token.GT) {
			return typeNode
		}
	}

	if p.peekTokenIs(token.QUESTION) {
		p.nextToken()
		typeNode.Nullable = true
	}

	typeNode.Finish = p.curToken.End
	return typeNode
}

func (p *Parser) parseContractBlock() *ast.ContractBlock {
	block := &ast.ContractBlock{Start: p.curToken.Pos}
	if !p.expectPeek(token.LBRACE) {
		return block
	}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		clause := p.parseContractClause()
		if clause != nil {
			block.Clauses = append(block.Clauses, *clause)
		}
		p.nextToken()
	}
	block.Finish = p.curToken.End
	return block
}

func (p *Parser) parseContractClause() *ast.ContractClause {
	clause := &ast.ContractClause{Start: p.curToken.Pos}
	if p.curToken.Type != token.RETURNS {
		p.errors = append(p.errors, fmt.Sprintf("expected 'returns' in contract clause, got %s", p.curToken.Type))
		return nil
	}
	if !p.expectPeek(token.LPAREN) {
		return clause
	}
	if !p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		clause.Guard = p.parseExpression(LOWEST)
	}
	if !p.expectPeek(token.RPAREN) {
		return clause
	}
	if !p.expectPeek(token.ARROW) {
		return clause
	}
	p.nextToken()
	clause.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token.SEMICOLON) {
		return clause
	}
	clause.Finish = p.curToken.End
	return clause
}

func (p *Parser) parsePattern() ast.Pattern {
	switch p.curToken.Type {
	case token.NUMBER:
		node := &ast.NumberLiteral{Value: p.curToken.Literal, Start: p.curToken.Pos, Finish: p.curToken.End}
		return &ast.LiteralPattern{Value: node}
	case token.STRING:
		node := &ast.StringLiteral{Value: p.curToken.Literal, Start: p.curToken.Pos, Finish: p.curToken.End}
		return &ast.LiteralPattern{Value: node}
	case token.TRUE, token.FALSE:
		node := &ast.BooleanLiteral{Value: p.curToken.Type == token.TRUE, Start: p.curToken.Pos, Finish: p.curToken.End}
		return &ast.LiteralPattern{Value: node}
	case token.NULL:
		node := &ast.NullLiteral{Start: p.curToken.Pos, Finish: p.curToken.End}
		return &ast.LiteralPattern{Value: node}
	case token.IDENT:
		ident := p.currentIdentifier()
		if p.peekTokenIs(token.LPAREN) {
			return p.parseStructPattern(ident)
		}
		return &ast.IdentifierPattern{Identifier: ident}
	case token.LBRACE:
		return p.parseObjectPattern()
	default:
		p.errors = append(p.errors, fmt.Sprintf("unexpected token in pattern: %s", p.curToken.Type))
		return nil
	}
}

func (p *Parser) parseStructPattern(name *ast.Identifier) ast.Pattern {
	pattern := &ast.StructPattern{Name: name, Start: name.Pos()}
	if !p.expectPeek(token.LPAREN) {
		return pattern
	}
	p.nextToken()
	if p.curTokenIs(token.RPAREN) {
		pattern.Finish = p.curToken.End
		return pattern
	}
	pattern.Fields = append(pattern.Fields, p.parsePattern())
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		pattern.Fields = append(pattern.Fields, p.parsePattern())
	}
	if !p.expectPeek(token.RPAREN) {
		return pattern
	}
	pattern.Finish = p.curToken.End
	return pattern
}

func (p *Parser) parseObjectPattern() ast.Pattern {
	pattern := &ast.ObjectPattern{Start: p.curToken.Pos}
	p.nextToken()
	if p.curTokenIs(token.RBRACE) {
		pattern.Finish = p.curToken.End
		return pattern
	}
	pattern.Pairs = append(pattern.Pairs, p.parsePatternPair())
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		pattern.Pairs = append(pattern.Pairs, p.parsePatternPair())
	}
	if !p.expectPeek(token.RBRACE) {
		return pattern
	}
	pattern.Finish = p.curToken.End
	return pattern
}

func (p *Parser) parsePatternPair() ast.PatternPair {
	pair := ast.PatternPair{}
	keyToken := p.curToken
	if keyToken.Type != token.STRING && keyToken.Type != token.IDENT {
		p.errors = append(p.errors, fmt.Sprintf("expected pattern key, got %s", keyToken.Type))
		return pair
	}
	pair.Key = keyToken.Literal
	pair.KeyIsIdentifier = keyToken.Type == token.IDENT
	if !p.expectPeek(token.COLON) {
		return pair
	}
	p.nextToken()
	pair.Value = p.parsePattern()
	return pair
}

func (p *Parser) parseIdentifier() ast.Expression {
	return p.currentIdentifier()
}

func (p *Parser) parseNumberLiteral() ast.Expression {
	lit := &ast.NumberLiteral{Value: p.curToken.Literal, Start: p.curToken.Pos, Finish: p.curToken.End}
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Value: p.curToken.Literal, Start: p.curToken.Pos, Finish: p.curToken.End}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Value: p.curToken.Type == token.TRUE, Start: p.curToken.Pos, Finish: p.curToken.End}
}

func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Start: p.curToken.Pos, Finish: p.curToken.End}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{Operator: p.curToken.Literal, Start: p.curToken.Pos}
	p.nextToken()
	expr.Right = p.parseExpression(PREFIX)
	if expr.Right != nil {
		expr.Finish = expr.Right.End()
	} else {
		expr.Finish = p.curToken.End
	}
	return expr
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Start: p.curToken.Pos}
	if p.peekTokenIs(token.RBRACKET) {
		p.nextToken()
		array.Finish = p.curToken.End
		return array
	}
	p.nextToken()
	array.Elements = append(array.Elements, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		array.Elements = append(array.Elements, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(token.RBRACKET) {
		return array
	}
	array.Finish = p.curToken.End
	return array
}

func (p *Parser) parseObjectLiteral() ast.Expression {
	obj := &ast.ObjectLiteral{Start: p.curToken.Pos}
	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		obj.Finish = p.curToken.End
		return obj
	}
	p.nextToken()
	obj.Pairs = append(obj.Pairs, p.parseObjectPair())
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		obj.Pairs = append(obj.Pairs, p.parseObjectPair())
	}
	if !p.expectPeek(token.RBRACE) {
		return obj
	}
	obj.Finish = p.curToken.End
	return obj
}

func (p *Parser) parseObjectPair() ast.ObjectPair {
	pair := ast.ObjectPair{}
	keyToken := p.curToken
	if keyToken.Type != token.STRING && keyToken.Type != token.IDENT {
		p.errors = append(p.errors, fmt.Sprintf("expected object key, got %s", keyToken.Type))
		return pair
	}
	pair.Key = keyToken.Literal
	pair.KeyIsIdentifier = keyToken.Type == token.IDENT
	if !p.expectPeek(token.COLON) {
		return pair
	}
	p.nextToken()
	pair.Value = p.parseExpression(LOWEST)
	return pair
}

func (p *Parser) parseAwaitExpression() ast.Expression {
	expr := &ast.AwaitExpression{Start: p.curToken.Pos}
	p.nextToken()
	expr.Expression = p.parseExpression(PREFIX)
	if expr.Expression != nil {
		expr.Finish = expr.Expression.End()
	}
	return expr
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{Left: left, Operator: p.curToken.Literal, Start: left.Pos()}
	precedence := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(precedence)
	if expr.Right != nil {
		expr.Finish = expr.Right.End()
	} else {
		expr.Finish = p.curToken.End
	}
	return expr
}

func (p *Parser) parseElvisExpression(left ast.Expression) ast.Expression {
	expr := &ast.ElvisExpression{Left: left, Start: left.Pos()}
	p.nextToken()
	expr.Right = p.parseExpression(ELVIS - 1)
	if expr.Right != nil {
		expr.Finish = expr.Right.End()
	}
	return expr
}

func (p *Parser) parseAssignmentExpression(left ast.Expression) ast.Expression {
	expr := &ast.AssignmentExpression{Target: left, Start: left.Pos()}
	precedence := p.curPrecedence()
	p.nextToken()
	expr.Value = p.parseExpression(precedence - 1)
	if expr.Value != nil {
		expr.Finish = expr.Value.End()
	}
	return expr
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Callee: function, Start: function.Pos()}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	exp.Finish = p.curToken.End
	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Collection: left, Start: left.Pos()}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		return exp
	}
	exp.Finish = p.curToken.End
	return exp
}

func (p *Parser) parseMemberExpression(object ast.Expression) ast.Expression {
	member := &ast.MemberExpression{Object: object, Optional: p.curToken.Type == token.SAFE_DOT, Start: object.Pos()}
	if !p.expectPeek(token.IDENT) {
		return member
	}
	member.Property = p.curToken.Literal
	member.Finish = p.curToken.End
	return member
}

func (p *Parser) parseNonNullAssertion(left ast.Expression) ast.Expression {
	assertion := &ast.NonNullAssertion{Expression: left, Start: left.Pos(), Finish: p.curToken.End}
	return assertion
}

func (p *Parser) parseExpressionList(end token.Type) []ast.Expression {
	list := []ast.Expression{}
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}
	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(end) {
		return list
	}
	return list
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			break
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) currentIdentifier() *ast.Identifier {
	return &ast.Identifier{Name: p.curToken.Literal, Start: p.curToken.Pos, Finish: p.curToken.End}
}

func (p *Parser) expectPeek(t token.Type) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) curTokenIs(t token.Type) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) peekError(t token.Type) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.Type) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}
