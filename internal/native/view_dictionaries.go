package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) dictionaryRead(property ir.Property) string {
	e.declarations = append(e.declarations, "#include \"view_dictionaries.h\"")
	kinds, ok := ir.DictionaryReadKinds(e.program, property.ViewContract)
	if !ok {
		panic("compiler bug: incomplete dictionary read contract")
	}
	mask := 0
	for _, kind := range kinds {
		mask |= 1 << map[string]int{"number": 1, "boolean": 2, "string": 3, "object": 4, "array": 5, "undefined": 9}[kind]
	}
	object := e.value(property.Object)
	key := e.value(property.DictionaryKey)
	result := e.temporary()
	e.line("adamic_view_dictionary_result %s = adamic_view_dictionary_source_read(%s, %s, %d, %d, %s, %s);", result, object, key, mask, property.ViewContract, cString(property.View), cString(property.ViewType))
	payload := result + ".value.payload"
	of := property.Of
	value := payload + "." + member(of)
	switch of {
	case ir.MaybeNumber:
		value = fmt.Sprintf("((adamic_maybe_number){%s.value.kind != adamic_view_union_undefined, %s.number})", result, payload)
	case ir.MaybeBoolean:
		value = fmt.Sprintf("((adamic_maybe_boolean){%s.value.kind != adamic_view_union_undefined, %s.boolean})", result, payload)
	case ir.Union:
		return e.own(ir.Union, fmt.Sprintf("adamic_view_dictionary_box(%s.value)", result))
	}
	if len(property.ViewAllowed) != 0 {
		tests := []string{}
		for _, allowed := range property.ViewAllowed {
			tests = append(tests, e.binary(ir.Equal, of.Present(), value, e.value(allowed)))
		}
		// Optional primitive finite contracts are currently refused by the lowerer.
		e.line("if (!(%s)) adamic_view_literal_failure(%s, %s, %d, %s);", strings.Join(tests, " || "), cString(property.View), cString(property.ViewType), of, payload)
	}
	if of == ir.Object {
		e.viewObjectUnion(property, value)
	}
	if of.IsReference() {
		return e.own(of, fmt.Sprintf("(%s)adamic_retain(%s)", cType(of), value))
	}
	return e.snapshot(of, value)
}
