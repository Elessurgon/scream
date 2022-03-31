package object

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type HashKey struct {
	Type Type

	Value uint64
}

type HashPair struct {
	Key Object

	Value Object
}

type Hash struct {
	Pairs  map[HashKey]HashPair
	offset int
}

func (h *Hash) Type() Type {
	return HASH_OBJ
}

func (h *Hash) Inspect() string {
	var out bytes.Buffer
	pairs := make([]string, 0)
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

func (h *Hash) InvokeMethod(method string, env Environment, args ...Object) Object {
	if method == "keys" {
		ents := len(h.Pairs)
		array := make([]Object, ents)
		i := 0
		for _, ent := range h.Pairs {
			array[i] = ent.Key
			i++
		}

		return &Array{Elements: array}
	}
	if method == "methods" {
		static := []string{"keys", "methods"}
		dynamic := env.Names("hash.")

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

func (h *Hash) Reset() {
	h.offset = 0
}

func (h *Hash) Next() (Object, Object, bool) {
	if h.offset < len(h.Pairs) {
		idx := 0

		for _, pair := range h.Pairs {
			if h.offset == idx {
				h.offset++
				return pair.Key, pair.Value, true
			}
			idx++
		}
	}

	return nil, &Integer{Value: 0}, false
}

func (h *Hash) ToInterface() interface{} {
	return "<HASH>"
}
