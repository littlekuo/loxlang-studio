package interpreter

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/littlekuo/glox-treewalk/internal/syntax"
)

var (
	errBreak    = errors.New("break")
	errContinue = errors.New("continue")
)

type ErrReturn struct {
	Value interface{}
}

func (e *ErrReturn) Error() string {
	return fmt.Sprintf("return %v", e.Value)
}

type Callable interface {
	Call(i *Interpreter, args []interface{}) syntax.Result
	Arity() int
}

type Loc struct {
	depth int
	idx   int
}

type Interpreter struct {
	interpretErr error
	env          *Environment
	localAccess  map[syntax.Expr]*Loc // track local variable access
	localDefs    map[syntax.Token]int // track local variable definition
	globals      *Environment
}

func NewInterpreter() *Interpreter {
	globals := NewEnvironment(nil)
	_ = globals.defineGlobal("clock", NewClock())
	return &Interpreter{
		localAccess: make(map[syntax.Expr]*Loc),
		localDefs:   make(map[syntax.Token]int),
		env:         globals,
		globals:     globals,
	}
}

func (a *Interpreter) define(name syntax.Token, value any) error {
	if idx, ok := a.localDefs[name]; ok {
		return a.env.defineLocal(idx, value)
	} else {
		return a.globals.defineGlobal(name.Lexeme, value)
	}
}

func (a *Interpreter) assign(expr *syntax.Assign, value interface{}) error {
	loc, ok := a.localAccess[expr]
	if ok {
		return a.env.assignAt(loc.depth, loc.idx, value)
	}
	return a.globals.assignGlobal(expr.Name, value)
}

func (a *Interpreter) GetError() error {
	return a.interpretErr
}

func (a *Interpreter) Interpret(stmts []syntax.Stmt) {
	for _, stmt := range stmts {
		if err := a.execute(stmt); err != nil {
			fmt.Printf("interpret error: %s\n", err.Error())
			a.interpretErr = err
			return
		}
	}
}

func (a *Interpreter) execute(stmt syntax.Stmt) error {
	return stmt.Accept(a)
}

func (a *Interpreter) resolve(expr syntax.Expr, depth int, idx int) {
	a.localAccess[expr] = &Loc{depth: depth, idx: idx}
}

func (a *Interpreter) recordLocalDefs(name syntax.Token, idx int) {
	a.localDefs[name] = idx
}

func (a *Interpreter) VisitReturnStmt(stmt *syntax.Return) error {
	if stmt.Value != nil {
		result := stmt.Value.Accept(a)
		if result.Err != nil {
			return result.Err
		}
		return &ErrReturn{Value: result.Value}
	}
	return nil
}

func (a *Interpreter) VisitFunctionStmt(stmt *syntax.Function) error {
	fn := NewLoxFunction(stmt, a.env)
	return a.define(stmt.Name, fn)
}

func (a *Interpreter) VisitBreakStmt(stmt *syntax.Break) error {
	return errBreak
}

func (a *Interpreter) VisitContinueStmt(stmt *syntax.Continue) error {
	return errContinue
}

func (a *Interpreter) VisitForDesugaredWhileStmt(stmt *syntax.ForDesugaredWhile) error {
	for {
		condResult := stmt.Condition.Accept(a)
		if condResult.Err != nil {
			return condResult.Err
		}
		if !isTruthy(condResult.Value) {
			break
		}
		if err := a.execute(stmt.Body); err != nil {
			if errors.Is(err, errBreak) {
				return nil
			} else if errors.Is(err, errContinue) {
			} else {
				return err
			}
		}
		if stmt.Increment != nil {
			result := a.executeExpr(stmt.Increment)
			if result.Err != nil {
				return result.Err
			}
		}
	}
	return nil
}

func (a *Interpreter) VisitWhileStmt(stmt *syntax.While) error {
	for {
		condResult := stmt.Condition.Accept(a)
		if condResult.Err != nil {
			return condResult.Err
		}
		if !isTruthy(condResult.Value) {
			break
		}
		if err := a.execute(stmt.Body); err != nil {
			if errors.Is(err, errBreak) {
				return nil
			} else if errors.Is(err, errContinue) {
				continue
			}
			return err
		}
	}
	return nil
}

func (a *Interpreter) VisitIfStmt(stmt *syntax.If) error {
	condResult := stmt.Condition.Accept(a)
	if condResult.Err != nil {
		return condResult.Err
	}
	if isTruthy(condResult.Value) {
		return a.execute(stmt.Thenbranch)
	} else if stmt.Elsebranch != nil {
		return a.execute(stmt.Elsebranch)
	}
	return nil
}

