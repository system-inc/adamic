package native

import (
	"fmt"
	"slices"

	"github.com/system-inc/adamic/internal/ir"
)

// Optional selection distinguishes an absent own callback from a prototype
// method. An absent own slot must not fall through to its prototype.
func (e *emitter) optionalCalleeLookup() string {
	code := fmt.Sprintf(`#include <string.h>
static adamic_closure *adamic_optional_callee(const adamic_object *object, const char *name, adamic_slot_cache *cache, %s *method, bool *invalid) {
	const adamic_accessor *accessor = adamic_accessor_find(object, name);
    if (accessor != NULL) {
        adamic_value value = adamic_accessor_get((adamic_object *)object, name);
        const adamic_heap *heap = value.reference;
        if (heap == NULL || heap == &adamic_null) return NULL;
        if (heap->kind == adamic_kind_closure) return value.reference;
        adamic_release(value.reference);
        *invalid = true;
        return NULL;
    }
    const adamic_shape *shape = object->shape;
	for (size_t index = 0; index < shape->count; index++) {
		if (strcmp(shape->names[index], name) == 0) {
			cache->shape = shape;
			cache->index = index;
			unsigned char type = adamic_object_field_types(object)[index];
            const adamic_value *slot = &object->slots[index];
            if (type == %d || type == %d) return NULL;
            if (type == 8) {
                const adamic_heap *heap = slot->reference;
                if (heap == NULL || heap->kind == adamic_kind_closure) return adamic_retain(slot->reference);
            } else if (type == 10) {
                const adamic_heap *heap = slot->reference;
                if (heap == NULL || heap == &adamic_null) return NULL;
                if (heap->kind == adamic_kind_closure) return adamic_retain(slot->reference);
            }
            *invalid = true;
            return NULL;
		}
	}
	for (size_t index = 0; shape->methods != NULL && index < shape->methods->count; index++) {
		if (strcmp(shape->methods->names[index], name) == 0) {
			*method = shape->methods->code[index];
			return NULL;
		}
	}
	return NULL;
}`, e.methodEntryType(), ir.NullRepresentation, ir.UndefinedRepresentation)
	if !slices.Contains(e.declarations, code) {
		e.declarations = append(e.declarations, code)
	}
	return "adamic_optional_callee"
}

// Receiver, closure and method entry are selected and owned before this branch.
// Only argument evaluation and invocation belong inside the present arm.
func (e *emitter) optionalSelectedCall(call ir.CallClosure, closure, receiver, method string, exactCount bool, invalid string) string {
	present := closure + " != NULL"
	if receiver != "" {
		selected := method + " != NULL"
		if e.program.ClosureConventionNeeded() {
			selected = fmt.Sprintf("(%s.counted ? %s.counted_code != NULL : %s.code != NULL)", method, method, method)
		}
		present += " || " + selected
	}
	if invalid != "" {
		present += " || " + invalid
	}
	of := call.OptionalResult
	call.Optional = false
	text, value, owned := e.asideWith(func() string {
		if call.OptionalPresent != 0 {
			e.line("%s = true;", e.localName(call.OptionalPresent-1))
		}
		value := e.callSelectedChecked(call, closure, receiver, method, exactCount, invalid)
		if of.IsMaybe() && call.Returns == of.Present() {
			return maybe(of, value)
		}
		return value
	})
	result := "0"
	if of != 0 {
		result = e.temporary()
		e.line("%s %s = %s;", cType(of), result, absent(of))
		if of.IsReference() {
			e.owned = append(e.owned, result)
		}
	}
	e.line("if (%s) {", present)
	e.out.WriteString(text)
	e.indent++
	if of == 0 {
		e.line("(void)%s;", value)
	} else if of.IsReference() {
		e.line("%s = %s;", result, retained(value))
	} else {
		e.line("%s = %s;", result, value)
	}
	for index := len(owned) - 1; index >= 0; index-- {
		e.line("adamic_release(%s);", owned[index])
	}
	e.indent--
	e.line("}")
	return result
}
