package syntax

import (
	"fmt"
)

/*
the precedence of the operators is as follows, from lowest to highest:

Operator    	          Associativity
Equality:    == !=	       Left
Comparison:  > >= < <=	   Left
Term: 	     - +	       Left
Factor: 	 / *	       Left
Unary: 	     ! -	       Right
*/

/*
expr -> equality
equality -> comparison ( ( "!=" | "==" ) comparison )*
comparison -> term ( ( ">" | ">=" | "<" | "<=" ) term )*
term -> factor ( ( "-" | "+" ) factor )*
factor -> unary ( ( "/" | "*" ) unary )*
unary -> ( "!" | "-" ) unary | primary
primary -> NUMBER | STRING | "false" | "true" | "nil" | "(" expr ")"
*/

type Parser struct {
	Tokens   []Token
	Current  int
	parseErr error
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		Tokens:  tokens,
		Current: 0,
	}
}

func (p *Parser) Parse() Expr {
	return p.parseExpr()
}

func (p *Parser) parseExpr() Expr {
	return p.parseEquality()
}

func (p *Parser) parseEquality() Expr {
	expr := p.parseComparison()

	for p.match(TOKEN_BANG_EQUAL, TOKEN_EQUAL_EQUAL) {
		op := p.previous()
		right := p.parseComparison()
		expr = NewBinary(expr, op, right)
	}

	return expr
}

func (p *Parser) parseComparison() Expr {
	expr := p.parseTerm()

	for p.match(TOKEN_GREATER, TOKEN_GREATER_EQUAL, TOKEN_LESS, TOKEN_LESS_EQUAL) {
		op := p.previous()
		right := p.parseTerm()
		expr = NewBinary(expr, op, right)
	}

	return expr
}

func (p *Parser) parseTerm() Expr {
	expr := p.parseFactor()

	for p.match(TOKEN_MINUS, TOKEN_PLUS) {
		op := p.previous()
		right := p.parseFactor()
		expr = NewBinary(expr, op, right)
	}

	return expr
}

func (p *Parser) parseFactor() Expr {
	expr := p.parseUnary()

	if p.match(TOKEN_SLASH, TOKEN_STAR) {
		op := p.previous()
		right := p.parseUnary()
		return NewBinary(expr, op, right)
	}

	return expr
}

func (p *Parser) parseUnary() Expr {
	if p.match(TOKEN_BANG, TOKEN_MINUS) {
		op := p.previous()
		right := p.parseUnary()
		return NewUnary(right, op)
	}

	return p.parsePrimary()
}

func (p *Parser) parsePrimary() Expr {
	if p.match(TOKEN_NUMBER) {
		return &Literal{Value: p.previous().Literal}
	}
	if p.match(TOKEN_STRING) {
		return &Literal{Value: p.previous().Literal}
	}
	if p.match(TOKEN_TRUE) {
		return &Literal{Value: true}
	}
	if p.match(TOKEN_FALSE) {
		return &Literal{Value: false}
	}
	if p.match(TOKEN_NIL) {
		return &Literal{Value: nil}
	}
	if p.match(TOKEN_LEFT_PAREN) {
		expr := p.parseExpr()
		if p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression") {
			return expr
		}
		return nil
	}
	p.error(p.peek(), "Expect expression")
	return nil
}

func (p *Parser) match(tokenTypes ...TokenType) bool {
	for _, type_ := range tokenTypes {
		if p.check(type_) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(tokenType TokenType) bool {
	if p.isEnd() {
		return false
	}
	return p.peek().TokenType == tokenType
}

func (p *Parser) isEnd() bool {
	return p.peek().TokenType == TOKEN_EOF
}

func (p *Parser) peek() Token {
	return p.Tokens[p.Current]
}

func (p *Parser) advance() Token {
	if !p.isEnd() {
		p.Current++
	}
	return p.Tokens[p.Current-1]
}

func (p *Parser) previous() Token {
	return p.Tokens[p.Current-1]
}

func (p *Parser) consume(tokenType TokenType, message string) bool {
	if p.check(tokenType) {
		p.advance()
		return true
	}
	p.error(p.peek(), message)
	return false
}

func (p *Parser) error(token Token, message string) {
	p.parseErr = fmt.Errorf("%s, at line %d, got %v instead", message, token.Line, token)
	fmt.Printf("parse error: %s\n", p.parseErr.Error())
	p.synchronize()
}

// Synchronize the parser when it encounters a syntax error.
//
//	just skip to the next statement.
func (p *Parser) synchronize() {
	p.advance()
	for !p.isEnd() {
		if p.previous().TokenType == TOKEN_SEMICOLON {
			return
		}
		switch p.peek().TokenType {
		case TOKEN_CLASS, TOKEN_FUN, TOKEN_VAR, TOKEN_FOR, TOKEN_IF, TOKEN_WHILE, TOKEN_PRINT, TOKEN_RETURN:
			return
		}
		p.advance()
	}
}

func (p *Parser) GetError() error {
	return p.parseErr
}