func (a *Interpreter) VisitVarStmt(stmt *syntax.Var) error {
	var value any
	if stmt.Initializer != nil {
		result := stmt.Initializer.Accept(a)
		if result.Err != nil {
			return result.Err
		}
		value = result.Value
	}
	return a.define(stmt.Name, value)
}

func (a *Interpreter) VisitBlockStmt(stmt *syntax.Block) error {
	previousEnv := a.env
	a.env = NewEnvironment(a.env)
	defer func() { a.env = previousEnv }()
	for _, stmt := range stmt.Statements {
		if err := a.execute(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (a *Interpreter) VisitExpressionStmt(stmt *syntax.Expression) error {
	result := stmt.Expression.Accept(a)
	if result.Err != nil {
		return result.Err
	}
	return nil
}

func (a *Interpreter) VisitPrintStmt(stmt *syntax.Print) error {
	result := stmt.Expression.Accept(a)
	if result.Err != nil {
		return result.Err
	}
	fmt.Printf("%v\n", result.Value)
	return nil
}

func (a *Interpreter) VisitLogicalExpr(expr *syntax.Logical) syntax.Result {
	left := expr.Left.Accept(a)
	if left.Err != nil {
		return syntax.Result{Err: left.Err}
	}
	if expr.Operator.TokenType == syntax.TOKEN_OR {
		if isTruthy(left.Value) {
			return left
		}
	} else {
		if !isTruthy(left.Value) {
			return left
		}
	}
	return expr.Right.Accept(a)
}

func (a *Interpreter) executeExpr(expr syntax.Expr) syntax.Result {
	return expr.Accept(a)
}

func (a *Interpreter) VisitAssignExpr(expr *syntax.Assign) syntax.Result {
	result := expr.Value.Accept(a)
	if result.Err != nil {
		return syntax.Result{Err: result.Err}
	}
	err := a.assign(expr, result.Value)
	if err != nil {
		return syntax.Result{Err: err}
	}
	return result
}

func (a *Interpreter) VisitLiteralExpr(expr *syntax.Literal) syntax.Result {
	return syntax.Result{Value: expr.Value}
}

func (a *Interpreter) VisitGroupingExpr(expr *syntax.Grouping) syntax.Result {
	return expr.Expression.Accept(a)
}

func (a *Interpreter) VisitUnaryExpr(expr *syntax.Unary) syntax.Result {
	right := expr.Right.Accept(a)
	if right.Err != nil {
		return syntax.Result{Err: right.Err}
	}
	switch expr.Operator.TokenType {
	case syntax.TOKEN_MINUS:
		if cErr := checkNumberOperand(expr.Operator, right.Value); cErr != nil {
			return syntax.Result{Err: cErr}
		}
		return syntax.Result{Value: -right.Value.(float64)}
	case syntax.TOKEN_BANG:
		return syntax.Result{Value: !isTruthy(right.Value)}
	}
	// unreachable
	return syntax.Result{Err: fmt.Errorf("unknown unary operator: %s", expr.Operator.Lexeme)}
}

func (a *Interpreter) VisitBinaryExpr(expr *syntax.Binary) syntax.Result {
	left := expr.Left.Accept(a)
	if left.Err != nil {
		return syntax.Result{Err: left.Err}
	}
	right := expr.Right.Accept(a)
	if right.Err != nil {
		return syntax.Result{Err: right.Err}
	}

	switch expr.Operator.TokenType {
	case syntax.TOKEN_MINUS:
		if cErr := checkNumberOperands(expr.Operator, left.Value, right.Value); cErr != nil {
			return syntax.Result{Err: cErr}
		}
		return syntax.Result{Value: left.Value.(float64) - right.Value.(float64)}
	case syntax.TOKEN_PLUS:
		if leftVal, ok := left.Value.(float64); ok {
			if rightVal, ok_ := right.Value.(float64); ok_ {
				return syntax.Result{Value: leftVal + rightVal}
			}
			return syntax.Result{Err: fmt.Errorf("right value is not a number: %v", left.Value)}
		}
		if leftVal, ok := left.Value.(string); ok {
			if rightVal, ok_ := right.Value.(string); ok_ {
				return syntax.Result{Value: leftVal + rightVal}
			}
			return syntax.Result{Err: fmt.Errorf("right value is not a string: %v", left.Value)}
		}
	case syntax.TOKEN_SLASH:
		if cErr := checkNumberOperands(expr.Operator, left.Value, right.Value); cErr != nil {
			return syntax.Result{Err: cErr}
		}
		if right.Value.(float64) == 0 {
			return syntax.Result{Err: fmt.Errorf("division by zero")}
		}
		return syntax.Result{Value: left.Value.(float64) / right.Value.(float64)}
	case syntax.TOKEN_STAR:
		if cErr := checkNumberOperands(expr.Operator, left.Value, right.Value); cErr != nil {
			return syntax.Result{Err: cErr}
		}
		return syntax.Result{Value: left.Value.(float64) * right.Value.(float64)}
	case syntax.TOKEN_GREATER:
		if cErr := checkNumberOperands(expr.Operator, left.Value, right.Value); cErr != nil {
			return syntax.Result{Err: cErr}
		}
		return syntax.Result{Value: left.Value.(float64) > right.Value.(float64)}
	case syntax.TOKEN_GREATER_EQUAL:
		if cErr := checkNumberOperands(expr.Operator, left.Value, right.Value); cErr != nil {
			return syntax.Result{Err: cErr}
		}
		return syntax.Result{Value: left.Value.(float64) >= right.Value.(float64)}
	case syntax.TOKEN_LESS:
		if cErr := checkNumberOperands(expr.Operator, left.Value, right.Value); cErr != nil {
			return syntax.Result{Err: cErr}
		}
		return syntax.Result{Value: left.Value.(float64) < right.Value.(float64)}
	case syntax.TOKEN_LESS_EQUAL:
		if cErr := checkNumberOperands(expr.Operator, left.Value, right.Value); cErr != nil {
			return syntax.Result{Err: cErr}
		}
		return syntax.Result{Value: left.Value.(float64) <= right.Value.(float64)}
	case syntax.TOKEN_BANG_EQUAL:
		return syntax.Result{Value: !isEqual(left.Value, right.Value)}
	case syntax.TOKEN_EQUAL_EQUAL:
		return syntax.Result{Value: isEqual(left.Value, right.Value)}
	}
	return syntax.Result{Err: fmt.Errorf("unknown unary operator: %s", expr.Operator.Lexeme)}
}

func (a *Interpreter) VisitCallExpr(expr *syntax.Call) syntax.Result {
	callee := expr.Callee.Accept(a)
	if callee.Err != nil {
		return syntax.Result{Err: callee.Err}
	}
	args := make([]any, len(expr.Arguments))
	for i, arg := range expr.Arguments {
		argVal := arg.Accept(a)
		if argVal.Err != nil {
			return syntax.Result{Err: argVal.Err}
		}
		args[i] = argVal.Value
	}

	if calleeVal, ok := callee.Value.(Callable); ok {
		if calleeVal.Arity() != len(args) {
			return syntax.Result{Err: fmt.Errorf("wrong number of arguments: want=%d, got=%d", calleeVal.Arity(), len(args))}
		}
		return calleeVal.Call(a, args)
	}
	return syntax.Result{Err: fmt.Errorf("can only call functions and classes")}
}

func (a *Interpreter) VisitAnonymousFunctionExpr(expr *syntax.AnonymousFunction) syntax.Result {
	loxFunc := NewLoxFunction(expr.Decl, a.env)
	return syntax.Result{Value: loxFunc}
}

func (a *Interpreter) VisitVariableExpr(expr *syntax.Variable) syntax.Result {
	return a.lookupVariable(expr.Name, expr)
}

func (a *Interpreter) lookupVariable(name syntax.Token, expr syntax.Expr) syntax.Result {
	loc, ok := a.localAccess[expr]
	var obj any
	var err error
	if ok {
		obj, err = a.env.getAt(loc.depth, loc.idx)
	} else {
		obj, err = a.globals.getGlobal(name)
	}
	if err != nil {
		return syntax.Result{Err: err}
	}
	return syntax.Result{Value: obj}
}

func isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}
	switch value.(type) {
	case bool:
		return value.(bool)
	default:
		return true
	}
}

func isEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return true
	}

	switch aVal := a.(type) {
	case float64:
		if bVal, ok := b.(float64); ok {
			return aVal == bVal
		}
		return false
	case string:
		if bVal, ok := b.(string); ok {
			return aVal == bVal
		}
		return false
	default:
		return reflect.DeepEqual(a, b)
	}
}

func checkNumberOperand(operator syntax.Token, operand interface{}) error {
	if _, ok := operand.(float64); !ok {
		return fmt.Errorf("operator %s: operand must be a number", syntax.TokenTypeStr[operator.TokenType])
	}
	return nil
}

func checkNumberOperands(operator syntax.Token, left, right interface{}) error {
	if _, ok := left.(float64); !ok {
		return fmt.Errorf("operator %s: left operand must be a number", syntax.TokenTypeStr[operator.TokenType])
	}
	if _, ok := right.(float64); !ok {
		return fmt.Errorf("operator %s: right operand must be a number", syntax.TokenTypeStr[operator.TokenType])
	}
	return nil
}
