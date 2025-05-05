package interpreter

import (
	"fmt"

	"github.com/littlekuo/glox-treewalk/internal/syntax"
)

type FuncType int

const (
	FuncTypeNone = iota
	FuncTypeFunction
)

type VarInfo struct {
	defined bool
	used    bool
	idx     int
}

type Resolver struct {
	interpreter *Interpreter
	scopes      []map[string]*VarInfo
	indices     []int // track indices of local scopes
	resolveErr  error
	curFuncType FuncType
}

func NewResolver(interpreter *Interpreter) *Resolver {
	return &Resolver{
		interpreter: interpreter,
		scopes:      make([]map[string]*VarInfo, 0),
	}
}

func (r *Resolver) GetError() error {
	return r.resolveErr
}

func (r *Resolver) Resolve(stmts []syntax.Stmt) {
	if rErr := r.resolveStmts(stmts); rErr != nil {
		fmt.Printf("resolve error: %s\n", rErr.Error())
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
	newScope := make(map[string]*VarInfo)
	r.scopes = append(r.scopes, newScope)
	r.indices = append(r.indices, 0)
}

func (r *Resolver) endScope() error {
	curScope := r.scopes[len(r.scopes)-1]
	for name, info := range curScope {
		if info.defined && !info.used {
			return fmt.Errorf("variable [%s] is not used in local", name)
		}
	}
	r.scopes = r.scopes[:len(r.scopes)-1]
	r.indices = r.indices[:len(r.indices)-1]
	return nil
}

func (r *Resolver) VisitBlockStmt(stmt *syntax.Block) error {
	r.beginScope()
	err := r.resolveStmts(stmt.Statements)
	if err != nil {
		return err
	}
	err = r.endScope()
	if err != nil {
		return err
	}
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
	curIdx := r.indices[len(r.indices)-1]
	scope[name.Lexeme] = &VarInfo{
		idx: curIdx,
	}
	r.indices[len(r.indices)-1]++
	return nil
}

func (r *Resolver) define(name syntax.Token) {
	if len(r.scopes) == 0 {
		return
	}
	scope := r.scopes[len(r.scopes)-1]
	if info, ok := scope[name.Lexeme]; ok {
		info.defined = true
		r.interpreter.recordLocalDefs(name, info.idx)
		return
	}
	panic("not found by name")
}

func (r *Resolver) peek() map[string]*VarInfo {
	if len(r.scopes) == 0 {
		panic("No scopes")
	}
	return r.scopes[len(r.scopes)-1]
}

func (r *Resolver) VisitVariableExpr(expr *syntax.Variable) syntax.Result {
	if len(r.scopes) > 0 {
		if info, ok := r.peek()[expr.Name.Lexeme]; ok && !info.defined {
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
		if info, ok := r.scopes[i][name.Lexeme]; ok {
			info.used = true
			r.interpreter.resolve(expr, len(r.scopes)-1-i, info.idx)
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
	if !stmt.Name.IsEmpty() {
		// means it is not anonymous function
		err := r.declare(stmt.Name)
		if err != nil {
			return err
		}
		r.define(stmt.Name)
	}
	return r.resolveFunctionStmt(stmt, FuncTypeFunction)
}

func (r *Resolver) resolveFunctionStmt(f *syntax.Function, funcType FuncType) error {
	enclosingFunc := r.curFuncType
	r.curFuncType = funcType
	defer func() { r.curFuncType = enclosingFunc }()
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
	if err := r.endScope(); err != nil {
		return err
	}
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
	if r.curFuncType == FuncTypeNone {
		r.resolveErr = fmt.Errorf("can't return from top-level code")
		return r.resolveErr
	}
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
