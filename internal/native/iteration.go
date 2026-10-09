package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// A protocol binding owns its receiver and cached own closure. A prototype
// method entry is immutable code, stored after the two ordinary captured cells.
// The normal closure destructor releases both cells and the whole allocation.
func (e *emitter) iteratorMethod(expression ir.IteratorMethod) string {
	object := e.value(expression.Object)
	e.temporaries++
	name := fmt.Sprintf("adamic_protocol_%d", e.temporaries)
	entryType := e.methodEntryType()
	methodCall := "method(object, NULL)"
	if e.program.ClosureConventionNeeded() {
		methodCall = "adamic_method_call(method, object, NULL, 0)"
	}
	closureCall := "adamic_closure_call(fn, NULL, 0)"
	if e.program.ClosureReceiversNeeded() {
		closureCall = "adamic_closure_receiver_call(fn, object, NULL, 0, 0)"
	}
	e.declarations = append(e.declarations, fmt.Sprintf(`#include <string.h>
static adamic_value %s_call(adamic_closure *self, adamic_value *arguments) {
 (void)arguments;
 adamic_object *object = self->cells[0]->value.reference;
 adamic_closure *fn = self->cells[1]->value.reference;
 %s method = *(%s *)(void *)(self->cells + 2);
 if (fn != NULL) return %s;
 return %s;
}
static adamic_closure *%s_bind(adamic_object *object) {
 static adamic_slot_cache cache;
 %s method = {0};
 bool present = adamic_accessor_find(object, %s) != NULL;
 for (size_t i = 0; i < object->shape->count && !present; i++) present = strcmp(object->shape->names[i], %s) == 0;
 for (size_t i = 0; object->shape->methods != NULL && i < object->shape->methods->count && !present; i++) present = strcmp(object->shape->methods->names[i], %s) == 0;
 if (!present && %t) return NULL;
 adamic_closure *fn;
 if (adamic_accessor_find(object, %s) != NULL) {
  fn = adamic_accessor_get(object, %s).reference;
  if (adamic_thrown != NULL) { adamic_release(fn); return NULL; }
 } else { fn = adamic_retain(adamic_object_callee(object, %s, &cache, &method)); }
 if (fn == NULL && %s) return NULL;
 adamic_closure *bound = adamic_allocate(sizeof *bound + 2 * sizeof bound->cells[0] + sizeof method, adamic_kind_closure);
 adamic_heap heap = bound->heap;
 *bound = (adamic_closure){.heap = heap, .code = %s_call, .count = 2};
 bound->cells[0] = adamic_cell_new((adamic_value){.reference = adamic_retain(object)}, true);
 bound->cells[1] = adamic_cell_new((adamic_value){.reference = fn}, true);
 *(%s *)(void *)(bound->cells + 2) = method;
 return bound;
}`, name, entryType, entryType, closureCall, methodCall, name, entryType,
		cString(expression.Name), cString(expression.Name), cString(expression.Name), expression.Optional,
		cString(expression.Name), cString(expression.Name), cString(expression.Name),
		map[bool]string{true: "method.code == NULL", false: "method == NULL"}[e.program.ClosureConventionNeeded()], name, entryType))
	result := e.own(ir.Closure, fmt.Sprintf("%s_bind(%s)", name, object))
	e.closureThrown()
	return result
}

func (e *emitter) iteratorField(expression ir.IteratorField) string {
	object := e.value(expression.Object)
	result := e.temporary()
	e.line("adamic_value %s = {.%s = %s};", result, member(expression.Of), slotted(expression.Of, zero(expression.Of)))
	e.line("if (adamic_accessor_find(%s, %s) != NULL) {", object, cString(expression.Name))
	e.indent++
	e.line("%s = adamic_accessor_get(%s, %s);", result, object, cString(expression.Name))
	e.closureThrown()
	e.indent--
	e.line("} else {")
	e.indent++
	slot := e.temporary()
	e.line("adamic_value *%s = adamic_object_optional_field(%s, %s, &%s);", slot, object, cString(expression.Name), e.cache())
	e.line("if (%s != NULL) %s = *%s;", slot, result, slot)
	if expression.Of.IsReference() {
		e.line("adamic_retain(%s.reference);", result)
	}
	e.indent--
	e.line("}")
	value := unslotted(expression.Of, result+"."+member(expression.Of))
	if expression.Of.IsReference() {
		e.owned = append(e.owned, result+".reference")
		return fmt.Sprintf("(%s)%s", cType(expression.Of), value)
	}
	return value
}
