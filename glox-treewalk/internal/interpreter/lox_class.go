package interpreter

import "github.com/littlekuo/glox-treewalk/internal/syntax"

type LoxClass struct {
	name    string
	methods map[string]*LoxFunction
}

func NewLoxClass(name string, methods map[string]*LoxFunction) *LoxClass {
	return &LoxClass{name: name, methods: methods}
}

func (c *LoxClass) String() string {
	return "<class " + c.name + ">"
}

func (c *LoxClass) Arity() int {
	if initializer := c.methods["init"]; initializer != nil {
		return initializer.Arity()
	}
	return 0
}

func (c *LoxClass) Call(interpreter *Interpreter, args []interface{}) syntax.Result {
	loxInstance := NewLoxInstance(c)
	initializer := c.FindMethod("init")
	if initializer != nil {
		initializer.Bind(loxInstance).Call(interpreter, args)
	}
	return syntax.Result{Value: loxInstance}
}

func (c *LoxClass) FindMethod(methodName string) *LoxFunction {
	if method, ok := c.methods[methodName]; ok {
		return method
	}
	return nil
}
