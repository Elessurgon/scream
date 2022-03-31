package object

import (
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type String struct {
	Value string

	offset int
}

func (s *String) Type() Type {
	return STRING_OBJ
}
func (s *String) Inspect() string {
	return s.Value
}

func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

func (s *String) InvokeMethod(method string, env Environment, args ...Object) Object {
	if method == "len" {
		return &Integer{Value: int64(utf8.RuneCountInString(s.Value))}
	}
	if method == "methods" {
		static := []string{"len", "methods", "ord", "to_i", "to_f"}
		dynamic := env.Names("string.")

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
	if method == "ord" {
		return &Integer{Value: int64(s.Value[0])}
	}
	if method == "to_i" {
		i, err := strconv.ParseInt(s.Value, 0, 64)
		if err != nil {
			i = 0
		}
		return &Integer{Value: int64(i)}
	}
	if method == "to_f" {
		i, err := strconv.ParseFloat(s.Value, 64)
		if err != nil {
			i = 0
		}
		return &Float{Value: i}
	}
	return nil
}

func (s *String) Reset() {
	s.offset = 0
}

func (s *String) Next() (Object, Object, bool) {

	if s.offset < utf8.RuneCountInString(s.Value) {
		s.offset++
		chars := []rune(s.Value)

		val := &String{Value: string(chars[s.offset-1])}

		return val, &Integer{Value: int64(s.offset - 1)}, true
	}

	return nil, &Integer{Value: 0}, false
}

func (s *String) ToInterface() interface{} {
	return s.Value
}
