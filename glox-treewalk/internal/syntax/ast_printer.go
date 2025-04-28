package syntax

import (
	"fmt"
	"strings"
)

type AstPrinter struct {
	desc string
}

func (a *AstPrinter) PrintStmts(stmts []Stmt) error {
	for _, stmt := range stmts {
		if err := stmt.Accept(a); err != nil {
			return err
		}
		a.desc += "\n"
	}
	fmt.Println(a.desc)
	return nil
}

func (a *AstPrinter) VisitBlockStmt(stmt *Block) error {
	a.desc += "(block"
	for _, stmt := range stmt.Statements {
		if err := stmt.Accept(a); err != nil {
			return err
		}
	}
	a.desc += ")"
	return nil
}

func (a *AstPrinter) VisitIfStmt(stmt *If) error {
	a.desc += "(if "
	a.desc += a.PrintExpr(stmt.Condition)
	a.desc += " "
	if err := stmt.Thenbranch.Accept(a); err != nil {
		return err
	}
	a.desc += " "
	if stmt.Elsebranch != nil {
		if err := stmt.Elsebranch.Accept(a); err != nil {
			return err
		}
	}
	a.desc += ")"
	return nil
}

func (a *AstPrinter) VisitExpressionStmt(stmt *Expression) error {
	a.desc += a.PrintExpr(stmt.Expression)
	return nil
}

func (a *AstPrinter) VisitPrintStmt(stmt *Print) error {
	a.desc += "(print "
	a.desc += a.PrintExpr(stmt.Expression)
	a.desc += ")"
	return nil
}

func (a *AstPrinter) VisitVarStmt(stmt *Var) error {
	a.desc += "(define " + stmt.Name.Lexeme + " "
	if stmt.Initializer != nil {
		a.desc += a.PrintExpr(stmt.Initializer)
	}
	a.desc += ")"
	return nil
}

func (a *AstPrinter) PrintExpr(expr Expr) string {
	return expr.Accept(a).Value.(string)
}

func (a *AstPrinter) VisitAssignExpr(expr *Assign) Result {
	return Result{Value: a.parenthesize("set", NewVariable(expr.Name), expr.Value)}
}

func (a *AstPrinter) VisitBinaryExpr(expr *Binary) Result {
	return Result{Value: a.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)}
}

func (a *AstPrinter) VisitGroupingExpr(expr *Grouping) Result {
	return Result{Value: a.parenthesize("group", expr.Expression)}
}

func (a *AstPrinter) VisitLiteralExpr(expr *Literal) Result {
	if expr.Value == nil {
		return Result{Value: "nil"}
	}
	return Result{Value: fmt.Sprintf("%v", expr.Value)}
}

func (a *AstPrinter) VisitUnaryExpr(expr *Unary) Result {
	return Result{Value: a.parenthesize(expr.Operator.Lexeme, expr.Right)}
}

func (a *AstPrinter) VisitVariableExpr(expr *Variable) Result {
	return Result{Value: expr.Name.Lexeme}
}

func (a *AstPrinter) parenthesize(operatorName string, exprs ...Expr) string {
	var builder strings.Builder

	builder.WriteString("(" + operatorName)
	for _, expr := range exprs {
		builder.WriteString(" ")
		builder.WriteString(a.PrintExpr(expr))
	}
	builder.WriteString(")")

	return builder.String()
}
