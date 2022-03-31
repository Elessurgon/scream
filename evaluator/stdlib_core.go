package evaluator

import (
	"fmt"
	"os"
	"regexp"
	"scream/lexer"
	"scream/object"
	"scream/parser"
	"strconv"
	"strings"
	"unicode/utf8"
)

func chmodFun(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2",
			len(args))
	}

	path := args[0].Inspect()
	mode := ""

	switch args[1].(type) {
	case *object.String:
		mode = args[1].(*object.String).Value
	default:
		return newError("Second argument must be string, got %v", args[1])
	}

	result, err := strconv.ParseInt(mode, 8, 64)
	if err != nil {
		return &object.Boolean{Value: false}
	}

	err = os.Chmod(path, os.FileMode(result))
	if err != nil {
		return &object.Boolean{Value: false}
	}
	return &object.Boolean{Value: true}
}

func evalFun(env *object.Environment, args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch args[0].(type) {
	case *object.String:
		txt := args[0].(*object.String).Value

		l := lexer.New(txt)

		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) == 0 {

			return (Eval(program, env))
		}

		fmt.Printf("Error parsing eval-string: %s", txt)
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		os.Exit(1)
	}
	return newError("argument to `eval` not supported, got=%s",
		args[0].Type())
}

func exitFun(args ...object.Object) object.Object {

	code := 0

	if len(args) > 0 {
		switch arg := args[0].(type) {
		case *object.Integer:
			code = int(arg.Value)
		case *object.Float:
			code = int(arg.Value)
		}
	}

	os.Exit(code)
	return NULL
}

func intFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch args[0].(type) {
	case *object.String:
		input := args[0].(*object.String).Value
		i, err := strconv.Atoi(input)
		if err == nil {
			return &object.Integer{Value: int64(i)}
		}
		return newError("Converting string '%s' to int failed %s", input, err.Error())

	case *object.Boolean:
		input := args[0].(*object.Boolean).Value
		if input {
			return &object.Integer{Value: 1}

		}
		return &object.Integer{Value: 0}
	case *object.Integer:

		return args[0]
	case *object.Float:
		input := args[0].(*object.Float).Value
		return &object.Integer{Value: int64(input)}
	default:
		return newError("argument to `int` not supported, got=%s",
			args[0].Type())
	}
}

func lenFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch arg := args[0].(type) {
	case *object.String:
		return &object.Integer{Value: int64(utf8.RuneCountInString(arg.Value))}
	case *object.Null:
		return &object.Integer{Value: 0}
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}
	default:
		return newError("argument to `len` not supported, got=%s",
			args[0].Type())
	}
}

func matchFun(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2",
			len(args))
	}

	if args[0].Type() != object.STRING_OBJ {
		return newError("argument to `match` must be STRING, got %s",
			args[0].Type())
	}
	if args[1].Type() != object.STRING_OBJ {
		return newError("argument to `match` must be STRING, got %s",
			args[1].Type())
	}

	reg := regexp.MustCompile(args[0].(*object.String).Value)
	res := reg.FindStringSubmatch(args[1].(*object.String).Value)

	if len(res) > 0 {

		newHash := make(map[object.HashKey]object.HashPair)

		if len(res) > 1 {
			for i := 1; i < len(res); i++ {

				k := &object.Integer{Value: int64(i - 1)}
				v := &object.String{Value: res[i]}

				newHashPair := object.HashPair{Key: k, Value: v}
				newHash[k.HashKey()] = newHashPair

			}
		}

		return &object.Hash{Pairs: newHash}
	}

	return NULL
}

func mkdirFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}

	if args[0].Type() != object.STRING_OBJ {
		return newError("argument to `mkdir` must be STRING, got %s",
			args[0].Type())
	}

	path := args[0].(*object.String).Value

	// Can't fail?
	mode, err := strconv.ParseInt("755", 8, 64)
	if err != nil {
		return &object.Boolean{Value: false}
	}

	err = os.MkdirAll(path, os.FileMode(mode))
	if err != nil {
		return &object.Boolean{Value: false}
	}
	return &object.Boolean{Value: true}

}

