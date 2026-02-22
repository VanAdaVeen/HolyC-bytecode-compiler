package parser

import (
	"fmt"
	"os"

	"holyc-compiler/pkg/lexer"
)

type Parser struct {
	lex    *lexer.Lexer
	cur    lexer.Token
	peek   lexer.Token
	Errors []string
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{lex: l}
	p.advance()
	p.advance()
	return p
}

func (p *Parser) advance() lexer.Token {
	prev := p.cur
	p.cur = p.peek
	p.peek = p.lex.NextToken()
	return prev
}

func (p *Parser) expect(t lexer.TokenType) lexer.Token {
	if p.cur.Type != t {
		p.errorf("expected token %d, got %d ('%s')", t, p.cur.Type, p.cur.Literal)
	}
	return p.advance()
}

func (p *Parser) errorf(format string, args ...any) {
	msg := fmt.Sprintf("%s:%d:%d: %s", p.lex.File, p.cur.Line, p.cur.Col, fmt.Sprintf(format, args...))
	p.Errors = append(p.Errors, msg)
	fmt.Fprintln(os.Stderr, msg)
}

func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.cur.Type == t {
			return true
		}
	}
	return false
}

func (p *Parser) Parse() *Program {
	prog := &Program{}
	for p.cur.Type != lexer.TOK_EOF {
		node := p.parseTopLevel()
		if node != nil {
			prog.Decls = append(prog.Decls, node)
		}
	}
	return prog
}

func (p *Parser) parseTopLevel() Node {
	if p.cur.Type == lexer.TOK_INCLUDE || p.cur.Type == lexer.TOK_DEFINE {
		p.advance()
		return nil
	}
	if lexer.IsType(p.cur.Type) {
		return p.parseDeclaration()
	}
	return p.parseStatement()
}

func (p *Parser) parseDeclaration() Node {
	typeName := p.cur.Literal
	p.advance()

	isPtr := false
	for p.cur.Type == lexer.TOK_STAR {
		isPtr = true
		typeName += " *"
		p.advance()
	}

	if p.cur.Type != lexer.TOK_IDENT {
		p.errorf("expected identifier after type")
		p.advance()
		return nil
	}

	name := p.cur.Literal
	p.advance()

	if p.cur.Type == lexer.TOK_LPAREN {
		return p.parseFuncDecl(typeName, name)
	}

	var init Node
	if p.cur.Type == lexer.TOK_ASSIGN {
		p.advance()
		init = p.parseExpression()
	}
	if p.cur.Type == lexer.TOK_SEMICOLON {
		p.advance()
	}
	return &VarDecl{TypeName: typeName, Name: name, Init: init, IsPtr: isPtr}
}

func (p *Parser) parseFuncDecl(retType, name string) Node {
	p.expect(lexer.TOK_LPAREN)
	var params []FuncParam
	for p.cur.Type != lexer.TOK_RPAREN && p.cur.Type != lexer.TOK_EOF {
		if len(params) > 0 {
			p.expect(lexer.TOK_COMMA)
		}
		params = append(params, p.parseFuncParam())
	}
	p.expect(lexer.TOK_RPAREN)
	body := p.parseBlock()
	return &FuncDecl{ReturnType: retType, Name: name, Params: params, Body: body}
}

func (p *Parser) parseFuncParam() FuncParam {
	param := FuncParam{}
	if lexer.IsType(p.cur.Type) {
		param.TypeName = p.cur.Literal
		p.advance()
		for p.cur.Type == lexer.TOK_STAR {
			param.TypeName += " *"
			p.advance()
		}
	}
	if p.cur.Type == lexer.TOK_IDENT {
		param.Name = p.cur.Literal
		p.advance()
	}
	if p.cur.Type == lexer.TOK_ASSIGN {
		p.advance()
		param.Default = p.parseExpression()
	}
	return param
}

func (p *Parser) parseBlock() *Block {
	p.expect(lexer.TOK_LBRACE)
	block := &Block{}
	for p.cur.Type != lexer.TOK_RBRACE && p.cur.Type != lexer.TOK_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Stmts = append(block.Stmts, stmt)
		}
	}
	p.expect(lexer.TOK_RBRACE)
	return block
}

func (p *Parser) parseStatement() Node {
	switch p.cur.Type {
	case lexer.TOK_LBRACE:
		return p.parseBlock()
	case lexer.TOK_RETURN:
		return p.parseReturn()
	case lexer.TOK_IF:
		return p.parseIf()
	case lexer.TOK_WHILE:
		return p.parseWhile()
	case lexer.TOK_FOR:
		return p.parseFor()
	case lexer.TOK_SEMICOLON:
		p.advance()
		return nil
	}
	if lexer.IsType(p.cur.Type) {
		return p.parseDeclaration()
	}
	expr := p.parseExpression()
	if p.cur.Type == lexer.TOK_SEMICOLON {
		p.advance()
	}
	return &ExprStmt{Expr: expr}
}

