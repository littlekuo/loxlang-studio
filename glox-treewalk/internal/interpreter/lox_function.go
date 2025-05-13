package interpreter

import (
	"errors"

	"github.com/littlekuo/glox-treewalk/internal/syntax"
)

type LoxFunction struct {
	declaration   *syntax.Function
	closure       *Environment
	isInitializer bool
}

func NewLoxFunction(f *syntax.Function, closure *Environment, isInitializer bool) *LoxFunction {
	return &LoxFunction{
		declaration:   f,
		closure:       closure,
		isInitializer: isInitializer,
	}
}

func (l *LoxFunction) Call(i *Interpreter, args []any) syntax.Result {
	previousEnv := i.env
	i.env = NewEnvironment(l.closure)
	defer func() {
		i.env = previousEnv
	}()
	for idx, param := range l.declaration.Params {
		if err := i.define(param, args[idx]); err != nil {
			return syntax.Result{Err: err}
		}
	}
	result := syntax.Result{}
	for _, stmt := range l.declaration.Body {
		if err := i.execute(stmt); err != nil {
			var ret *ErrReturn
			if errors.As(err, &ret) {
				result.Value = ret.Value
				break
			}
			return syntax.Result{Err: err}
		}
	}
	if l.isInitializer {
		instance, err := l.closure.getAt(0, 0)
		if err != nil {
			return syntax.Result{Err: err}
		}
		return syntax.Result{Value: instance}
	}
	return result
}

func (l *LoxFunction) Arity() int {
	return len(l.declaration.Params)
}

func (l *LoxFunction) String() string {
	if l.declaration.Name.IsEmpty() {
		return "<anonymous fn>"
	}
	return "<fn " + l.declaration.Name.Lexeme + ">"
}

func (l *LoxFunction) Bind(instance *LoxInstance) *LoxFunction {
	environment := NewEnvironment(l.closure)
	environment.defineLocal(0, instance)
	return NewLoxFunction(l.declaration, environment, l.isInitializer)
}
