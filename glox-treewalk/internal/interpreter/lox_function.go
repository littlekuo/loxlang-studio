package interpreter

import (
	"errors"

	"github.com/littlekuo/glox-treewalk/internal/syntax"
)

type LoxFunction struct {
	declaration *syntax.Function
	closure     *Environment
}

func NewLoxFunction(f *syntax.Function, closure *Environment) *LoxFunction {
	return &LoxFunction{
		declaration: f,
		closure:     closure,
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
	for _, stmt := range l.declaration.Body {
		if err := i.execute(stmt); err != nil {
			var ret *ErrReturn
			if errors.As(err, &ret) {
				return syntax.Result{Value: ret.Value}
			}
			return syntax.Result{Err: err}
		}
	}
	return syntax.Result{}
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