func (p *Parser) parseReturn() Node {
	p.advance()
	var val Node
	if p.cur.Type != lexer.TOK_SEMICOLON {
		val = p.parseExpression()
	}
	if p.cur.Type == lexer.TOK_SEMICOLON {
		p.advance()
	}
	return &ReturnStmt{Value: val}
}

func (p *Parser) parseIf() Node {
	p.advance()
	p.expect(lexer.TOK_LPAREN)
	cond := p.parseExpression()
	p.expect(lexer.TOK_RPAREN)
	body := p.parseStatement()
	var elseBody Node
	if p.cur.Type == lexer.TOK_ELSE {
		p.advance()
		elseBody = p.parseStatement()
	}
	return &IfStmt{Cond: cond, Body: body, Else: elseBody}
}

func (p *Parser) parseWhile() Node {
	p.advance()
	p.expect(lexer.TOK_LPAREN)
	cond := p.parseExpression()
	p.expect(lexer.TOK_RPAREN)
	body := p.parseStatement()
	return &WhileStmt{Cond: cond, Body: body}
}

func (p *Parser) parseFor() Node {
	p.advance()
	p.expect(lexer.TOK_LPAREN)
	var init Node
	if p.cur.Type != lexer.TOK_SEMICOLON {
		if lexer.IsType(p.cur.Type) {
			init = p.parseDeclaration()
		} else {
			init = p.parseExpression()
			p.expect(lexer.TOK_SEMICOLON)
		}
	} else {
		p.advance()
	}
	var cond Node
	if p.cur.Type != lexer.TOK_SEMICOLON {
		cond = p.parseExpression()
	}
	p.expect(lexer.TOK_SEMICOLON)
	var post Node
	if p.cur.Type != lexer.TOK_RPAREN {
		post = p.parseExpression()
	}
	p.expect(lexer.TOK_RPAREN)
	body := p.parseStatement()
	return &ForStmt{Init: init, Cond: cond, Post: post, Body: body}
}

// ---- Expression parsing (Pratt / precedence climbing) ----

func (p *Parser) parseExpression() Node { return p.parseAssign() }

func (p *Parser) parseAssign() Node {
	left := p.parseOr()
	if p.match(lexer.TOK_ASSIGN, lexer.TOK_PLUS_EQ, lexer.TOK_MINUS_EQ, lexer.TOK_STAR_EQ,
		lexer.TOK_SLASH_EQ, lexer.TOK_PERCENT_EQ, lexer.TOK_AMP_EQ, lexer.TOK_PIPE_EQ,
		lexer.TOK_CARET_EQ, lexer.TOK_SHL_EQ, lexer.TOK_SHR_EQ) {
		op := p.cur.Type
		p.advance()
		return &AssignExpr{Op: op, Target: left, Value: p.parseAssign()}
	}
	return left
}

func (p *Parser) parseOr() Node {
	left := p.parseAnd()
	for p.cur.Type == lexer.TOK_OR_OR {
		p.advance()
		left = &BinaryExpr{Op: lexer.TOK_OR_OR, Left: left, Right: p.parseAnd()}
	}
	return left
}

func (p *Parser) parseAnd() Node {
	left := p.parseBitOr()
	for p.cur.Type == lexer.TOK_AND_AND {
		p.advance()
		left = &BinaryExpr{Op: lexer.TOK_AND_AND, Left: left, Right: p.parseBitOr()}
	}
	return left
}

func (p *Parser) parseBitOr() Node {
	left := p.parseBitXor()
	for p.cur.Type == lexer.TOK_PIPE {
		p.advance()
		left = &BinaryExpr{Op: lexer.TOK_PIPE, Left: left, Right: p.parseBitXor()}
	}
	return left
}

func (p *Parser) parseBitXor() Node {
	left := p.parseBitAnd()
	for p.cur.Type == lexer.TOK_CARET {
		p.advance()
		left = &BinaryExpr{Op: lexer.TOK_CARET, Left: left, Right: p.parseBitAnd()}
	}
	return left
}

func (p *Parser) parseBitAnd() Node {
	left := p.parseEquality()
	for p.cur.Type == lexer.TOK_AMP {
		p.advance()
		left = &BinaryExpr{Op: lexer.TOK_AMP, Left: left, Right: p.parseEquality()}
	}
	return left
}

func (p *Parser) parseEquality() Node {
	left := p.parseComparison()
	for p.cur.Type == lexer.TOK_EQ || p.cur.Type == lexer.TOK_NEQ {
		op := p.cur.Type
		p.advance()
		left = &BinaryExpr{Op: op, Left: left, Right: p.parseComparison()}
	}
	return left
}

func (p *Parser) parseComparison() Node {
	left := p.parseShift()
	for p.match(lexer.TOK_LT, lexer.TOK_GT, lexer.TOK_LTE, lexer.TOK_GTE) {
		op := p.cur.Type
		p.advance()
		left = &BinaryExpr{Op: op, Left: left, Right: p.parseShift()}
	}
	return left
}

