package object

import (
	"fmt"
)

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]Object), outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Use to set a new variable: let a = 5
func (e *Environment) Define(name string, value Object) Object {
	e.store[name] = value
	return value
}

// Assign value to a variable
func (e *Environment) Assign(name string, value Object) (Object, error) {
	if _, isFound := e.store[name]; isFound {
		e.store[name] = value
		return value, nil
	}
	if e.outer != nil {
		return e.outer.Assign(name, value)
	}

	return nil, fmt.Errorf("Missing variable %s ", name)
}

// GET: Get value of a variable
func (e *Environment) Get(name string) (Object, bool) {
	value, isFound := e.store[name]
	if !isFound && e.outer != nil {
		value, isFound = e.outer.Get(name)
	}
	return value, isFound
}