func openFun(args ...object.Object) object.Object {

	path := ""
	mode := "r"
	if len(args) < 1 {
		return newError("wrong number of arguments. got=%d, want=1+",
			len(args))
	}

	switch args[0].(type) {
	case *object.String:
		path = args[0].(*object.String).Value
	default:
		return newError("argument to `file` not supported, got=%s",
			args[0].Type())

	}

	if len(args) > 1 {
		switch args[1].(type) {
		case *object.String:
			mode = args[1].(*object.String).Value
		default:
			return newError("argument to `file` not supported, got=%s",
				args[0].Type())

		}
	}

	file := &object.File{Filename: path}
	file.Open(mode)
	return (file)
}

func pragmaFun(args ...object.Object) object.Object {

	if len(args) > 1 {
		return newError("wrong number of arguments. got=%d, want=0|1",
			len(args))
	}

	if len(args) == 1 {
		switch args[0].(type) {
		case *object.String:
			input := args[0].(*object.String).Value
			input = strings.ToLower(input)

			if strings.HasPrefix(input, "no-") {
				real := strings.TrimPrefix(input, "no-")
				delete(PRAGMAS, real)
			} else {
				PRAGMAS[input] = 1
			}
		default:
			return newError("argument to `pragma` not supported, got=%s",
				args[0].Type())
		}
	}

	len := len(PRAGMAS)

	array := make([]object.Object, len)

	i := 0
	for key := range PRAGMAS {
		array[i] = &object.String{Value: key}
		i++

	}
	return &object.Array{Elements: array}
}

func pushFun(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `push` must be ARRAY, got=%s",
			args[0].Type())
	}
	arr := args[0].(*object.Array)
	length := len(arr.Elements)
	newElements := make([]object.Object, length+1)
	copy(newElements, arr.Elements)
	newElements[length] = args[1]
	return &object.Array{Elements: newElements}
}

func putsFun(args ...object.Object) object.Object {
	for _, arg := range args {
		fmt.Print(arg.Inspect())
	}
	return NULL
}

func printfFun(args ...object.Object) object.Object {

	out := sprintfFun(args...)

	if out.Type() == object.STRING_OBJ {
		fmt.Print(out.(*object.String).Value)

	}

	return NULL
}

func sprintfFun(args ...object.Object) object.Object {

	if len(args) < 1 {
		return &object.Null{}
	}

	if args[0].Type() != object.STRING_OBJ {
		return &object.Null{}
	}

	fs := args[0].(*object.String).Value

	// Convert the arguments to something go's sprintf
	// code will understand.
	argLen := len(args)
	fmtArgs := make([]interface{}, argLen-1)

	for i, v := range args[1:] {
		fmtArgs[i] = v.ToInterface()
	}

	out := fmt.Sprintf(fs, fmtArgs...)

	return &object.String{Value: out}
}

func statFun(args ...object.Object) object.Object {

	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	path := args[0].Inspect()
	info, err := os.Stat(path)

	res := make(map[object.HashKey]object.HashPair)
	if err != nil {
		return &object.Hash{Pairs: res}
	}

	sizeData := &object.Integer{Value: info.Size()}
	sizeKey := &object.String{Value: "size"}
	sizeHash := object.HashPair{Key: sizeKey, Value: sizeData}
	res[sizeKey.HashKey()] = sizeHash

	mtimeData := &object.Integer{Value: info.ModTime().Unix()}
	mtimeKey := &object.String{Value: "mtime"}
	mtimeHash := object.HashPair{Key: mtimeKey, Value: mtimeData}
	res[mtimeKey.HashKey()] = mtimeHash

	permData := &object.String{Value: info.Mode().String()}
	permKey := &object.String{Value: "perm"}
	permHash := object.HashPair{Key: permKey, Value: permData}
	res[permKey.HashKey()] = permHash

	m := fmt.Sprintf("%04o", info.Mode().Perm())
	modeData := &object.String{Value: m}
	modeKey := &object.String{Value: "mode"}
	modeHash := object.HashPair{Key: modeKey, Value: modeData}
	res[modeKey.HashKey()] = modeHash

	typeStr := "unknown"
	if info.Mode().IsDir() {
		typeStr = "directory"
	}
	if info.Mode().IsRegular() {
		typeStr = "file"
	}

	typeData := &object.String{Value: typeStr}
	typeKey := &object.String{Value: "type"}
	typeHash := object.HashPair{Key: typeKey, Value: typeData}
	res[typeKey.HashKey()] = typeHash

	return &object.Hash{Pairs: res}

}

