package object

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() Type {
	return RETURN_VALUE_OBJ
}
func (rv *ReturnValue) Inspect() string {
	return rv.Value.Inspect()
}

func (rv *ReturnValue) InvokeMethod(method string, env Environment, args ...Object) Object {
	return nil
}

func (rv *ReturnValue) ToInterface() interface{} {
	return "<RETURN_VALUE>"
}
