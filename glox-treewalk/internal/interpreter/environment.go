package interpreter

import (
	"fmt"

	"github.com/littlekuo/glox-treewalk/internal/syntax"
)

type Environment struct {
	values    []interface{}          // valid for local scope
	valueMap  map[string]interface{} // valid for global scope
	enclosing *Environment
}

func NewEnvironment(e *Environment) *Environment {
	if e == nil {
		// e is nil means top level
		return &Environment{
			valueMap: make(map[string]interface{}),
		}
	}
	return &Environment{
		values:    make([]interface{}, 0),
		enclosing: e,
	}
}

// define in global scope
func (e *Environment) defineGlobal(name string, val any) error {
	if e.valueMap == nil {
		panic("valueMap is nil")
	}
	if _, ok := e.valueMap[name]; ok {
		return fmt.Errorf("re-define variable %s", name)
	}
	e.valueMap[name] = val
	return nil
}

// get in global scope
func (e *Environment) getGlobal(name syntax.Token) (interface{}, error) {
	val, ok := e.valueMap[name.Lexeme]
	if ok {
		return val, nil
	}
	return nil, fmt.Errorf("undefined variable '%s'", name.Lexeme)
}

// assign in global scope
func (e *Environment) assignGlobal(name syntax.Token, value any) error {
	if _, ok := e.valueMap[name.Lexeme]; ok {
		e.valueMap[name.Lexeme] = value
		return nil
	}
	return fmt.Errorf("undefined variable '%s'", name.Lexeme)
}

// define in local scope
func (e *Environment) defineLocal(idx int, value any) error {
	if idx >= len(e.values) {
		e.values = append(e.values, make([]interface{}, idx+1-len(e.values))...)
	}
	e.values[idx] = value
	return nil
}

func (e *Environment) getLocal(idx int) (interface{}, error) {
	if idx >= len(e.values) {
		return nil, fmt.Errorf("undefined idx %d", idx)
	}
	return e.values[idx], nil
}

func (e *Environment) assignLocal(idx int, value any) error {
	if idx >= len(e.values) {
		return fmt.Errorf("undefined idx %d", idx)
	}
	e.values[idx] = value
	return nil
}

func (e *Environment) assignAt(distance int, idx int, value any) error {
	return e.ancestor(distance).assignLocal(idx, value)
}

func (e *Environment) getAt(distance int, idx int) (interface{}, error) {
	return e.ancestor(distance).getLocal(idx)
}

func (e *Environment) ancestor(distance int) *Environment {
	env := e
	for i := 0; i < distance; i++ {
		env = env.enclosing
	}
	return env
}
