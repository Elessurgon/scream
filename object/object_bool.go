package object

import (
	"fmt"
	"sort"
	"strings"
)

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() Type {
	return BOOLEAN_OBJ
}
func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%t", b.Value)
}

func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	} else {
		value = 0
	}
	return HashKey{Type: b.Type(), Value: value}
}

func (b *Boolean) InvokeMethod(method string, env Environment, args ...Object) Object {
	if method == "methods" {
		static := []string{"methods"}
		dynamic := env.Names("bool.")

		var names []string
		names = append(names, static...)
		for _, e := range dynamic {
			bits := strings.Split(e, ".")
			names = append(names, bits[1])
		}
		sort.Strings(names)

		result := make([]Object, len(names))
		for i, txt := range names {
			result[i] = &String{Value: txt}
		}
		return &Array{Elements: result}
	}
	return nil
}

func (b *Boolean) ToInterface() interface{} {
	return b.Value
}
