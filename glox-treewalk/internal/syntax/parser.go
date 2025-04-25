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



expression ->  assignment
assignment ->  IDENTIFIER "=" assignment
                | equality ;
equality   ->  comparison ( ( "!=" | "==" ) comparison )*
comparison ->  term ( ( ">" | ">=" | "<" | "<=" ) term )*
term       ->  factor ( ( "-" | "+" ) factor )*
factor     ->  unary ( ( "/" | "*" ) unary )*
unary      ->  ( "!" | "-" ) unary | primary
primary    ->  NUMBER | STRING | "false" | "true" | "nil" | "(" expr ")" | IDENTIFIER


program        → declaration* EOF ;

declaration    → varDecl
               | statement ;

varDecl        → "var" IDENTIFIER ( "=" expression )? ";" ;


statement      → exprStmt
               | printStmt ;

exprStmt       →  expression ";" ;
printStmt      → "print" expression ";" ;
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

func (p *Parser) Parse() []Stmt {
	stmts := make([]Stmt, 0)
	for !p.isEnd() {
		stmt, err := p.parseDeclaration()
		if err != nil {
			// record the last error
			fmt.Printf("parse Err:%s\n", err.Error())
			p.parseErr = err
			p.synchronize()
			continue
		}
		stmts = append(stmts, stmt)
	}
	return stmts
}

func (p *Parser) parseDeclaration() (Stmt, error) {
	if p.match(TOKEN_VAR) {
		return p.parseVarDecl()
	}
	return p.parseStmt()
}

func (p *Parser) parseVarDecl() (Stmt, error) {
	if cErr := p.consume(TOKEN_IDENTIFIER, "Expect variable name"); cErr != nil {
		return nil, cErr
	}
	name := p.previous()
	var initializer Expr
	if p.match(TOKEN_EQUAL) {
		var pErr error
		initializer, pErr = p.parseExpr()
		if pErr != nil {
			return nil, pErr
		}
	}

	if cErr := p.consume(TOKEN_SEMICOLON, "Expect ';' after variable declaration"); cErr != nil {
		return nil, cErr
	}
	return NewVar(name, initializer), nil
}

func (p *Parser) parseStmt() (Stmt, error) {
	if p.match(TOKEN_PRINT) {
		return p.parsePrintStmt()
	}
	return p.parseExprStmt()
}

func (p *Parser) parsePrintStmt() (Stmt, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if cErr := p.consume(TOKEN_SEMICOLON, "Expect ';' after value"); cErr != nil {
		return nil, cErr
	}
	return &Print{Expression: expr}, nil
}

func (p *Parser) parseExprStmt() (Stmt, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if cErr := p.consume(TOKEN_SEMICOLON, "Expect ';' after value"); cErr != nil {
		return nil, cErr
	}
	return &Expression{Expression: expr}, nil
}

func (p *Parser) parseExpr() (Expr, error) {
	return p.parseAssignment()
}

func (p *Parser) parseAssignment() (Expr, error) {
	expr, pErr := p.parseEquality()
	if pErr != nil {
		return nil, pErr
	}

	if p.match(TOKEN_EQUAL) {
		equalToken := p.previous()
		value, pErr := p.parseAssignment()
		if pErr != nil {
			return nil, pErr
		}

		if variable, ok := expr.(*Variable); ok {
			return NewAssign(variable.Name, value), nil
		}

		return nil, p.error(equalToken, "Invalid assignment target.")
	}
	return expr, nil
}

func (p *Parser) parseEquality() (Expr, error) {
	expr, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.match(TOKEN_BANG_EQUAL, TOKEN_EQUAL_EQUAL) {
		op := p.previous()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, op, right)
	}

	return expr, nil
}

func (p *Parser) parseComparison() (Expr, error) {
	expr, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	for p.match(TOKEN_GREATER, TOKEN_GREATER_EQUAL, TOKEN_LESS, TOKEN_LESS_EQUAL) {
		op := p.previous()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, op, right)
	}

	return expr, nil
}

func (p *Parser) parseTerm() (Expr, error) {
	expr, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	for p.match(TOKEN_MINUS, TOKEN_PLUS) {
		op := p.previous()
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, op, right)
	}

	return expr, nil
}

func (p *Parser) parseFactor() (Expr, error) {
	expr, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	if p.match(TOKEN_SLASH, TOKEN_STAR) {
		op := p.previous()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, op, right)
	}

	return expr, nil
}

func (p *Parser) parseUnary() (Expr, error) {
	if p.match(TOKEN_BANG, TOKEN_MINUS) {
		op := p.previous()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return NewUnary(right, op), nil
	}

	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (Expr, error) {
	if p.match(TOKEN_NUMBER) {
		return &Literal{Value: p.previous().Literal}, nil
	}
	if p.match(TOKEN_STRING) {
		return &Literal{Value: p.previous().Literal}, nil
	}
	if p.match(TOKEN_TRUE) {
		return &Literal{Value: true}, nil
	}
	if p.match(TOKEN_FALSE) {
		return &Literal{Value: false}, nil
	}
	if p.match(TOKEN_NIL) {
		return &Literal{Value: nil}, nil
	}
	if p.match(TOKEN_IDENTIFIER) {
		return NewVariable(p.previous()), nil
	}
	if p.match(TOKEN_LEFT_PAREN) {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if cErr := p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression"); cErr != nil {
			return nil, cErr
		}
		return expr, nil
	}
	return nil, p.error(p.peek(), "Expect expression")
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

func (p *Parser) consume(tokenType TokenType, message string) error {
	if p.check(tokenType) {
		p.advance()
		return nil
	}
	return p.error(p.peek(), message)
}

func (p *Parser) error(token Token, message string) error {
	return fmt.Errorf("%s, at line %d, got %v instead", message, token.Line, token)
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
