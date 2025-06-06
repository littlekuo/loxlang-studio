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



expression     ->  assignment
assignment     ->  (call "." )? IDENTIFIER "=" assignment
                   | logical_or

logical_or     ->  logical_and ( "or" logical_and )*
logical_and    ->  equality ( "and" equality )*
equality       ->  comparison ( ( "!=" | "==" ) comparison )*
comparison     ->  term ( ( ">" | ">=" | "<" | "<=" ) term )*
term           ->  factor ( ( "-" | "+" ) factor )*
factor         ->  unary ( ( "/" | "*" ) unary )*
unary          ->  ( "!" | "-" ) unary | call
call           → primary ( "(" arguments? ")" | "." IDENTIFIER )* ;
primary        ->  NUMBER | STRING | "false" | "true" | "nil" | "(" expression ")" | IDENTIFIER
                 | anonymous_func | super "." IDENTIFIER
anonymous_func ->  "fun" "(" parameters? ")" block
arguments      ->  expression ( "," expression )* ;

program        -> declaration* EOF

declaration    -> classDecl
                | funDecl
                | varDecl
                | statement


classDecl      -> "class" IDENTIFIER ( "<" IDENTIFIER )?
                  "{" function* "}" ;
funDecl        -> "fun" function
function       -> IDENTIFIER "(" parameters? ")" block
parameters     -> IDENTIFIER ( "," IDENTIFIER )*

varDecl        -> "var" IDENTIFIER ( "=" expression )? ";"


statement      -> exprStmt
                | printStmt
			    | block
			    | ifStmt
			    | whileStmt
			    | forStmt
			    | breakStmt
			    | continueStmt
                | returnStmt

exprStmt       ->  expression ";" ;
printStmt      -> "print" expression ";"
block          -> "{" declaration* "}"
ifStmt         -> "if" "(" expression ")" statement ( "else" statement )?
whileStmt      -> "while" "(" expression ")" statement
forStmt        -> "for" "(" ( varDecl | exprStmt | ";" )
				  expression? ";" expression? ")" statement
breakStmt      -> "break" ";"
continueStmt   -> "continue" ";"
returnStmt     -> "return" expression? ";"
*/

type Parser struct {
	Tokens    []Token
	Current   int
	parseErr  error
	loopDepth int
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
	if p.match(TOKEN_CLASS) {
		return p.parseClassDecl()
	}
	if p.match(TOKEN_FUN) {
		return p.parseFunction(false, "function")
	}
	if p.match(TOKEN_VAR) {
		return p.parseVarDecl()
	}
	return p.parseStmt()
}

func (p *Parser) parseClassDecl() (*Class, error) {
	if cErr := p.consume(TOKEN_IDENTIFIER, "expect class name"); cErr != nil {
		return nil, cErr
	}
	name := p.previous()
	var superClass *Variable = nil
	if p.match(TOKEN_LESS) {
		if cErr := p.consume(TOKEN_IDENTIFIER, "expect superclass name"); cErr != nil {
			return nil, cErr
		}
		superClass = NewVariable(p.previous())
	}
	if p.match(TOKEN_LEFT_BRACE) {
		methods := make([]*Function, 0)
		for !p.check(TOKEN_RIGHT_BRACE) && !p.isEnd() {
			method, err := p.parseFunction(false, "method")
			if err != nil {
				return nil, err
			}
			methods = append(methods, method)
		}
		if cErr := p.consume(TOKEN_RIGHT_BRACE, "expect '}' after class body"); cErr != nil {
			return nil, cErr
		}
		return NewClass(name, superClass, methods), nil
	}
	return nil, p.error(p.peek(), "expect '{' after class name")
}

func (p *Parser) parseFunction(anonymous bool, kind string) (*Function, error) {
	var name Token
	if !anonymous {
		if err := p.consume(TOKEN_IDENTIFIER, "expect "+kind+" name."); err != nil {
			return nil, err
		}
		name = p.previous()
	}
	if cErr := p.consume(TOKEN_LEFT_PAREN, "expect '(' after "+kind+" name"); cErr != nil {
		return nil, cErr
	}
	params := make([]Token, 0)
	if !p.check(TOKEN_RIGHT_PAREN) {
		for {
			if len(params) >= 255 {
				return nil, p.error(p.peek(), "can't have more than 255 parameters")
			}
			if pErr := p.consume(TOKEN_IDENTIFIER, "expect parameter name"); pErr != nil {
				return nil, pErr
			}
			params = append(params, p.previous())
			if !p.match(TOKEN_COMMA) {
				break
			}
		}
	}
	if cErr := p.consume(TOKEN_RIGHT_PAREN, "expect ')' after parameters"); cErr != nil {
		return nil, cErr
	}
	if cErr := p.consume(TOKEN_LEFT_BRACE, "expect '{' before "+kind+" body"); cErr != nil {
		return nil, cErr
	}
	body, err := p.parseBlocks()
	if err != nil {
		return nil, err
	}
	return NewFunction(name, params, body), nil
}

