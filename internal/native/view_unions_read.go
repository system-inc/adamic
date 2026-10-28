package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func init() { checkedViewReadAdapter = readUnionViewAt }

// Registration alone is not admission. Each handled read carries a complete
// contract and runs the selector before returning any payload to its caller.
func readUnionViewAt(e *emitter, property ir.Property, object string) (string, bool) {
	id := property.ViewContract
	if id <= 0 || int(id) > len(e.program.ViewContracts) {
		return "", false
	}
	contract := e.program.ViewContracts[id-1]
	if contract.Kind != ir.ViewUnion && !contract.FixedTuple && !(contract.Kind == ir.ViewCallable && contract.ProducerCertified) {
		return "", false
	}
	// An absent optional field reads undefined, so its declared union has to hold undefined. A
	// callable's recorded signature has no undefined to stand for.
	if contract.Unsupported != "" || property.Absent && !contract.Undefined || (property.Optional || property.Absent) && contract.Of == ir.Closure || property.Method {
		panic("compiler bug: unavailable checked union read")
	}
	e.declarations = append(e.declarations, "#include \"view_unions_mixed.h\"")
	e.declarations = append(e.declarations, "#include \"view_callables_contract.h\"")
	if object == "" {
		object = e.value(property.Object)
	}
	snapshot := e.temporary()
	if property.Optional {
		// ?. reads undefined from an absent receiver before any slot is looked at. A present one is
		// checked as any other read is.
		receiver := e.temporary()
		e.line("%s %s = %s;", cType(property.Object.Type()), receiver, object)
		e.line("adamic_view_union_value %s = %s == NULL ? (adamic_view_union_value){adamic_view_union_undefined, {.reference = NULL}} : adamic_object_view_union_snapshot(%s, %s, &%s, %s, %s, %t);", snapshot, receiver, receiver, cString(property.Name), e.cache(), cString(property.View), cString(contract.Name), property.Absent)
		object = receiver
	} else {
		e.line("adamic_view_union_value %s = adamic_object_view_union_snapshot(%s, %s, &%s, %s, %s, %t);", snapshot, object, cString(property.Name), e.cache(), cString(property.View), cString(contract.Name), property.Absent)
	}
	// An unknown physical slot cannot be interpreted or counted as a reference.
	e.line("if (%s.kind == adamic_view_union_unknown) {", snapshot)
	e.line("(void)adamic_view_mixed_union_select(&%s, NULL, 0, NULL, NULL, %s, %s);", snapshot, cString(property.View), cString(contract.Name))
	e.line("}")
	boxed := e.temporary()
	e.line("adamic_heap *%s;", boxed)
	e.line("if (%s.kind == adamic_view_union_number) %s = adamic_box_number(%s.payload.number);", snapshot, boxed, snapshot)
	e.line("else if (%s.kind == adamic_view_union_boolean) %s = %s.payload.boolean ? (adamic_heap *)&adamic_box_true : (adamic_heap *)&adamic_box_false;", snapshot, boxed, snapshot)
	e.line("else if (%s.kind == adamic_view_union_null) %s = &adamic_null;", snapshot, boxed)
	e.line("else if (%s.kind >= adamic_view_union_string && %s.kind <= adamic_view_union_function) %s = adamic_retain(%s.payload.reference);", snapshot, snapshot, boxed, snapshot)
	e.line("else %s = NULL;", boxed)
	value := e.own(ir.Union, boxed)
	if contract.Of == ir.Closure {
		recorded := e.temporary()
		e.line("const adamic_callable_signature *%s = NULL;", recorded)
		for _, producer := range e.unionCallableProducers() {
			e.line("if (%s.kind == adamic_view_union_function && %s) %s = %s;", snapshot, e.unionClosureCodeIdentity("((adamic_closure *)"+value+")", producer.Function), recorded, producer.Signature)
		}
		expected := e.untaggedCallableUnionExpected(property, recorded, e.unionCallableExpected(property))
		e.certifyUntaggedCallableRecorded(property, "((adamic_closure *)"+value+")", recorded, expected)
		e.line("(void)adamic_view_callable_shape(%s, %s, %s, %s, false);", value, recorded, expected, cString(property.View))
		return fmt.Sprintf("((adamic_closure *)%s)", value), true
	}
	if property.Of.IsMaybe() {
		e.checkMaybeUnionView(property, object, value, snapshot)
		kind, payload := "adamic_view_union_number", "number"
		if property.Of == ir.MaybeBoolean {
			kind, payload = "adamic_view_union_boolean", "boolean"
		}
		// Select the payload only after checking its kind. A missing value has
		// its own presence bit, distinct from a present zero or false.
		return e.snapshot(property.Of, fmt.Sprintf("(%s.kind == %s ? %s : %s)", snapshot, kind, maybe(property.Of, snapshot+".payload."+payload), zero(property.Of))), true
	}
	// Membership is the guarantee: it stops a value of the wrong kind and an object of the wrong
	// shape alike. The kind check below is the cast's backstop, so a value read as one
	// representation is that representation even where membership has nothing to tell apart.
	e.presentReceiver(property, object, func() { e.viewUntaggedObjectUnion(property, value) })
	if property.Of == ir.Union {
		return value, true
	}
	if property.Of == ir.Object || property.Of == ir.Array || property.Of == ir.Closure || property.Of == ir.String {
		kind := map[ir.Type]string{ir.Object: "adamic_view_union_object", ir.Array: "adamic_view_union_array", ir.Closure: "adamic_view_union_function", ir.String: "adamic_view_union_string"}[property.Of]
		allowNull := contract.Null && property.Of == contract.Of
		e.line("if (%s.kind != %s && !(%s && %s.kind == adamic_view_union_undefined) && !(%t && %s.kind == adamic_view_union_null)) { (void)adamic_view_mixed_union_select(&%s, NULL, 0, NULL, NULL, %s, %s); }", snapshot, kind, allowUndefined(contract, property, object), snapshot, allowNull, snapshot, snapshot, cString(property.View), cString(contract.Name))
		if allowNull {
			return fmt.Sprintf("((%s)(%s == &adamic_null ? NULL : %s))", cType(property.Of), value, value), true
		}
		return fmt.Sprintf("((%s)%s)", cType(property.Of), value), true
	}
	panic("compiler bug: narrowed scalar union view requires a checked conversion")
}

