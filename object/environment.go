package object

import (
	"fmt"
	"os"
	"strings"
)

type Environment struct {
	store map[string]Object

	readonly map[string]bool

	outer *Environment

	permit []string
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	r := make(map[string]bool)
	return &Environment{store: s, readonly: r, outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewTemporaryScope(outer *Environment, keys []string) *Environment {
	env := NewEnvironment()
	env.outer = outer
	env.permit = keys
	return env
}

func (e *Environment) Names(prefix string) []string {
	var ret []string

	for key := range e.store {
		if strings.HasPrefix(key, prefix) {
			ret = append(ret, key)
		}

		if strings.HasPrefix(key, "object.") {
			ret = append(ret, key)
		}
	}
	return ret
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {

	cur := e.store[name]
	if cur != nil && e.readonly[name] {
		fmt.Printf("Attempting to modify '%s' denied; it was defined as a constant.\n", name)
		os.Exit(3)
	}

	if len(e.permit) > 0 {
		for _, v := range e.permit {
			if v == name {
				e.store[name] = val
				return val
			}
		}
		if e.outer != nil {
			return e.outer.Set(name, val)
		}
		fmt.Printf("scoping weirdness; please report a bug\n")
		os.Exit(5)
	}
	e.store[name] = val
	return val
}

func (e *Environment) SetConst(name string, val Object) Object {
	e.store[name] = val
	e.readonly[name] = true
	return val
}
