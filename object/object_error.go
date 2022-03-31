package object

type Error struct {
	Message string
}

func (e *Error) Type() Type {
	return ERROR_OBJ
}
func (e *Error) Inspect() string {
	return "ERROR: " + e.Message
}

func (e *Error) InvokeMethod(method string, env Environment, args ...Object) Object {

	return nil
}

func (e *Error) ToInterface() interface{} {
	return "<ERROR>"
}
