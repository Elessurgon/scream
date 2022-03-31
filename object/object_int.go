package object

import (
	"fmt"
	"sort"
	"strings"
)

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i *Integer) Type() Type {
	return INTEGER_OBJ
}
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (i *Integer) InvokeMethod(method string, env Environment, args ...Object) Object {
	if method == "chr" {
		return &String{Value: string(rune(i.Value))}
	}
	if method == "methods" {
		static := []string{"chr", "methods"}
		dynamic := env.Names("integer.")

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

func (i *Integer) ToInterface() interface{} {
	return i.Value
}
