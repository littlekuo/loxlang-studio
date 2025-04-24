package syntax

import (
	"fmt"
	"strings"
)

type AstPrinter struct{}

func (a AstPrinter) Print(expr Expr) string {
	result := expr.Accept(a)
	return result.Value.(string)
}

func (a AstPrinter) VisitBinaryExpr(expr *Binary) Result {
	return Result{Value: a.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)}
}

func (a AstPrinter) VisitGroupingExpr(expr *Grouping) Result {
	return Result{Value: a.parenthesize("group", expr.Expression)}
}

func (a AstPrinter) VisitLiteralExpr(expr *Literal) Result {
	if expr.Value == nil {
		return Result{Value: "nil"}
	}
	return Result{Value: fmt.Sprintf("%v", expr.Value)}
}

func (a AstPrinter) VisitUnaryExpr(expr *Unary) Result {
	return Result{Value: a.parenthesize(expr.Operator.Lexeme, expr.Right)}
}

func (a AstPrinter) parenthesize(name string, exprs ...Expr) string {
	var builder strings.Builder

	builder.WriteString("(" + name)
	for _, expr := range exprs {
		builder.WriteString(" ")
		builder.WriteString(a.Print(expr))
	}
	builder.WriteString(")")

	return builder.String()
}
