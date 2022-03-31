package evaluator

import (
	"path/filepath"

	"scream/object"
)

func dirGlob(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	pattern := args[0].(*object.String).Value

	entries, err := filepath.Glob(pattern)
	if err != nil {
		return NULL
	}

	l := len(entries)
	result := make([]object.Object, l)
	for i, txt := range entries {
		result[i] = &object.String{Value: txt}
	}
	return &object.Array{Elements: result}
}

func init() {
	RegisterBuiltin("directory.glob",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (dirGlob(args...))
		})
}
