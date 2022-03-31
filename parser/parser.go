package parser

import (
	"fmt"
	"scream/ast"
	"scream/lexer"
	"scream/token"
	"strconv"
	"strings"
)

type (
	prefixParseFn  func() ast.Expression
	infixParseFn   func(ast.Expression) ast.Expression
	postfixParseFn func() ast.Expression
)

const (
	_ int = iota
	LOWEST
	COND         // OR or AND
	ASSIGN       // =
	TERNARY      // ? :
	EQUALS       // == or !=
	REGEXP_MATCH // !~ ~=
	LESSGREATER  // > or <
	SUM          // + or -
	PRODUCT      // * or /
	POWER        // **
	MOD          // %
	PREFIX       // -X or !X
	CALL         // myFunction(X)
	DOTDOT       // ..
	INDEX        // array[index], map[key]
	HIGHEST
)

var precedences = map[token.Type]int{
	token.QUESTION:     TERNARY,
	token.ASSIGN:       ASSIGN,
	token.DOTDOT:       DOTDOT,
	token.EQ:           EQUALS,
	token.NOT_EQ:       EQUALS,
	token.LT:           LESSGREATER,
	token.LT_EQUALS:    LESSGREATER,
	token.GT:           LESSGREATER,
	token.GT_EQUALS:    LESSGREATER,
	token.CONTAINS:     REGEXP_MATCH,
	token.NOT_CONTAINS: REGEXP_MATCH,

	token.PLUS:            SUM,
	token.PLUS_EQUALS:     SUM,
	token.MINUS:           SUM,
	token.MINUS_EQUALS:    SUM,
	token.SLASH:           PRODUCT,
	token.SLASH_EQUALS:    PRODUCT,
	token.ASTERISK:        PRODUCT,
	token.ASTERISK_EQUALS: PRODUCT,
	token.POW:             POWER,
	token.MOD:             MOD,
	token.AND:             COND,
	token.OR:              COND,
	token.LPAREN:          CALL,
	token.PERIOD:          CALL,
	token.LBRACKET:        INDEX,
}

type Parser struct {
	l *lexer.Lexer

	prevToken token.Token
	curToken  token.Token

	peekToken token.Token

	errors         []string
	prefixParseFns map[token.Type]prefixParseFn

	infixParseFns map[token.Type]infixParseFn

	postfixParseFns map[token.Type]postfixParseFn

	tern bool
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	p.nextToken()
	p.nextToken()
	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.registerPrefix(token.BACKTICK, p.parseBacktickLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.DEFINE_FUNCTION, p.parseFunctionDefinition)
	p.registerPrefix(token.EOF, p.parsingBroken)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.FOR, p.parseForLoopExpression)
	p.registerPrefix(token.FOREACH, p.parseForEach)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.ILLEGAL, p.parsingBroken)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.NULL, p.parseNull)
	p.registerPrefix(token.REGEXP, p.parseRegexpLiteral)
	p.registerPrefix(token.REGEXP, p.parseRegexpLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.SWITCH, p.parseSwitchStatement)
	p.registerPrefix(token.TRUE, p.parseBoolean)

	p.infixParseFns = make(map[token.Type]infixParseFn)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK_EQUALS, p.parseAssignExpression)
	p.registerInfix(token.CONTAINS, p.parseInfixExpression)
	p.registerInfix(token.DOTDOT, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GT_EQUALS, p.parseInfixExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.LT_EQUALS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS_EQUALS, p.parseAssignExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)
	p.registerInfix(token.NOT_CONTAINS, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.PERIOD, p.parseMethodCallExpression)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.PLUS_EQUALS, p.parseAssignExpression)
	p.registerInfix(token.POW, p.parseInfixExpression)
	p.registerInfix(token.QUESTION, p.parseTernaryExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.SLASH_EQUALS, p.parseAssignExpression)

	p.postfixParseFns = make(map[token.Type]postfixParseFn)
	p.registerPostfix(token.MINUS_MINUS, p.parsePostfixExpression)
	p.registerPostfix(token.PLUS_PLUS, p.parsePostfixExpression)

	return p
}

