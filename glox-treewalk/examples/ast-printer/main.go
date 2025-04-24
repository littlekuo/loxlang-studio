package main

import (
	"fmt"

	"github.com/littlekuo/go-lox/internal/syntax"
)

func main() {
	// example:
	expression := &syntax.Binary{
		Left: &syntax.Unary{
			Operator: syntax.Token{TokenType: syntax.TOKEN_MINUS, Lexeme: "-", Line: 1},
			Right:    &syntax.Literal{Value: 123},
		},
		Operator: syntax.Token{TokenType: syntax.TOKEN_STAR, Lexeme: "*", Line: 1},
		Right: &syntax.Grouping{
			Expression: &syntax.Literal{Value: 45.67},
		},
	}

	printer := &syntax.AstPrinter{}
	fmt.Println(printer.Print(expression))
}
