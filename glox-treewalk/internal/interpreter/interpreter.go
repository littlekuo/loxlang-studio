package interpreter

import (
	"fmt"
	"reflect"

	"github.com/littlekuo/glox-treewalk/internal/syntax"
)

type Environment struct {
	valueMap map[string]interface{}
}

func (e *Environment) define(name string, val any) {
	if e.valueMap == nil {
		e.valueMap = make(map[string]interface{})
	}
	e.valueMap[name] = val
}

func (e *Environment) get(name syntax.Token) (interface{}, error) {
	val, ok := e.valueMap[name.Lexeme]
	if !ok {
		return nil, fmt.Errorf("undefined variable '%s'", name.Lexeme)
	}
	return val, nil
}

func (e *Environment) assign(name syntax.Token, value any) {
	if e.valueMap == nil {
		e.valueMap = make(map[string]interface{})
	}
	e.valueMap[name.Lexeme] = value
}

type Interpreter struct {
	interpretErr error
	env          *Environment
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		env: &Environment{},
	}
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

func (a *Interpreter) VisitVarStmt(stmt *syntax.Var) error {
	var value any
	if stmt.Initializer != nil {
		result := stmt.Initializer.Accept(a)
		if result.Err != nil {
			return result.Err
		}
		value = result.Value
	}
	a.env.define(stmt.Name.Lexeme, value)

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

func (a *Interpreter) VisitAssignExpr(expr *syntax.Assign) syntax.Result {
	result := expr.Value.Accept(a)
	if result.Err != nil {
		return syntax.Result{Err: result.Err}
	}
	a.env.assign(expr.Name, result.Value)
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

func (a *Interpreter) VisitVariableExpr(expr *syntax.Variable) syntax.Result {
	val, err := a.env.get(expr.Name)
	if err != nil {
		return syntax.Result{Err: err}
	}
	return syntax.Result{Value: val}
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