func hashKeys(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	if args[0].Type() != object.HASH_OBJ {
		return newError("argument to `keys` must be HASH, got=%s",
			args[0].Type())
	}

	hash := args[0].(*object.Hash)
	ents := len(hash.Pairs)

	array := make([]object.Object, ents)

	i := 0
	for _, ent := range hash.Pairs {
		array[i] = ent.Key
		i++
	}

	return &object.Array{Elements: array}
}

func hashDelete(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2",
			len(args))
	}
	if args[0].Type() != object.HASH_OBJ {
		return newError("argument to `delete` must be HASH, got=%s",
			args[0].Type())
	}

	hash := args[0].(*object.Hash)

	key, ok := args[1].(object.Hashable)
	if !ok {
		return newError("key `delete` into HASH must be Hashable, got=%s",
			args[1].Type())
	}

	newHash := make(map[object.HashKey]object.HashPair)

	for k, v := range hash.Pairs {
		if k != key.HashKey() {
			newHash[k] = v
		}
	}
	return &object.Hash{Pairs: newHash}
}

func setFun(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("wrong number of arguments. got=%d, want=2",
			len(args))
	}
	if args[0].Type() != object.HASH_OBJ {
		return newError("argument to `set` must be HASH, got=%s",
			args[0].Type())
	}
	key, ok := args[1].(object.Hashable)
	if !ok {
		return newError("key `set` into HASH must be Hashable, got=%s",
			args[1].Type())
	}
	newHash := make(map[object.HashKey]object.HashPair)
	hash := args[0].(*object.Hash)
	for k, v := range hash.Pairs {
		newHash[k] = v
	}
	newHashKey := key.HashKey()
	newHashPair := object.HashPair{Key: args[1], Value: args[2]}
	newHash[newHashKey] = newHashPair
	return &object.Hash{Pairs: newHash}
}

func strFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}

	out := args[0].Inspect()
	return &object.String{Value: out}
}

func typeFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch args[0].(type) {
	case *object.String:
		return &object.String{Value: "string"}
	case *object.Regexp:
		return &object.String{Value: "regexp"}
	case *object.Boolean:
		return &object.String{Value: "bool"}
	case *object.Builtin:
		return &object.String{Value: "builtin"}
	case *object.File:
		return &object.String{Value: "file"}
	case *object.Array:
		return &object.String{Value: "array"}
	case *object.Function:
		return &object.String{Value: "function"}
	case *object.Integer:
		return &object.String{Value: "integer"}
	case *object.Float:
		return &object.String{Value: "float"}
	case *object.Hash:
		return &object.String{Value: "hash"}
	default:
		return newError("argument to `type` not supported, got=%s",
			args[0].Type())
	}
}

func unlinkFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}

	path := args[0].Inspect()

	err := os.Remove(path)
	if err != nil {
		return &object.Boolean{Value: false}
	}
	return &object.Boolean{Value: true}
}

func init() {
	RegisterBuiltin("chmod",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (chmodFun(args...))
		})
	RegisterBuiltin("delete",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (hashDelete(args...))
		})
	RegisterBuiltin("eval",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (evalFun(env, args...))
		})
	RegisterBuiltin("exit",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (exitFun(args...))
		})
	RegisterBuiltin("int",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (intFun(args...))
		})
	RegisterBuiltin("keys",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (hashKeys(args...))
		})
	RegisterBuiltin("LEN",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (lenFun(args...))
		})
	RegisterBuiltin("match",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (matchFun(args...))
		})
	RegisterBuiltin("mkdir",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (mkdirFun(args...))
		})
	RegisterBuiltin("pragma",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (pragmaFun(args...))
		})
	RegisterBuiltin("open",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (openFun(args...))
		})
	RegisterBuiltin("APPEND",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (pushFun(args...))
		})
	RegisterBuiltin("PRINT",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (putsFun(args...))
		})
	RegisterBuiltin("printf",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (printfFun(args...))
		})
	RegisterBuiltin("set",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (setFun(args...))
		})
	RegisterBuiltin("sprintf",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (sprintfFun(args...))
		})
	RegisterBuiltin("stat",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (statFun(args...))
		})
	RegisterBuiltin("string",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (strFun(args...))
		})
	RegisterBuiltin("type",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (typeFun(args...))
		})
	RegisterBuiltin("unlink",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (unlinkFun(args...))
		})
}