func (p *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}
func (p *Parser) registerInfix(tokenType token.Type, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) registerPostfix(tokenType token.Type, fn postfixParseFn) {
	p.postfixParseFns[tokenType] = fn
}
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.Type) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead around line %d", t, p.curToken.Type, p.l.GetLine())
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.prevToken = p.curToken
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.CONST:
		return p.parseConstStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.ASSIGN) {

		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	for !p.curTokenIs(token.SEMICOLON) {

		if p.curTokenIs(token.EOF) {
			p.errors = append(p.errors, "unterminated let statement")
			return nil
		}

		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseConstStatement() *ast.ConstStatement {
	stmt := &ast.ConstStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	for !p.curTokenIs(token.SEMICOLON) {

		if p.curTokenIs(token.EOF) {
			p.errors = append(p.errors, "unterminated const statement")
			return nil
		}

		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)
	for !p.curTokenIs(token.SEMICOLON) {

		if p.curTokenIs(token.EOF) {
			p.errors = append(p.errors, "unterminated return statement")
			return nil
		}

		p.nextToken()
	}
	return stmt
}

func (p *Parser) noPrefixParseFnError(t token.Type) {
	msg := fmt.Sprintf("no prefix parse function for %s found around line %d", t, p.l.GetLine())
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	postfix := p.postfixParseFns[p.curToken.Type]
	if postfix != nil {
		return (postfix())
	}
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) parsingBroken() ast.Expression {
	return nil
}
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	var value int64
	var err error

	if strings.HasPrefix(p.curToken.Literal, "0b") {
		value, err = strconv.ParseInt(p.curToken.Literal[2:], 2, 64)
	} else if strings.HasPrefix(p.curToken.Literal, "0x") {
		value, err = strconv.ParseInt(p.curToken.Literal[2:], 16, 64)
	} else {
		value, err = strconv.ParseInt(p.curToken.Literal, 10, 64)
	}

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer around line %d", p.curToken.Literal, p.l.GetLine())
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	flo := &ast.FloatLiteral{Token: p.curToken}
	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float around line %d", p.curToken.Literal, p.l.GetLine())
		p.errors = append(p.errors, msg)
		return nil
	}
	flo.Value = value
	return flo
}

func (p *Parser) parseSwitchStatement() ast.Expression {

	expression := &ast.SwitchExpression{Token: p.curToken}

	expression.Value = p.parseBracketExpression()
	if expression.Value == nil {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) {

		if p.curTokenIs(token.EOF) {
			p.errors = append(p.errors, "unterminated switch statement")
			return nil
		}
		tmp := &ast.CaseExpression{Token: p.curToken}
		if p.curTokenIs(token.DEFAULT) {

			tmp.Default = true

		} else if p.curTokenIs(token.CASE) {

			p.nextToken()

			if p.curTokenIs(token.DEFAULT) {
				tmp.Default = true
			} else {
				tmp.Expr = append(tmp.Expr, p.parseExpression(LOWEST))
				for p.peekTokenIs(token.COMMA) {
					p.nextToken()

					p.nextToken()

					tmp.Expr = append(tmp.Expr, p.parseExpression(LOWEST))
				}
			}
		} else {
			p.errors = append(p.errors, fmt.Sprintf("expected case|default, got %s", p.curToken.Type))
			return nil
		}

		if !p.expectPeek(token.LBRACE) {

			msg := fmt.Sprintf("expected token to be '{', got %s instead", p.curToken.Type)
			p.errors = append(p.errors, msg)
			fmt.Printf("error\n")
			return nil
		}

		tmp.Block = p.parseBlockStatement()

		if !p.curTokenIs(token.RBRACE) {
			msg := fmt.Sprintf("Syntax Error: expected token to be '}', got %s instead", p.curToken.Type)
			p.errors = append(p.errors, msg)
			fmt.Printf("error\n")
			return nil

		}
		p.nextToken()
		expression.Choices = append(expression.Choices, tmp)
	}

	count := 0
	for _, c := range expression.Choices {
		if c.Default {
			count++
		}
	}
	if count > 1 {
		msg := "A switch-statement should only have one default block"
		p.errors = append(p.errors, msg)
		return nil

	}
	return expression

}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}
func (p *Parser) parseNull() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
}
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parsePostfixExpression() ast.Expression {
	expression := &ast.PostfixExpression{
		Token:    p.prevToken,
		Operator: p.curToken.Literal,
	}
	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) parseTernaryExpression(condition ast.Expression) ast.Expression {
	if p.tern {
		msg := fmt.Sprintf("nested ternary expressions are illegal, around line %d", p.l.GetLine())
		p.errors = append(p.errors, msg)
		return nil
	}

	p.tern = true
	defer func() { p.tern = false }()

	expression := &ast.TernaryExpression{
		Token:     p.curToken,
		Condition: condition,
	}
	p.nextToken() //skip the '?'
	precedence := p.curPrecedence()
	expression.IfTrue = p.parseExpression(precedence)

	if !p.expectPeek(token.COLON) { //skip the ":"
		return nil
	}

	p.nextToken()
	expression.IfFalse = p.parseExpression(precedence)

	p.tern = false
	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}
	if expression == nil {
		p.errors = append(p.errors, "unexpected nil expression")
		return nil
	}

	expression.Condition = p.parseBracketExpression()

	if expression.Condition == nil {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		msg := fmt.Sprintf("expected '{' but got %s", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	expression.Consequence = p.parseBlockStatement()
	if expression.Consequence == nil {
		p.errors = append(p.errors, "unexpected nil expression")
		return nil
	}

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if p.peekTokenIs(token.IF) {

			p.nextToken()

			expression.Alternative = &ast.BlockStatement{
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: p.parseIfExpression(),
					},
				},
			}

			return expression
		}

		if !p.expectPeek(token.LBRACE) {
			msg := fmt.Sprintf("expected '{' but got %s", p.curToken.Literal)
			p.errors = append(p.errors, msg)
			return nil
		}
		expression.Alternative = p.parseBlockStatement()
		if expression.Alternative == nil {
			p.errors = append(p.errors, "unexpected nil expression")
			return nil
		}
	}
	return expression
}

