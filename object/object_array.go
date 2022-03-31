package object

import (
	"bytes"
	"sort"
	"strings"
)

type Array struct {
	Elements []Object
	offset   int
}

func (ao *Array) Type() Type {
	return ARRAY_OBJ
}
func (ao *Array) Inspect() string {
	var out bytes.Buffer
	elements := make([]string, 0)
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

func (ao *Array) InvokeMethod(method string, env Environment, args ...Object) Object {
	if method == "len" {
		return &Integer{Value: int64(len(ao.Elements))}
	}
	if method == "methods" {
		static := []string{"len", "methods"}
		dynamic := env.Names("array.")

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

func (ao *Array) Reset() {
	ao.offset = 0
}

func (ao *Array) Next() (Object, Object, bool) {
	if ao.offset < len(ao.Elements) {
		ao.offset++

		element := ao.Elements[ao.offset-1]
		return element, &Integer{Value: int64(ao.offset - 1)}, true
	}

	return nil, &Integer{Value: 0}, false
}

func (ao *Array) ToInterface() interface{} {
	return "<ARRAY>"
}
