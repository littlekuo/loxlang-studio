package interpreter

import (
	"fmt"

	"github.com/littlekuo/glox-treewalk/internal/syntax"
)

type LoxInstance struct {
	loxClass *LoxClass
	fields   map[string]interface{}
}

func NewLoxInstance(loxClass *LoxClass) *LoxInstance {
	return &LoxInstance{
		loxClass: loxClass,
		fields:   make(map[string]interface{}),
	}
}

func (i *LoxInstance) String() string {
	return "<instance of " + i.loxClass.name + ">"
}

func (i *LoxInstance) Get(name syntax.Token) (interface{}, error) {
	if value, ok := i.fields[name.Lexeme]; ok {
		return value, nil
	}

	return nil, fmt.Errorf("undefined property %s", name.Lexeme)
}

func (i *LoxInstance) Set(name syntax.Token, value interface{}) error {
	i.fields[name.Lexeme] = value
	return nil
}
