package evaluator

import (
	"os"

	"scream/object"
)

func envFun(args ...object.Object) object.Object {

	env := os.Environ()
	newHash := make(map[object.HashKey]object.HashPair)

	for i := 1; i < len(env); i++ {

		k := &object.String{Value: env[i]}
		v := &object.String{Value: os.Getenv(env[i])}

		newHashPair := object.HashPair{Key: k, Value: v}
		newHash[k.HashKey()] = newHashPair
	}

	return &object.Hash{Pairs: newHash}
}

func getEnvFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	if args[0].Type() != object.STRING_OBJ {
		return newError("argument must be a string, got=%s",
			args[0].Type())
	}
	input := args[0].(*object.String).Value
	return &object.String{Value: os.Getenv(input)}

}

func setEnvFun(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	if args[0].Type() != object.STRING_OBJ {
		return newError("argument must be a string, got=%s",
			args[0].Type())
	}
	if args[1].Type() != object.STRING_OBJ {
		return newError("argument must be a string, got=%s",
			args[1].Type())
	}
	name := args[0].(*object.String).Value
	value := args[1].(*object.String).Value
	os.Setenv(name, value)
	return NULL
}

func init() {
	RegisterBuiltin("os.getenv",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (getEnvFun(args...))
		})
	RegisterBuiltin("os.setenv",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (setEnvFun(args...))
		})
	RegisterBuiltin("os.environment",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (envFun(args...))
		})
}
