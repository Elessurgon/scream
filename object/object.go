package object

type Type string

const (
	INTEGER_OBJ      = "INTEGER"
	FLOAT_OBJ        = "FLOAT"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	STRING_OBJ       = "STRING"
	BUILTIN_OBJ      = "BUILTIN"
	ARRAY_OBJ        = "ARRAY"
	HASH_OBJ         = "HASH"
	FILE_OBJ         = "FILE"
	REGEXP_OBJ       = "REGEXP"
)

type Object interface {
	Type() Type
	Inspect() string
	InvokeMethod(method string, env Environment, args ...Object) Object
	ToInterface() interface{}
}

type Hashable interface {
	HashKey() HashKey
}

type Iterable interface {
	Reset()
	Next() (Object, Object, bool)
}
