package object

type BuiltinFunction func(env *Environment, args ...Object) Object
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() Type {
	return BUILTIN_OBJ
}

func (b *Builtin) Inspect() string {
	return "builtin function"
}

func (b *Builtin) InvokeMethod(method string, env Environment, args ...Object) Object {
	if method == "methods" {
		names := []string{"methods"}

		result := make([]Object, len(names))
		for i, txt := range names {
			result[i] = &String{Value: txt}
		}
		return &Array{Elements: result}
	}
	return nil
}

func (b *Builtin) ToInterface() interface{} {
	return "<BUILTIN>"
}
