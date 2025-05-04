package interpreter

import (
	"fmt"
	"github.com/littlekuo/glox-treewalk/internal/syntax"
)

type Resolver struct {
	interpreter *Interpreter
	scopes      []map[string]bool
	resolveErr  error
}

func NewResolver(interpreter *Interpreter) *Resolver {
	return &Resolver{
		interpreter: interpreter,
		scopes:      make([]map[string]bool, 0),
	}
}

func (r *Resolver) GetError() error {
	return r.resolveErr
}

func (r *Resolver) Resolve(stmts []syntax.Stmt) {
	if rErr := r.resolveStmts(stmts); rErr != nil {
		r.resolveErr = rErr
	}
}

func (r *Resolver) resolveStmts(statements []syntax.Stmt) error {
	for _, stmt := range statements {
		err := r.resolveStmt(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Resolver) resolveStmt(statement syntax.Stmt) error {
	return statement.Accept(r)
}

func (r *Resolver) resolveExpr(expr syntax.Expr) syntax.Result {
	return expr.Accept(r)
}

func (r *Resolver) beginScope() {
	newScope := make(map[string]bool)
	r.scopes = append(r.scopes, newScope)
}

func (r *Resolver) endScope() {
	r.scopes = r.scopes[:len(r.scopes)-1]
}

func (r *Resolver) VisitBlockStmt(stmt *syntax.Block) error {
	r.beginScope()
	err := r.resolveStmts(stmt.Statements)
	if err != nil {
		return err
	}
	r.endScope()
	return nil
}

func (r *Resolver) VisitVarStmt(stmt *syntax.Var) error {
	err := r.declare(stmt.Name)
	if err != nil {
		return err
	}
	if stmt.Initializer != nil {
		result := r.resolveExpr(stmt.Initializer)
		if result.Err != nil {
			return result.Err
		}
	}
	r.define(stmt.Name)
	return nil
}

func (r *Resolver) declare(name syntax.Token) error {
	if len(r.scopes) == 0 {
		return nil
	}
	scope := r.scopes[len(r.scopes)-1]
	if _, ok := scope[name.Lexeme]; ok {
		return fmt.Errorf("re-declare variable [%s]", name.Lexeme)
	}
	scope[name.Lexeme] = false
	return nil
}

func (r *Resolver) define(name syntax.Token) {
	if len(r.scopes) == 0 {
		return
	}
	scope := r.scopes[len(r.scopes)-1]
	scope[name.Lexeme] = true
}

func (r *Resolver) peek() map[string]bool {
	if len(r.scopes) == 0 {
		panic("No scopes")
	}
	return r.scopes[len(r.scopes)-1]
}

func (r *Resolver) VisitVariableExpr(expr *syntax.Variable) syntax.Result {
	if len(r.scopes) > 0 {
		if initialized, ok := r.peek()[expr.Name.Lexeme]; ok && !initialized {
			return syntax.Result{
				Err: fmt.Errorf("can't read local variable [%s] in its own initializer", expr.Name.Lexeme),
			}
		}
	}

	r.resolveLocal(expr, expr.Name)
	return syntax.Result{}
}

func (r *Resolver) resolveLocal(expr syntax.Expr, name syntax.Token) {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if _, ok := r.scopes[i][name.Lexeme]; ok {
			r.interpreter.resolve(expr, len(r.scopes)-1-i)
			return
		}
	}
}

func (r *Resolver) VisitAssignExpr(expr *syntax.Assign) syntax.Result {
	result := r.resolveExpr(expr.Value)
	if result.Err != nil {
		return result
	}
	r.resolveLocal(expr, expr.Name)
	return syntax.Result{}
}

func (r *Resolver) VisitFunctionStmt(stmt *syntax.Function) error {
	err := r.declare(stmt.Name)
	if err != nil {
		return err
	}
	r.define(stmt.Name)
	return r.resolveFunctionStmt(stmt)
}

func (r *Resolver) resolveFunctionStmt(f *syntax.Function) error {
	r.beginScope()
	for _, param := range f.Params {
		err := r.declare(param)
		if err != nil {
			return err
		}
		r.define(param)
	}
	if err := r.resolveStmts(f.Body); err != nil {
		return err
	}
	r.endScope()
	return nil
}

func (r *Resolver) VisitExpressionStmt(stmt *syntax.Expression) error {
	result := r.resolveExpr(stmt.Expression)
	if result.Err != nil {
		return result.Err
	}
	return nil
}

func (r *Resolver) VisitIfStmt(stmt *syntax.If) error {
	result := r.resolveExpr(stmt.Condition)
	if result.Err != nil {
		return result.Err
	}
	if err := r.resolveStmt(stmt.Thenbranch); err != nil {
		return err
	}
	if stmt.Elsebranch != nil {
		if err := r.resolveStmt(stmt.Elsebranch); err != nil {
			return err
		}
	}
	return nil
}

func (r *Resolver) VisitPrintStmt(stmt *syntax.Print) error {
	result := r.resolveExpr(stmt.Expression)
	if result.Err != nil {
		return result.Err
	}
	return nil
}

func (r *Resolver) VisitReturnStmt(stmt *syntax.Return) error {
	if stmt.Value != nil {
		result := r.resolveExpr(stmt.Value)
		if result.Err != nil {
			return result.Err
		}
	}
	return nil
}

func (r *Resolver) VisitWhileStmt(stmt *syntax.While) error {
	result := r.resolveExpr(stmt.Condition)
	if result.Err != nil {
		return result.Err
	}
	if err := r.resolveStmt(stmt.Body); err != nil {
		return err
	}
	return nil
}

func (r *Resolver) VisitBreakStmt(stmt *syntax.Break) error {
	return nil
}

func (r *Resolver) VisitContinueStmt(stmt *syntax.Continue) error {
	return nil
}

func (r *Resolver) VisitForDesugaredWhileStmt(stmt *syntax.ForDesugaredWhile) error {
	result := r.resolveExpr(stmt.Condition)
	if result.Err != nil {
		return result.Err
	}
	if err := r.resolveStmt(stmt.Body); err != nil {
		return err
	}
	if stmt.Increment != nil {
		result = r.resolveExpr(stmt.Increment)
		if result.Err != nil {
			return result.Err
		}
	}
	return nil
}

func (r *Resolver) VisitBinaryExpr(expr *syntax.Binary) syntax.Result {
	result := r.resolveExpr(expr.Left)
	if result.Err != nil {
		return result
	}
	result = r.resolveExpr(expr.Right)
	if result.Err != nil {
		return result
	}
	return syntax.Result{}
}

func (r *Resolver) VisitCallExpr(expr *syntax.Call) syntax.Result {
	result := r.resolveExpr(expr.Callee)
	if result.Err != nil {
		return result
	}
	for _, argument := range expr.Arguments {
		result = r.resolveExpr(argument)
		if result.Err != nil {
			return result
		}
	}
	return syntax.Result{}
}

func (r *Resolver) VisitGroupingExpr(expr *syntax.Grouping) syntax.Result {
	return r.resolveExpr(expr.Expression)
}

func (r *Resolver) VisitLiteralExpr(expr *syntax.Literal) syntax.Result {
	return syntax.Result{}
}

func (r *Resolver) VisitLogicalExpr(expr *syntax.Logical) syntax.Result {
	result := r.resolveExpr(expr.Left)
	if result.Err != nil {
		return result
	}
	result = r.resolveExpr(expr.Right)
	if result.Err != nil {
		return result
	}
	return syntax.Result{}
}

func (r *Resolver) VisitUnaryExpr(expr *syntax.Unary) syntax.Result {
	result := r.resolveExpr(expr.Right)
	if result.Err != nil {
		return result
	}
	return syntax.Result{}
}

func (r *Resolver) VisitAnonymousFunctionExpr(expr *syntax.AnonymousFunction) syntax.Result {
	err := r.resolveStmt(expr.Decl)
	if err != nil {
		return syntax.Result{Err: err}
	}
	return syntax.Result{}
}