func (p *Parser) parseShift() Node {
	left := p.parseAddSub()
	for p.cur.Type == lexer.TOK_SHL || p.cur.Type == lexer.TOK_SHR {
		op := p.cur.Type
		p.advance()
		left = &BinaryExpr{Op: op, Left: left, Right: p.parseAddSub()}
	}
	return left
}

func (p *Parser) parseAddSub() Node {
	left := p.parseMulDiv()
	for p.cur.Type == lexer.TOK_PLUS || p.cur.Type == lexer.TOK_MINUS {
		op := p.cur.Type
		p.advance()
		left = &BinaryExpr{Op: op, Left: left, Right: p.parseMulDiv()}
	}
	return left
}

func (p *Parser) parseMulDiv() Node {
	left := p.parsePower()
	for p.match(lexer.TOK_STAR, lexer.TOK_SLASH, lexer.TOK_PERCENT) {
		op := p.cur.Type
		p.advance()
		left = &BinaryExpr{Op: op, Left: left, Right: p.parsePower()}
	}
	return left
}

func (p *Parser) parsePower() Node {
	left := p.parseUnary()
	if p.cur.Type == lexer.TOK_BACKTICK {
		p.advance()
		left = &BinaryExpr{Op: lexer.TOK_BACKTICK, Left: left, Right: p.parseUnary()}
	}
	return left
}

func (p *Parser) parseUnary() Node {
	switch p.cur.Type {
	case lexer.TOK_MINUS:
		p.advance()
		return &UnaryExpr{Op: lexer.TOK_MINUS, Operand: p.parseUnary()}
	case lexer.TOK_TILDE:
		p.advance()
		return &UnaryExpr{Op: lexer.TOK_TILDE, Operand: p.parseUnary()}
	case lexer.TOK_BANG:
		p.advance()
		return &UnaryExpr{Op: lexer.TOK_BANG, Operand: p.parseUnary()}
	case lexer.TOK_PLUS_PLUS:
		p.advance()
		return &UnaryExpr{Op: lexer.TOK_PLUS_PLUS, Operand: p.parseUnary()}
	case lexer.TOK_MINUS_MINUS:
		p.advance()
		return &UnaryExpr{Op: lexer.TOK_MINUS_MINUS, Operand: p.parseUnary()}
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() Node {
	left := p.parsePrimary()
	for {
		switch p.cur.Type {
		case lexer.TOK_PLUS_PLUS:
			p.advance()
			left = &PostfixExpr{Op: lexer.TOK_PLUS_PLUS, Operand: left}
		case lexer.TOK_MINUS_MINUS:
			p.advance()
			left = &PostfixExpr{Op: lexer.TOK_MINUS_MINUS, Operand: left}
		case lexer.TOK_LPAREN:
			if ident, ok := left.(*Identifier); ok {
				left = p.parseCallExpr(ident.Name)
			} else {
				return left
			}
		case lexer.TOK_LBRACKET:
			p.advance()
			index := p.parseExpression()
			p.expect(lexer.TOK_RBRACKET)
			left = &IndexExpr{Array: left, Index: index}
		case lexer.TOK_DOT:
			p.advance()
			left = &MemberExpr{Object: left, Member: p.expect(lexer.TOK_IDENT).Literal, Arrow: false}
		case lexer.TOK_ARROW:
			p.advance()
			left = &MemberExpr{Object: left, Member: p.expect(lexer.TOK_IDENT).Literal, Arrow: true}
		default:
			return left
		}
	}
}

func (p *Parser) parseCallExpr(name string) Node {
	p.expect(lexer.TOK_LPAREN)
	var args []Node
	for p.cur.Type != lexer.TOK_RPAREN && p.cur.Type != lexer.TOK_EOF {
		if len(args) > 0 {
			p.expect(lexer.TOK_COMMA)
		}
		args = append(args, p.parseExpression())
	}
	p.expect(lexer.TOK_RPAREN)
	return &CallExpr{Func: name, Args: args}
}

func (p *Parser) parsePrimary() Node {
	switch p.cur.Type {
	case lexer.TOK_INT, lexer.TOK_CHAR:
		val := p.cur.IntVal
		p.advance()
		return &IntLiteral{Value: val}
	case lexer.TOK_FLOAT:
		val := p.cur.FloatVal
		p.advance()
		return &FloatLiteral{Value: val}
	case lexer.TOK_STRING:
		val := p.cur.Literal
		p.advance()
		return &StringLiteral{Value: val}
	case lexer.TOK_IDENT:
		name := p.cur.Literal
		p.advance()
		return &Identifier{Name: name}
	case lexer.TOK_SIZEOF:
		p.advance()
		p.expect(lexer.TOK_LPAREN)
		typeName := p.cur.Literal
		p.advance()
		p.expect(lexer.TOK_RPAREN)
		return &SizeofExpr{TypeName: typeName}
	case lexer.TOK_LPAREN:
		p.advance()
		expr := p.parseExpression()
		p.expect(lexer.TOK_RPAREN)
		return expr
	}
	p.errorf("unexpected token in expression: '%s'", p.cur.Literal)
	p.advance()
	return &IntLiteral{Value: 0}
}
