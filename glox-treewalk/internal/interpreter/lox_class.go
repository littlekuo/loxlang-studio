package interpreter

import "github.com/littlekuo/glox-treewalk/internal/syntax"

type LoxClass struct {
	name string
}

func NewLoxClass(name string) *LoxClass {
	return &LoxClass{name: name}
}

func (c *LoxClass) String() string {
	return "<class " + c.name + ">"
}

func (c *LoxClass) Arity() int {
	return 0
}

func (c *LoxClass) Call(interpreter *Interpreter, args []interface{}) syntax.Result {
	loxInstance := NewLoxInstance(c)
	return syntax.Result{Value: loxInstance}
}
