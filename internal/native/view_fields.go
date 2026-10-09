package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// The runtime returns a borrowed slot snapshot only after validating its physical storage.
// Lowering must supply a complete runtime contract before using this for an admitted cast.
type checkedViewReadHook func(*emitter, ir.Property) (string, bool)

var checkedViewReadAdapter checkedViewReadHook

func (e *emitter) viewField(property ir.Property) string {
	if property.Namespace {
		return e.namespaceField(property)
	}
	if checkedViewReadAdapter != nil {
		if value, handled := checkedViewReadAdapter(e, property); handled {
			return value
		}
	}
	if property.Unset {
		return e.placeholderViewField(property)
	}
	return e.viewFieldObject(property, e.value(property.Object))
}

func (e *emitter) viewFieldObject(property ir.Property, object string) string {
	if property.Absent && (property.Of == ir.MaybeNumber || property.Of == ir.MaybeBoolean || property.Of == ir.String) {
		return e.optionalViewField(property, object)
	}
	slot := e.temporary()
	names := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.Object: "object", ir.Array: "array", ir.Map: "Map", ir.Closure: "function"}
	name, supported := names[property.Of]
	if !supported || property.Optional || property.Absent || property.Method {
		panic("compiler bug: incomplete checked field contract")
	}
	if property.ViewType != "" {
		name = property.ViewType
	}
	e.line("adamic_value %s = adamic_object_view(%s, %s, &%s, %d, %s, %s);", slot, object, cString(property.Name), e.cache(), property.Of, cString(name), cString(property.View))
	value := unslotted(property.Of, fmt.Sprintf("%s.%s", slot, member(property.Of)))
	if len(property.ViewAllowed) != 0 {
		tests := []string{}
		for _, allowed := range property.ViewAllowed {
			tests = append(tests, e.binary(ir.Equal, property.Of, value, e.value(allowed)))
		}
		e.line("if (!(%s)) adamic_view_literal_failure(%s, %s, %d, %s);", strings.Join(tests, " || "), cString(property.View), cString(name), property.Of, slot)
	}
	if property.Of.IsReference() {
		return e.own(property.Of, fmt.Sprintf("adamic_retain((%s)%s)", cType(property.Of), value))
	}
	return e.snapshot(property.Of, value)
}

func (e *emitter) optionalViewField(property ir.Property, object string) string {
	slot := e.temporary()
	e.line("adamic_value %s = adamic_object_optional_view(%s, %s, &%s, %d, %s, %s);", slot, object, cString(property.Name), e.cache(), property.Of, cString(property.ViewType), cString(property.View))
	value := unslotted(property.Of, fmt.Sprintf("%s.%s", slot, member(property.Of)))
	if len(property.ViewAllowed) != 0 {
		present, scalar, of := value+" != NULL", value, property.Of
		if property.Of.IsMaybe() {
			present, scalar, of = "("+value+").present", "("+value+").number", property.Of.Present()
			if of == ir.Boolean {
				scalar = "(" + value + ").boolean"
			}
		}
		tests := []string{}
		for _, allowed := range property.ViewAllowed {
			tests = append(tests, e.binary(ir.Equal, of, scalar, e.value(allowed)))
		}
		e.line("if (%s && !(%s)) adamic_view_literal_failure(%s, %s, %d, (adamic_value){.%s = %s});", present, strings.Join(tests, " || "), cString(property.View), cString(property.ViewType), of, member(of), scalar)
	}
	if property.Of.IsReference() {
		return e.own(property.Of, fmt.Sprintf("adamic_retain(%s)", value))
	}
	return e.snapshot(property.Of, value)
}

// V1 records which checked field contracts permit absence. Required-field
// construction keeps its existing write path and readiness transitions.
func (e *emitter) optionalViewFieldName(name string) bool {
	for _, contract := range e.program.ViewContracts {
		for _, field := range contract.Fields {
			if field.Name == name && field.Optional {
				return true
			}
		}
	}
	return false
}
