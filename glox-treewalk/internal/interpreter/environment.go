package interpreter

import (
	"fmt"
	"github.com/littlekuo/glox-treewalk/internal/syntax"
)

type Environment struct {
	valueMap  map[string]interface{}
	enclosing *Environment
}

func NewEnvironment(e *Environment) *Environment {
	// e is nil means top level
	return &Environment{
		valueMap:  make(map[string]interface{}),
		enclosing: e,
	}
}

func (e *Environment) define(name string, val any) error {
	if e.valueMap == nil {
		e.valueMap = make(map[string]interface{})
	}
	if _, ok := e.valueMap[name]; ok {
		return fmt.Errorf("re-declare variable %s", name)
	}
	e.valueMap[name] = val
	return nil
}

func (e *Environment) get(name syntax.Token) (interface{}, error) {
	val, ok := e.valueMap[name.Lexeme]
	if ok {
		return val, nil
	}
	if e.enclosing != nil {
		return e.enclosing.get(name)
	}
	return nil, fmt.Errorf("undefined variable '%s'", name.Lexeme)
}

func (e *Environment) assign(name syntax.Token, value any) error {
	if _, ok := e.valueMap[name.Lexeme]; ok {
		e.valueMap[name.Lexeme] = value
		return nil
	}
	if e.enclosing != nil {
		return e.enclosing.assign(name, value)
	}
	return fmt.Errorf("undefined variable '%s'", name.Lexeme)
}

func (e *Environment) assignAt(distance int, name syntax.Token, value any) error {
	return e.ancestor(distance).assign(name, value)
}

func (e *Environment) getAt(distance int, name syntax.Token) (interface{}, error) {
	return e.ancestor(distance).get(name)
}

func (e *Environment) ancestor(distance int) *Environment {
	env := e
	for i := 0; i < distance; i++ {
		env = env.enclosing
	}
	return env
}
