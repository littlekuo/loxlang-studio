package interpreter

import (
	"time"

	"github.com/littlekuo/glox-treewalk/internal/syntax"
)

type Clock struct {
}

func NewClock() *Clock {
	return &Clock{}
}

func (c *Clock) Arity() int {
	return 0
}

func (c *Clock) Call(interpreter *Interpreter, args []any) syntax.Result {
	return syntax.Result{Value: float64(time.Now().UnixMilli())}
}

func (c *Clock) String() string {
	return "<native fn>"
}
