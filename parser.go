package main

import (
	"fmt"
	"os"
)

type Parser struct {
	lexer  *Lexer
	cur    Token
	peek   Token
	errors []string
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{lexer: l}
	p.advance()
	p.advance()
	return p
}

func (p *Parser) advance() Token {
	prev := p.cur
	p.cur = p.peek
	p.peek = p.lexer.NextToken()
	return prev
}

func (p *Parser) expect(t TokenType) Token {
	if p.cur.Type != t {
		p.errorf("expected token %d, got %d ('%s')", t, p.cur.Type, p.cur.Literal)
	}
	return p.advance()
}

func (p *Parser) errorf(format string, args ...any) {
	msg := fmt.Sprintf("%s:%d:%d: %s", p.lexer.file, p.cur.Line, p.cur.Col, fmt.Sprintf(format, args...))
	p.errors = append(p.errors, msg)
	fmt.Fprintln(os.Stderr, msg)
}

func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.cur.Type == t {
			return true
		}
	}
	return false
}

// Parse parses an entire program (list of top-level declarations/statements).
func (p *Parser) Parse() *Program {
	prog := &Program{}
	for p.cur.Type != TOK_EOF {
		node := p.parseTopLevel()
		if node != nil {
			prog.Decls = append(prog.Decls, node)
		}
	}
	return prog
}

func (p *Parser) parseTopLevel() Node {
	// Skip preprocessor directives
	if p.cur.Type == TOK_INCLUDE || p.cur.Type == TOK_DEFINE {
		p.advance()
		return nil
	}
	// Type declaration → function or variable
	if IsType(p.cur.Type) {
		return p.parseDeclaration()
	}
	// Expression statement
	return p.parseStatement()
}

func (p *Parser) parseDeclaration() Node {
	typeName := p.cur.Literal
	p.advance()

	isPtr := false
	for p.cur.Type == TOK_STAR {
		isPtr = true
		typeName += " *"
		p.advance()
	}

	if p.cur.Type != TOK_IDENT {
		p.errorf("expected identifier after type")
		p.advance()
		return nil
	}

	name := p.cur.Literal
	p.advance()

	// Function declaration: I64 Foo(...)
	if p.cur.Type == TOK_LPAREN {
		return p.parseFuncDecl(typeName, name)
	}

	// Variable declaration: I64 x = expr;
	var init Node
	if p.cur.Type == TOK_ASSIGN {
		p.advance()
		init = p.parseExpression()
	}
	if p.cur.Type == TOK_SEMICOLON {
		p.advance()
	}
	return &VarDecl{TypeName: typeName, Name: name, Init: init, IsPtr: isPtr}
}

func (p *Parser) parseFuncDecl(retType, name string) Node {
	p.expect(TOK_LPAREN)
	var params []FuncParam
	for p.cur.Type != TOK_RPAREN && p.cur.Type != TOK_EOF {
		if len(params) > 0 {
			p.expect(TOK_COMMA)
		}
		param := p.parseFuncParam()
		params = append(params, param)
	}
	p.expect(TOK_RPAREN)

	body := p.parseBlock()
	return &FuncDecl{ReturnType: retType, Name: name, Params: params, Body: body}
}

func (p *Parser) parseFuncParam() FuncParam {
	param := FuncParam{}
	if IsType(p.cur.Type) {
		param.TypeName = p.cur.Literal
		p.advance()
		for p.cur.Type == TOK_STAR {
			param.TypeName += " *"
			p.advance()
		}
	}
	if p.cur.Type == TOK_IDENT {
		param.Name = p.cur.Literal
		p.advance()
	}
	// Default value: I64 x = 42
	if p.cur.Type == TOK_ASSIGN {
		p.advance()
		param.Default = p.parseExpression()
	}
	return param
}

func (p *Parser) parseBlock() *Block {
	p.expect(TOK_LBRACE)
	block := &Block{}
	for p.cur.Type != TOK_RBRACE && p.cur.Type != TOK_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Stmts = append(block.Stmts, stmt)
		}
	}
	p.expect(TOK_RBRACE)
	return block
}

func (p *Parser) parseStatement() Node {
	switch p.cur.Type {
	case TOK_LBRACE:
		return p.parseBlock()
	case TOK_RETURN:
		return p.parseReturn()
	case TOK_IF:
		return p.parseIf()
	case TOK_WHILE:
		return p.parseWhile()
	case TOK_FOR:
		return p.parseFor()
	case TOK_SEMICOLON:
		p.advance()
		return nil
	}
	// Variable declaration inside function
	if IsType(p.cur.Type) {
		return p.parseDeclaration()
	}
	// Expression statement
	expr := p.parseExpression()
	if p.cur.Type == TOK_SEMICOLON {
		p.advance()
	}
	return &ExprStmt{Expr: expr}
}

