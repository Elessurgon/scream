package object

type Regexp struct {
	Value string

	Flags string
}

func (r *Regexp) Type() Type {
	return REGEXP_OBJ
}

func (r *Regexp) Inspect() string {
	return r.Value
}

func (r *Regexp) InvokeMethod(method string, env Environment, args ...Object) Object {
	return nil
}

func (r *Regexp) ToInterface() interface{} {
	return "<REGEXP>"
}