func (p *Parser) parseBracketExpression() ast.Expression {

	if !p.expectPeek(token.LPAREN) {
		msg := fmt.Sprintf("expected '(' but got %s", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()

	tmp := p.parseExpression(LOWEST)
	if tmp == nil {
		return nil
	}
	if !p.expectPeek(token.RPAREN) {
		msg := fmt.Sprintf("expected ')' but got %s", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	return tmp
}

func (p *Parser) parseForLoopExpression() ast.Expression {
	expression := &ast.ForLoopExpression{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Consequence = p.parseBlockStatement()
	return expression
}

func (p *Parser) parseForEach() ast.Expression {
	expression := &ast.ForeachStatement{Token: p.curToken}

	p.nextToken()
	expression.Ident = p.curToken.Literal

	if p.peekTokenIs(token.COMMA) {
		p.nextToken()

		if !p.peekTokenIs(token.IDENT) {
			p.errors = append(p.errors, fmt.Sprintf("second argument to foreach must be ident, got %v", p.peekToken))
			return nil
		}
		p.nextToken()

		expression.Index = expression.Ident
		expression.Ident = p.curToken.Literal

	}
	if !p.expectPeek(token.IN) {
		return nil
	}
	p.nextToken()

	expression.Value = p.parseExpression(LOWEST)
	if expression.Value == nil {
		return nil
	}
	p.nextToken()
	expression.Body = p.parseBlockStatement()

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {

	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) {

		if p.curTokenIs(token.EOF) {
			p.errors = append(p.errors,
				"unterminated block statement")
			return nil
		}

		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	lit.Defaults, lit.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()
	return lit
}

func (p *Parser) parseFunctionDefinition() ast.Expression {
	p.nextToken()
	lit := &ast.FunctionDefineLiteral{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Defaults, lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()

	return lit
}
func (p *Parser) parseFunctionParameters() (map[string]ast.Expression, []*ast.Identifier) {

	m := make(map[string]ast.Expression)

	identifiers := make([]*ast.Identifier, 0)

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return m, identifiers
	}
	p.nextToken()

	for !p.curTokenIs(token.RPAREN) {

		if p.curTokenIs(token.EOF) {
			p.errors = append(p.errors, "unterminated function parameters")
			return nil, nil
		}
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
		p.nextToken()
		if p.curTokenIs(token.ASSIGN) {
			p.nextToken()
			m[ident.Value] = p.parseExpressionStatement().Expression
			p.nextToken()
		}
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	return m, identifiers
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseRegexpLiteral() ast.Expression {
	flags := ""
	val := p.curToken.Literal
	if strings.HasPrefix(val, "(?") {
		val = strings.TrimPrefix(val, "(?")

		i := 0
		for i < len(val) {

			if val[i] == ')' {

				val = val[i+1:]
				break
			} else {
				flags += string(val[i])
			}

			i++
		}
	}
	return &ast.RegexpLiteral{Token: p.curToken, Value: val, Flags: flags}
}

func (p *Parser) parseBacktickLiteral() ast.Expression {
	return &ast.BacktickLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseExpressionList(end token.Type) []ast.Expression {
	list := make([]ast.Expression, 0)
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
		return nil
	}
	return list
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) parseAssignExpression(name ast.Expression) ast.Expression {
	stmt := &ast.AssignStatement{Token: p.curToken}
	if n, ok := name.(*ast.Identifier); ok {
		stmt.Name = n
	} else {
		msg := fmt.Sprintf("expected assign token to be IDENT, got %s instead around line %d", name.TokenLiteral(), p.l.GetLine())
		p.errors = append(p.errors, msg)
	}

	oper := p.curToken
	p.nextToken()

	switch oper.Type {
	case token.PLUS_EQUALS:
		stmt.Operator = "+="
	case token.MINUS_EQUALS:
		stmt.Operator = "-="
	case token.SLASH_EQUALS:
		stmt.Operator = "/="
	case token.ASTERISK_EQUALS:
		stmt.Operator = "*="
	default:
		stmt.Operator = "="
	}
	stmt.Value = p.parseExpression(LOWEST)
	return stmt
}
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)
	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		value := p.parseExpression(LOWEST)
		hash.Pairs[key] = value
		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}
	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return hash
}

func (p *Parser) parseMethodCallExpression(obj ast.Expression) ast.Expression {
	methodCall := &ast.ObjectCallExpression{Token: p.curToken, Object: obj}
	p.nextToken()
	name := p.parseIdentifier()
	p.nextToken()
	methodCall.Call = p.parseCallExpression(name)
	return methodCall
}

func (p *Parser) curTokenIs(t token.Type) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.Type) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}

	p.peekError(t)
	return false
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
