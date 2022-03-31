package object

import (
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
)

type Float struct {
	Value float64
}

func (f *Float) Inspect() string {
	return strconv.FormatFloat(f.Value, 'f', -1, 64)
}

func (f *Float) Type() Type {
	return FLOAT_OBJ
}

func (f *Float) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(f.Inspect()))
	return HashKey{Type: f.Type(), Value: h.Sum64()}
}

func (f *Float) InvokeMethod(method string, env Environment, args ...Object) Object {
	if method == "methods" {
		static := []string{"methods"}
		dynamic := env.Names("float.")

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

func (f *Float) ToInterface() interface{} {
	return f.Value
}