func (p *Parser) parseReturn() Node {
	p.advance() // skip 'return'
	var val Node
	if p.cur.Type != TOK_SEMICOLON {
		val = p.parseExpression()
	}
	if p.cur.Type == TOK_SEMICOLON {
		p.advance()
	}
	return &ReturnStmt{Value: val}
}

func (p *Parser) parseIf() Node {
	p.advance() // skip 'if'
	p.expect(TOK_LPAREN)
	cond := p.parseExpression()
	p.expect(TOK_RPAREN)
	body := p.parseStatement()
	var elseBody Node
	if p.cur.Type == TOK_ELSE {
		p.advance()
		elseBody = p.parseStatement()
	}
	return &IfStmt{Cond: cond, Body: body, Else: elseBody}
}

func (p *Parser) parseWhile() Node {
	p.advance() // skip 'while'
	p.expect(TOK_LPAREN)
	cond := p.parseExpression()
	p.expect(TOK_RPAREN)
	body := p.parseStatement()
	return &WhileStmt{Cond: cond, Body: body}
}

func (p *Parser) parseFor() Node {
	p.advance() // skip 'for'
	p.expect(TOK_LPAREN)
	var init Node
	if p.cur.Type != TOK_SEMICOLON {
		if IsType(p.cur.Type) {
			init = p.parseDeclaration()
		} else {
			init = p.parseExpression()
			p.expect(TOK_SEMICOLON)
		}
	} else {
		p.advance()
	}
	var cond Node
	if p.cur.Type != TOK_SEMICOLON {
		cond = p.parseExpression()
	}
	p.expect(TOK_SEMICOLON)
	var post Node
	if p.cur.Type != TOK_RPAREN {
		post = p.parseExpression()
	}
	p.expect(TOK_RPAREN)
	body := p.parseStatement()
	return &ForStmt{Init: init, Cond: cond, Post: post, Body: body}
}

// ---- Expression parsing (Pratt / precedence climbing) ----

func (p *Parser) parseExpression() Node {
	return p.parseAssign()
}

func (p *Parser) parseAssign() Node {
	left := p.parseOr()
	if p.match(TOK_ASSIGN, TOK_PLUS_EQ, TOK_MINUS_EQ, TOK_STAR_EQ,
		TOK_SLASH_EQ, TOK_PERCENT_EQ, TOK_AMP_EQ, TOK_PIPE_EQ,
		TOK_CARET_EQ, TOK_SHL_EQ, TOK_SHR_EQ) {
		op := p.cur.Type
		p.advance()
		val := p.parseAssign()
		return &AssignExpr{Op: op, Target: left, Value: val}
	}
	return left
}