func (p *Parser) parseVarDecl() (Stmt, error) {
	if cErr := p.consume(TOKEN_IDENTIFIER, "expect variable name"); cErr != nil {
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

	if cErr := p.consume(TOKEN_SEMICOLON, "expect ';' after variable declaration"); cErr != nil {
		return nil, cErr
	}
	return NewVar(name, initializer), nil
}

func (p *Parser) parseStmt() (Stmt, error) {
	if p.match(TOKEN_IF) {
		return p.parseIfStmt()
	}
	if p.match(TOKEN_PRINT) {
		return p.parsePrintStmt()
	}
	if p.match(TOKEN_WHILE) {
		return p.parseWhileStmt()
	}
	if p.match(TOKEN_FOR) {
		return p.parseForStmt()
	}
	if p.match(TOKEN_BREAK) {
		return p.parseBreakStmt()
	}
	if p.match(TOKEN_CONTINUE) {
		return p.parseContinueStmt()
	}
	if p.match(TOKEN_RETURN) {
		return p.parseReturnStmt()
	}
	if p.match(TOKEN_LEFT_BRACE) {
		blocks, bErr := p.parseBlocks()
		if bErr != nil {
			return nil, bErr
		}
		return NewBlock(blocks), nil
	}
	return p.parseExprStmt()
}

func (p *Parser) parseReturnStmt() (Stmt, error) {
	keyword := p.previous()
	var value Expr
	if !p.check(TOKEN_SEMICOLON) {
		var err error
		value, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}
	if cErr := p.consume(TOKEN_SEMICOLON, "expect ';' after return value"); cErr != nil {
		return nil, cErr
	}
	return NewReturn(keyword, value), nil
}

func (p *Parser) parseBreakStmt() (Stmt, error) {
	if p.loopDepth == 0 {
		return nil, p.error(p.previous(), "break not inside loop")
	}
	if cErr := p.consume(TOKEN_SEMICOLON, "expect ';' after break"); cErr != nil {
		return nil, cErr
	}
	return NewBreak(p.previous()), nil
}

func (p *Parser) parseContinueStmt() (Stmt, error) {
	if p.loopDepth == 0 {
		return nil, p.error(p.previous(), "continue not inside loop")
	}
	if cErr := p.consume(TOKEN_SEMICOLON, "expect ';' after continue"); cErr != nil {
		return nil, cErr
	}
	return NewContinue(p.previous()), nil
}

// desugar for loop
func (p *Parser) parseForStmt() (Stmt, error) {
	p.loopDepth++
	defer func() { p.loopDepth-- }()
	var err error
	if err = p.consume(TOKEN_LEFT_PAREN, "expect '(' after 'for'"); err != nil {
		return nil, err
	}
	var initializer Stmt
	if p.match(TOKEN_SEMICOLON) {
	} else if p.match(TOKEN_VAR) {
		initializer, err = p.parseVarDecl()
		if err != nil {
			return nil, err
		}
	} else {
		initializer, err = p.parseExprStmt()
		if err != nil {
			return nil, err
		}
	}
	var condition Expr
	if !p.check(TOKEN_SEMICOLON) {
		condition, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}
	if cErr := p.consume(TOKEN_SEMICOLON, "expect ';' after loop condition"); cErr != nil {
		return nil, cErr
	}
	var increment Expr
	if !p.check(TOKEN_RIGHT_PAREN) {
		increment, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}
	if cErr := p.consume(TOKEN_RIGHT_PAREN, "expect ')' after for clauses"); cErr != nil {
		return nil, cErr
	}
	var body Stmt
	body, err = p.parseStmt()
	if err != nil {
		return nil, err
	}
	if condition == nil {
		// if condition is nil, use true
		condition = NewLiteral(true)
	}
	if increment == nil {
		body = NewWhile(condition, body)
	} else {
		body = NewForDesugaredWhile(condition, body, increment)
	}
	if initializer != nil {
		body = NewBlock([]Stmt{initializer, body})
	}
	return body, nil
}

func (p *Parser) parseWhileStmt() (Stmt, error) {
	p.loopDepth++
	defer func() { p.loopDepth-- }()
	if err := p.consume(TOKEN_LEFT_PAREN, "expect '(' after 'while'."); err != nil {
		return nil, err
	}
	condition, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if cErr := p.consume(TOKEN_RIGHT_PAREN, "expect ')' after while condition"); cErr != nil {
		return nil, cErr
	}
	body, err := p.parseStmt()
	if err != nil {
		return nil, err
	}
	return NewWhile(condition, body), nil
}

func (p *Parser) parseIfStmt() (Stmt, error) {
	if cErr := p.consume(TOKEN_LEFT_PAREN, "expect '(' after 'if'"); cErr != nil {
		return nil, cErr
	}
	condition, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if cErr := p.consume(TOKEN_RIGHT_PAREN, "expect ')' after if condition"); cErr != nil {
		return nil, cErr
	}
	thenBranch, err := p.parseStmt()
	if err != nil {
		return nil, err
	}
	var elseBranch Stmt
	if p.match(TOKEN_ELSE) {
		elseBranch, err = p.parseStmt()
		if err != nil {
			return nil, err
		}
	}
	return NewIf(condition, thenBranch, elseBranch), nil
}

func (p *Parser) parseBlocks() ([]Stmt, error) {
	stmts := make([]Stmt, 0)
	for !p.check(TOKEN_RIGHT_BRACE) && !p.isEnd() {
		stmt, err := p.parseDeclaration()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
	if cErr := p.consume(TOKEN_RIGHT_BRACE, "expect '}' after block"); cErr != nil {
		return nil, cErr
	}
	return stmts, nil
}

func (p *Parser) parsePrintStmt() (Stmt, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if cErr := p.consume(TOKEN_SEMICOLON, "expect ';' after value"); cErr != nil {
		return nil, cErr
	}
	return &Print{Expression: expr}, nil
}

func (p *Parser) parseExprStmt() (Stmt, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if cErr := p.consume(TOKEN_SEMICOLON, "expect ';' after value"); cErr != nil {
		return nil, cErr
	}
	return &Expression{Expression: expr}, nil
}

func (p *Parser) parseExpr() (Expr, error) {
	return p.parseAssignment()
}

func (p *Parser) parseAssignment() (Expr, error) {
	expr, err := p.parseLogicalOr()
	if err != nil {
		return nil, err
	}

	if p.match(TOKEN_EQUAL) {
		equalToken := p.previous()
		value, pErr := p.parseAssignment()
		if pErr != nil {
			return nil, pErr
		}

		if variable, ok := expr.(*Variable); ok {
			return NewAssign(variable.Name, value), nil
		} else if variable, ok := expr.(*Get); ok {
			return NewSet(variable.Object, variable.Name, value), nil
		}

		return nil, p.error(equalToken, "Invalid assignment target.")
	}
	return expr, nil
}

func (p *Parser) parseLogicalOr() (Expr, error) {
	expr, err := p.parseLogicalAnd()
	if err != nil {
		return nil, err
	}

	for p.match(TOKEN_OR) {
		op := p.previous()
		right, err := p.parseLogicalAnd()
		if err != nil {
			return nil, err
		}
		expr = NewLogical(expr, op, right)
	}

	return expr, nil
}

func (p *Parser) parseLogicalAnd() (Expr, error) {
	expr, err := p.parseEquality()
	if err != nil {
		return nil, err
	}

	for p.match(TOKEN_AND) {
		op := p.previous()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		expr = NewLogical(expr, op, right)
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

	return p.parseCall()
}

func (p *Parser) parseCall() (Expr, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(TOKEN_LEFT_PAREN) {
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(TOKEN_DOT) {
			err := p.consume(TOKEN_IDENTIFIER, "expect property name after '.'")
			if err != nil {
				return nil, err
			}
			expr = NewGet(expr, p.previous())
		} else {
			break
		}
	}

	return expr, nil
}

func (p *Parser) finishCall(callee Expr) (Expr, error) {
	args := make([]Expr, 0)
	if !p.check(TOKEN_RIGHT_PAREN) {
		for {
			if len(args) >= 255 {
				return nil, p.error(p.peek(), "can't have more than 255 arguments.")
			}
			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, expr)
			if !p.match(TOKEN_COMMA) {
				break
			}
		}
	}
	if cErr := p.consume(TOKEN_RIGHT_PAREN, "expect ')' after arguments"); cErr != nil {
		return nil, cErr
	}
	return NewCall(callee, p.previous(), args), nil
}

func (p *Parser) parseAnonymousFunction() (Expr, error) {
	decl, err := p.parseFunction(true, "function")
	if err != nil {
		return nil, err
	}
	return NewAnonymousFunction(decl), nil
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
	if p.match(TOKEN_THIS) {
		return NewThis(p.previous()), nil
	}
	if p.match(TOKEN_IDENTIFIER) {
		return NewVariable(p.previous()), nil
	}
	if p.match(TOKEN_FUN) {
		return p.parseAnonymousFunction()
	}
	if p.match(TOKEN_SUPER) {
		keyword := p.previous()
		if err := p.consume(TOKEN_DOT, "expect '.' after 'super'"); err != nil {
			return nil, err
		}
		if err := p.consume(TOKEN_IDENTIFIER, "expect superclass method name"); err != nil {
			return nil, err
		}
		return NewSuper(keyword, p.previous()), nil
	}
	if p.match(TOKEN_LEFT_PAREN) {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if cErr := p.consume(TOKEN_RIGHT_PAREN, "expect ')' after expression"); cErr != nil {
			return nil, cErr
		}
		return expr, nil
	}
	return nil, p.error(p.peek(), "expect expression")
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
