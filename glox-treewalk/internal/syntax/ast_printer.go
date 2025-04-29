package syntax

import (
	"fmt"
	"strings"
)

type AstPrinter struct {
	desc  string
	ident int
}

func (a *AstPrinter) TopPrintStmts(stmts []Stmt) error {
	fmt.Println("------- result -------")
	for _, stmt := range stmts {
		a.ident = 0
		if err := a.printStmt(stmt); err != nil {
			return err
		}
	}
	fmt.Println(a.desc)
	return nil
}

func (a *AstPrinter) printStmt(stmt Stmt) error {
	if err := stmt.Accept(a); err != nil {
		return err
	}
	a.desc += "\n"
	return nil
}

func (a *AstPrinter) VisitBreakStmt(stmt *Break) error {
	a.desc += indentString(a.ident, "break")
	return nil
}

func (a *AstPrinter) VisitBlockStmt(stmt *Block) error {
	a.desc += indentString(a.ident, "(block \n")
	a.ident += 2
	for _, stmt := range stmt.Statements {
		if err := a.printStmt(stmt); err != nil {
			return err
		}
	}
	a.ident -= 2
	a.desc += indentString(a.ident, ")")
	return nil
}

func (a *AstPrinter) VisitWhileStmt(stmt *While) error {
	a.desc += indentString(a.ident, "(while ")
	a.desc += a.PrintExpr(stmt.Condition)
	a.desc += "\n"
	a.ident += 2
	if err := a.printStmt(stmt.Body); err != nil {
		return err
	}
	a.ident -= 2
	a.desc += indentString(a.ident, ")")
	return nil
}

func (a *AstPrinter) VisitIfStmt(stmt *If) error {
	a.desc += indentString(a.ident, "(if ")
	a.desc += a.PrintExpr(stmt.Condition)
	a.desc += "\n"
	a.ident += 2
	if err := a.printStmt(stmt.Thenbranch); err != nil {
		return err
	}
	a.ident -= 2
	a.desc += indentString(a.ident, "else")
	a.ident += 2
	if stmt.Elsebranch != nil {
		if err := a.printStmt(stmt.Elsebranch); err != nil {
			return err
		}
	}
	a.ident -= 2
	a.desc += indentString(a.ident, ")")
	return nil
}

func (a *AstPrinter) VisitExpressionStmt(stmt *Expression) error {
	a.desc += indentString(a.ident, a.PrintExpr(stmt.Expression))
	return nil
}

func (a *AstPrinter) VisitPrintStmt(stmt *Print) error {
	a.desc += indentString(a.ident, "(print ")
	a.desc += a.PrintExpr(stmt.Expression)
	a.desc += ")"
	return nil
}

func (a *AstPrinter) VisitVarStmt(stmt *Var) error {
	a.desc += indentString(a.ident, "(define "+stmt.Name.Lexeme+" ")
	if stmt.Initializer != nil {
		a.desc += a.PrintExpr(stmt.Initializer)
	}
	if stmt.Initializer == nil {
		a.desc += "nil"
	}
	a.desc += ")"
	return nil
}

func (a *AstPrinter) PrintExpr(expr Expr) string {
	return expr.Accept(a).Value.(string)
}

func (a *AstPrinter) VisitLogicalExpr(expr *Logical) Result {
	return Result{Value: a.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)}
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

func indentString(repeat int, content string) string {
	return strings.Repeat(" ", repeat) + content
}