func (p *Parser) parseOr() Node {
	left := p.parseAnd()
	for p.cur.Type == TOK_OR_OR {
		p.advance()
		right := p.parseAnd()
		left = &BinaryExpr{Op: TOK_OR_OR, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseAnd() Node {
	left := p.parseBitOr()
	for p.cur.Type == TOK_AND_AND {
		p.advance()
		right := p.parseBitOr()
		left = &BinaryExpr{Op: TOK_AND_AND, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseBitOr() Node {
	left := p.parseBitXor()
	for p.cur.Type == TOK_PIPE {
		p.advance()
		right := p.parseBitXor()
		left = &BinaryExpr{Op: TOK_PIPE, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseBitXor() Node {
	left := p.parseBitAnd()
	for p.cur.Type == TOK_CARET {
		p.advance()
		right := p.parseBitAnd()
		left = &BinaryExpr{Op: TOK_CARET, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseBitAnd() Node {
	left := p.parseEquality()
	for p.cur.Type == TOK_AMP {
		p.advance()
		right := p.parseEquality()
		left = &BinaryExpr{Op: TOK_AMP, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseEquality() Node {
	left := p.parseComparison()
	for p.cur.Type == TOK_EQ || p.cur.Type == TOK_NEQ {
		op := p.cur.Type
		p.advance()
		right := p.parseComparison()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseComparison() Node {
	left := p.parseShift()
	for p.match(TOK_LT, TOK_GT, TOK_LTE, TOK_GTE) {
		op := p.cur.Type
		p.advance()
		right := p.parseShift()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseShift() Node {
	left := p.parseAddSub()
	for p.cur.Type == TOK_SHL || p.cur.Type == TOK_SHR {
		op := p.cur.Type
		p.advance()
		right := p.parseAddSub()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseAddSub() Node {
	left := p.parseMulDiv()
	for p.cur.Type == TOK_PLUS || p.cur.Type == TOK_MINUS {
		op := p.cur.Type
		p.advance()
		right := p.parseMulDiv()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseMulDiv() Node {
	left := p.parsePower()
	for p.match(TOK_STAR, TOK_SLASH, TOK_PERCENT) {
		op := p.cur.Type
		p.advance()
		right := p.parsePower()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parsePower() Node {
	left := p.parseUnary()
	// HolyC uses ` for power: a`b = a^b
	if p.cur.Type == TOK_BACKTICK {
		p.advance()
		right := p.parseUnary()
		left = &BinaryExpr{Op: TOK_BACKTICK, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseUnary() Node {
	if p.cur.Type == TOK_MINUS {
		p.advance()
		operand := p.parseUnary()
		return &UnaryExpr{Op: TOK_MINUS, Operand: operand}
	}
	if p.cur.Type == TOK_TILDE {
		p.advance()
		operand := p.parseUnary()
		return &UnaryExpr{Op: TOK_TILDE, Operand: operand}
	}
	if p.cur.Type == TOK_BANG {
		p.advance()
		operand := p.parseUnary()
		return &UnaryExpr{Op: TOK_BANG, Operand: operand}
	}
	if p.cur.Type == TOK_PLUS_PLUS {
		p.advance()
		operand := p.parseUnary()
		return &UnaryExpr{Op: TOK_PLUS_PLUS, Operand: operand}
	}
	if p.cur.Type == TOK_MINUS_MINUS {
		p.advance()
		operand := p.parseUnary()
		return &UnaryExpr{Op: TOK_MINUS_MINUS, Operand: operand}
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() Node {
	left := p.parsePrimary()
	for {
		if p.cur.Type == TOK_PLUS_PLUS {
			p.advance()
			left = &PostfixExpr{Op: TOK_PLUS_PLUS, Operand: left}
		} else if p.cur.Type == TOK_MINUS_MINUS {
			p.advance()
			left = &PostfixExpr{Op: TOK_MINUS_MINUS, Operand: left}
		} else if p.cur.Type == TOK_LPAREN {
			// Function call on identifier
			if ident, ok := left.(*Identifier); ok {
				left = p.parseCallExpr(ident.Name)
			} else {
				break
			}
		} else if p.cur.Type == TOK_LBRACKET {
			p.advance()
			index := p.parseExpression()
			p.expect(TOK_RBRACKET)
			left = &IndexExpr{Array: left, Index: index}
		} else if p.cur.Type == TOK_DOT {
			p.advance()
			member := p.expect(TOK_IDENT).Literal
			left = &MemberExpr{Object: left, Member: member, Arrow: false}
		} else if p.cur.Type == TOK_ARROW {
			p.advance()
			member := p.expect(TOK_IDENT).Literal
			left = &MemberExpr{Object: left, Member: member, Arrow: true}
		} else {
			break
		}
	}
	return left
}

func (p *Parser) parseCallExpr(name string) Node {
	p.expect(TOK_LPAREN)
	var args []Node
	for p.cur.Type != TOK_RPAREN && p.cur.Type != TOK_EOF {
		if len(args) > 0 {
			p.expect(TOK_COMMA)
		}
		args = append(args, p.parseExpression())
	}
	p.expect(TOK_RPAREN)
	return &CallExpr{Func: name, Args: args}
}

func (p *Parser) parsePrimary() Node {
	switch p.cur.Type {
	case TOK_INT, TOK_CHAR:
		val := p.cur.IntVal
		p.advance()
		return &IntLiteral{Value: val}
	case TOK_FLOAT:
		val := p.cur.FloatVal
		p.advance()
		return &FloatLiteral{Value: val}
	case TOK_STRING:
		val := p.cur.Literal
		p.advance()
		return &StringLiteral{Value: val}
	case TOK_IDENT:
		name := p.cur.Literal
		p.advance()
		return &Identifier{Name: name}
	case TOK_SIZEOF:
		p.advance()
		p.expect(TOK_LPAREN)
		typeName := p.cur.Literal
		p.advance()
		p.expect(TOK_RPAREN)
		return &SizeofExpr{TypeName: typeName}
	case TOK_LPAREN:
		p.advance()
		expr := p.parseExpression()
		p.expect(TOK_RPAREN)
		return expr
	}
	p.errorf("unexpected token in expression: '%s'", p.cur.Literal)
	p.advance()
	return &IntLiteral{Value: 0}
}