// Membership proves the declared union; the kind check also protects a read
// whose checker representation has narrowed that union to one maybe scalar.
func (e *emitter) checkMaybeUnionView(property ir.Property, object, value, snapshot string) {
	e.presentReceiver(property, object, func() { e.viewUntaggedObjectUnion(property, value) })
	contract := e.program.ViewContracts[property.ViewContract-1]
	kind := "adamic_view_union_number"
	if property.Of == ir.MaybeBoolean {
		kind = "adamic_view_union_boolean"
	}
	e.line("if (%s.kind != %s && !(%s && %s.kind == adamic_view_union_undefined)) { (void)adamic_view_mixed_union_select(&%s, NULL, 0, NULL, NULL, %s, %s); }", snapshot, kind, allowUndefined(contract, property, object), snapshot, snapshot, cString(property.View), cString(contract.Name))
}

// allowUndefined says, in C, when a read may give undefined: when the contract holds it for the
// representation read, or, through ?., when the receiver is absent.
func allowUndefined(contract ir.ViewContract, property ir.Property, object string) string {
	allowed := fmt.Sprint(contract.Undefined && property.Of == contract.Of)
	if property.Optional {
		return "(" + allowed + " || " + object + " == NULL)"
	}
	return allowed
}

// presentReceiver emits check only for a receiver that is there: ?. through an absent one has read
// undefined, which a contract without undefined would refuse.
func (e *emitter) presentReceiver(property ir.Property, object string, check func()) {
	if !property.Optional {
		check()
		return
	}
	e.line("if (%s != NULL) {", object)
	check()
	e.line("}")
}
