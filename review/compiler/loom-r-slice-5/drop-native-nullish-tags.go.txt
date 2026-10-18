package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// A structural view observes an honest unset, but a present payload still passes
// the existing physical-kind check. The cached slot may belong to a static parent.
func (e *emitter) placeholderViewField(property ir.Property) string {
	if property.Of != ir.Union {
		panic("compiler bug: placeholder view has untagged storage")
	}
	present := property.UnsetType
	optional, absent := property.Optional, property.Absent
	object := e.value(property.Object)
	slot, owner, kind, result := e.temporary(), e.temporary(), e.temporary(), e.temporary()
	cache := e.cache()
	e.line("adamic_heap *%s = NULL;", result)
	if optional {
		e.line("if (%s != NULL) {", object)
	}
	lookup := "adamic_object_field"
	if absent {
		lookup = "adamic_object_optional_field"
	}
	e.line("adamic_value *%s = %s(%s, %s, &%s);", slot, lookup, object, cString(property.Name), cache)
	if absent {
		e.line("if (%s != NULL) {", slot)
	}
	e.line("const adamic_object *%s = NULL;", owner)
	e.line("%s = adamic_object_field_owned(%s, %s, &%s, &%s);", slot, object, cString(property.Name), cache, owner)
	e.line("unsigned char %s = adamic_object_field_types(%s)[adamic_slot_index(%s, %s)];", kind, owner, owner, slot)
	// Copy a tagged nullish payload even when its readiness bit is unset.
	e.line("if (%s == %d && (%s->reference == NULL || %s->reference == &adamic_null)) {", kind, ir.Union, slot, slot)
	e.line("%s = adamic_retain(%s->reference);", result, slot)
	unset := fmt.Sprintf("!adamic_object_initialized(%s)[adamic_slot_index(%s, %s)]", owner, owner, slot)
	unset += fmt.Sprintf(" || (%s == %d && !adamic_maybe_number_unpack(%s->number).present)", kind, ir.MaybeNumber, slot)
	unset += fmt.Sprintf(" || (%s == %d && !adamic_maybe_boolean_unpack(%s->maybe_boolean).present)", kind, ir.MaybeBoolean, slot)
	if present.IsReference() {
		unset += fmt.Sprintf(" || (%s == %d && %s->reference == NULL)", kind, present, slot)
	}
	e.line("} else if (!(%s)) {", unset)
	savedOwned := e.owned
	e.owned = nil
	property.Optional, property.Absent, property.Unset = false, false, false
	property.Of = present
	value := e.viewFieldObject(property, object)
	boxed, fresh := converted(present, ir.Union, value)
	if fresh {
		boxed = e.own(ir.Union, boxed)
		e.taken(boxed)
	} else if present.IsReference() {
		e.taken(value)
	}
	e.line("%s = %s;", result, boxed)
	for _, owned := range e.owned {
		e.line("adamic_release(%s);", owned)
	}
	e.owned = savedOwned
	e.line("}")
	if absent {
		e.line("}")
	}
	if optional {
		e.line("}")
	}
	e.hold(result)
	return result
}

func zeroPlaceholder(of ir.Type) string {
	if of.IsMaybe() {
		return zero(of)
	}
	return "NULL"
}
