package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Presence is queried before readiness and payload. A present undefined value
// is admitted only after its physical representation and readiness are checked.
func (e *emitter) optionalStringViewField(property ir.Property) string {
	object := e.value(property.Object)
	cache, slot, result := e.cache(), e.temporary(), e.temporary()
	e.line("adamic_string *%s = NULL;", result)
	e.line("if (%s == NULL) (void)adamic_object_view(%s, %s, &%s, %d, %s, %s);", object, object, cString(property.Name), cache, ir.String, cString(property.ViewType), cString(property.View))
	owner := e.temporary()
	e.line("const adamic_object *%s = %s;", owner, object)
	e.line("adamic_value *%s = adamic_object_optional_field(%s, %s, &%s);", slot, owner, cString(property.Name), cache)
	e.line("while (%s != NULL && %s->class != NULL && %s->class->is_static) {", slot, owner, owner)
	e.line("size_t %s_flag = %s->class->static_flags[%s.index];", slot, owner, cache)
	e.line("if (%s_flag == 0 || %s->slots[%s_flag - 1].number != 0 || %s->class->static_parent == 0) break;", slot, owner, slot, owner)
	e.line("%s = %s->slots[%s->class->static_parent - 1].reference;", owner, owner, owner)
	e.line("%s = adamic_object_optional_field(%s, %s, &%s);", slot, owner, cString(property.Name), cache)
	e.line("}")
	e.line("if (%s == NULL) {", slot)
	if !property.Absent {
		e.line("(void)adamic_object_view(%s, %s, &%s, %d, %s, %s);", object, cString(property.Name), cache, ir.String, cString(property.ViewType), cString(property.View))
	}
	e.line("} else {")
	e.line("(void)adamic_object_read(%s, %s, &%s, %s);", owner, cString(property.Name), cache, cString(property.View))
	e.line("unsigned char %s_type = adamic_object_field_types(%s)[%s.index];", slot, owner, cache)
	if e.viewStringUndefined(property) {
		e.line("if (!((%s_type == %d || %s_type == %d || %s_type == %d) && %s->reference == NULL)) {", slot, ir.String, slot, ir.Object, slot, ir.Union, slot)
	} else {
		e.line("{")
	}
	e.line("adamic_value %s_value = adamic_object_view(%s, %s, &%s, %d, %s, %s);", slot, owner, cString(property.Name), cache, ir.String, cString(property.ViewType), cString(property.View))
	e.line("%s = %s_value.reference;", result, slot)
	if len(property.ViewAllowed) != 0 {
		tests := []string{}
		for _, allowed := range property.ViewAllowed {
			tests = append(tests, e.binary(ir.Equal, ir.String, result, e.value(allowed)))
		}
		e.line("if (!(%s)) adamic_view_literal_failure(%s, %s, %d, %s_value);", strings.Join(tests, " || "), cString(property.View), cString(property.ViewType), ir.String, slot)
	}
	e.line("}")
	e.line("}")
	return e.own(ir.String, fmt.Sprintf("adamic_retain(%s)", result))
}

// Optional receivers short-circuit before looking for a required field.
func (e *emitter) optionalReceiverViewField(property ir.Property) string {
	object := e.value(property.Object)
	result, slot := e.temporary(), e.temporary()
	of := property.Type()
	e.line("%s %s = %s;", cType(of), result, absent(of))
	e.line("if (%s != NULL) {", object)
	e.line("adamic_value %s = adamic_object_view(%s, %s, &%s, %d, %s, %s);", slot, object, cString(property.Name), e.cache(), property.Of, cString(property.ViewType), cString(property.View))
	value := fmt.Sprintf("%s.%s", slot, member(property.Of))
	if len(property.ViewAllowed) > 0 {
		tests := []string{}
		for _, allowed := range property.ViewAllowed {
			tests = append(tests, e.binary(ir.Equal, property.Of, value, e.value(allowed)))
		}
		e.line("if (!(%s)) adamic_view_literal_failure(%s, %s, %d, %s);", strings.Join(tests, " || "), cString(property.View), cString(property.ViewType), property.Of, slot)
	}
	if property.Of == ir.Number {
		e.line("%s = (adamic_maybe_number){true, %s};", result, value)
	} else {
		e.line("%s = %s;", result, value)
	}
	e.line("}")
	if of.IsReference() {
		return e.own(of, fmt.Sprintf("adamic_retain(%s)", result))
	}
	return e.snapshot(of, result)
}
