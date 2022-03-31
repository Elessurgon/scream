package object

import (
	"bytes"
	"sort"
	"strings"

	"scream/ast"
)

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Defaults   map[string]ast.Expression
	Env        *Environment
}

func (f *Function) Type() Type {
	return FUNCTION_OBJ
}

func (f *Function) Inspect() string {
	var out bytes.Buffer
	parameters := make([]string, 0)
	for _, p := range f.Parameters {
		parameters = append(parameters, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(parameters, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}

func (f *Function) InvokeMethod(method string, env Environment, args ...Object) Object {
	if method == "methods" {
		static := []string{"methods"}
		dynamic := env.Names("function.")

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

func (f *Function) ToInterface() interface{} {
	return "<FUNCTION>"
}
