package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// The runtime returns a borrowed slot snapshot only after validating its physical storage.
// Lowering must supply a complete runtime contract before using this for an admitted cast.
func (e *emitter) viewField(property ir.Property) string {
	of := property.Type()
	object := e.value(property.Object)
	slot := e.temporary()
	names := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.Object: "object", ir.Array: "array", ir.Map: "Map", ir.Closure: "function", ir.MaybeNumber: "number | undefined", ir.MaybeBoolean: "boolean | undefined"}
	name, supported := names[of]
	if !supported || property.Method {
		panic("compiler bug: incomplete checked field contract")
	}
	if property.ViewType != "" && (property.Of != ir.Closure || property.ViewContract != 0 && e.program.ViewContracts[property.ViewContract-1].Result != 0) {
		name = property.ViewType
	}
	if property.Absent || property.Optional || of == ir.MaybeNumber || of == ir.MaybeBoolean {
		e.line("adamic_value %s = adamic_object_optional_view(%s, %s, &%s, %d, %s, %s, %t, %t);", slot, object, cString(property.Name), e.cache(), of, cString(name), cString(property.View), property.Absent, property.Optional)
	} else {
		e.line("adamic_value %s = adamic_object_view(%s, %s, &%s, %d, %s, %s);", slot, object, cString(property.Name), e.cache(), of, cString(name), cString(property.View))
	}
	value := unslotted(of, fmt.Sprintf("%s.%s", slot, member(of)))
	if of == ir.MaybeBoolean {
		value = fmt.Sprintf("((adamic_maybe_boolean){%s.reference != NULL, %s.reference != NULL && ((adamic_boolean_box *)%s.reference)->boolean})", slot, slot, slot)
	}
	if len(property.ViewAllowed) != 0 {
		tests := []string{}
		for _, allowed := range property.ViewAllowed {
			actual := value
			of := property.Type()
			if of == ir.MaybeNumber || of == ir.MaybeBoolean {
				actual = "(" + value + ").number"
				of = ir.Number
				if property.Type() == ir.MaybeBoolean {
					actual = "(" + value + ").boolean"
					of = ir.Boolean
				}
			}
			tests = append(tests, e.binary(ir.Equal, of, actual, e.value(allowed)))
		}
		literalOf, literalValue := of, slot
		if of == ir.MaybeNumber {
			literalOf = ir.Number
			literalValue = "((adamic_value){.number = (" + value + ").number})"
		}
		if of == ir.MaybeBoolean {
			literalOf = ir.Boolean
			literalValue = "((adamic_value){.boolean = (" + value + ").boolean})"
		}
		e.line("if (!(%s)) adamic_view_literal_failure(%s, %s, %d, %s);", func() string {
			test := strings.Join(tests, " || ")
			if of == ir.MaybeNumber || of == ir.MaybeBoolean {
				return "!(" + value + ").present || (" + test + ")"
			}
			if property.Absent && of.IsReference() {
				return value + " == NULL || (" + test + ")"
			}
			return test
		}(), cString(property.View), cString(name), literalOf, literalValue)
	}
	if of == ir.Object {
		if property.Optional || property.Absent {
			e.line("if (%s != NULL) {", value)
			e.viewObjectUnion(property, value)
			e.line("}")
		} else {
			e.viewObjectUnion(property, value)
		}
	}
	if of == ir.Closure && property.ViewContract != 0 && e.program.ViewContracts[property.ViewContract-1].Result != 0 {
		callable := e.temporary()
		e.line("adamic_closure *%s = (adamic_closure *)%s;", callable, value)
		value = e.emitViewCallableCertificate(property, callable)
	}
	if of.IsReference() {
		return e.own(of, fmt.Sprintf("adamic_retain((%s)%s)", cType(of), value))
	}
	return e.snapshot(of, value)
}
