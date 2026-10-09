package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func init() { checkedViewReadAdapter = readUnionView }

// Registration alone is not admission. Each handled read carries a complete
// contract and runs the selector before returning any payload to its caller.
func readUnionView(e *emitter, property ir.Property) (string, bool) {
	id := property.ViewContract
	if id <= 0 || int(id) > len(e.program.ViewContracts) {
		return "", false
	}
	contract := e.program.ViewContracts[id-1]
	if contract.Kind != ir.ViewUnion && !contract.FixedTuple && !(contract.Kind == ir.ViewCallable && contract.ProducerCertified) {
		return "", false
	}
	if property.Absent && (property.Of == ir.MaybeNumber || property.Of == ir.MaybeBoolean || property.Of == ir.String) {
		return e.optionalViewField(property, e.value(property.Object)), true
	}
	if contract.Unsupported != "" || property.Optional || property.Absent || property.Method {
		panic("compiler bug: unavailable checked union read")
	}
	e.declarations = append(e.declarations, "#include \"view_unions_mixed.h\"")
	e.declarations = append(e.declarations, "#include \"view_callables_contract.h\"")
	object := e.value(property.Object)
	snapshot := e.temporary()
	e.line("adamic_view_union_value %s = adamic_object_view_union_snapshot(%s, %s, &%s, %s, %s, false);", snapshot, object, cString(property.Name), e.cache(), cString(property.View), cString(contract.Name))
	// An unknown physical slot cannot be interpreted or counted as a reference.
	e.line("if (%s.kind == adamic_view_union_unknown) {", snapshot)
	e.line("(void)adamic_view_mixed_union_select(&%s, NULL, 0, NULL, NULL, %s, %s);", snapshot, cString(property.View), cString(contract.Name))
	e.line("}")
	boxed := e.temporary()
	e.line("adamic_heap *%s;", boxed)
	e.line("if (%s.kind == adamic_view_union_number) %s = adamic_box_number(%s.payload.number);", snapshot, boxed, snapshot)
	e.line("else if (%s.kind == adamic_view_union_boolean) %s = %s.payload.boolean ? (adamic_heap *)&adamic_box_true : (adamic_heap *)&adamic_box_false;", snapshot, boxed, snapshot)
	e.line("else if (%s.kind == adamic_view_union_null) %s = &adamic_null;", snapshot, boxed)
	e.line("else %s = adamic_retain(%s.payload.reference);", boxed, snapshot)
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
	e.viewUntaggedObjectUnion(property, value)
	if property.Of == ir.Union {
		return value, true
	}
	if property.Of == ir.Object || property.Of == ir.Array || property.Of == ir.Closure || property.Of == ir.String {
		kind := map[ir.Type]string{ir.Object: "adamic_view_union_object", ir.Array: "adamic_view_union_array", ir.Closure: "adamic_view_union_function", ir.String: "adamic_view_union_string"}[property.Of]
		allowUndefined := contract.Undefined && property.Of == contract.Of
		allowNull := contract.Null && property.Of == contract.Of
		e.line("if (%s.kind != %s && !(%t && %s.kind == adamic_view_union_undefined) && !(%t && %s.kind == adamic_view_union_null)) { (void)adamic_view_mixed_union_select(&%s, NULL, 0, NULL, NULL, %s, %s); }", snapshot, kind, allowUndefined, snapshot, allowNull, snapshot, snapshot, cString(property.View), cString(contract.Name))
		if allowNull {
			return fmt.Sprintf("((%s)(%s == &adamic_null ? NULL : %s))", cType(property.Of), value, value), true
		}
		return fmt.Sprintf("((%s)%s)", cType(property.Of), value), true
	}
	panic("compiler bug: narrowed scalar union view requires a checked conversion")
}
