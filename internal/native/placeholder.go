package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// A structural view observes an honest unset, but a present payload still passes
// the existing physical-kind check. The cached slot may belong to a static parent.
func (nativeEmitter *emitter) placeholderViewField(property ir.Property, object string) string {
	if property.Of != ir.Union {
		panic("compiler bug: placeholder view has untagged storage")
	}
	present := property.UnsetType
	optional, absent := property.Optional, property.Absent
	if object == "" {
		object = nativeEmitter.value(property.Object)
	}
	slot, owner, kind, result := nativeEmitter.temporary(), nativeEmitter.temporary(), nativeEmitter.temporary(), nativeEmitter.temporary()
	cache := nativeEmitter.cache()
	nativeEmitter.line("adamic_heap *%s = NULL;", result)
	if optional {
		nativeEmitter.line("if (%s != NULL) {", object)
	}
	lookup := "adamic_object_field"
	if absent {
		lookup = "adamic_object_optional_field"
	}
	nativeEmitter.line("adamic_value *%s = %s(%s, %s, &%s);", slot, lookup, object, cString(property.Name), cache)
	if absent {
		nativeEmitter.line("if (%s != NULL) {", slot)
	}
	nativeEmitter.line("const adamic_object *%s = NULL;", owner)
	nativeEmitter.line(
		"%s = adamic_object_field_owned(%s, %s, &%s, &%s);",
		slot,
		object,
		cString(property.Name),
		cache,
		owner,
	)
	nativeEmitter.line(
		"unsigned char %s = adamic_object_field_types(%s)[adamic_slot_index(%s, %s)];",
		kind,
		owner,
		owner,
		slot,
	)
	// Copy a tagged nullish payload even when its readiness bit is unset.
	nativeEmitter.line(
		"if (%s == %d || %s == %d || (%s == %d && (%s->reference == NULL || %s->reference == &adamic_null))) {",
		kind,
		ir.UndefinedRepresentation,
		kind,
		ir.NullRepresentation,
		kind,
		ir.Union,
		slot,
		slot,
	)
	nativeEmitter.line("%s = adamic_retain(%s->reference);", result, slot)
	unset := fmt.Sprintf("!adamic_object_initialized(%s)[adamic_slot_index(%s, %s)]", owner, owner, slot)
	unset += fmt.Sprintf(
		" || (%s == %d && !adamic_maybe_number_unpack(%s->number).present)",
		kind,
		ir.MaybeNumber,
		slot,
	)
	unset += fmt.Sprintf(
		" || (%s == %d && !adamic_maybe_boolean_unpack(%s->maybe_boolean).present)",
		kind,
		ir.MaybeBoolean,
		slot,
	)
	if present.IsReference() {
		unset += fmt.Sprintf(" || (%s == %d && %s->reference == NULL)", kind, present, slot)
	}
	nativeEmitter.line("} else if (!(%s)) {", unset)
	// A list of the branch's own, which the statement's ladder can't name past the branch: a check in
	// it is written in place.
	savedOwned, savedLadder := nativeEmitter.owned, nativeEmitter.ladder
	nativeEmitter.owned, nativeEmitter.ladder = nil, nil
	property.Optional, property.Absent, property.Unset = false, false, false
	property.Of = present
	value := nativeEmitter.viewFieldObject(property, object)
	boxed, fresh := converted(present, ir.Union, value)
	if fresh {
		boxed = nativeEmitter.own(ir.Union, boxed)
		nativeEmitter.taken(boxed)
	} else if present.IsReference() {
		nativeEmitter.taken(value)
	}
	nativeEmitter.line("%s = %s;", result, boxed)
	for _, owned := range nativeEmitter.owned {
		nativeEmitter.line("adamic_release(%s);", owned)
	}
	nativeEmitter.owned, nativeEmitter.ladder = savedOwned, savedLadder
	nativeEmitter.line("}")
	if absent {
		nativeEmitter.line("}")
	}
	if optional {
		nativeEmitter.line("}")
	}
	nativeEmitter.hold(result)
	return result
}

func zeroPlaceholder(of ir.Type) string {
	if of.IsMaybe() {
		return zero(of)
	}
	return "NULL"
}
