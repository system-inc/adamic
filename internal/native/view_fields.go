package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// The runtime returns a borrowed slot snapshot only after validating its physical storage.
// Lowering must supply a complete runtime contract before using this for an admitted cast.
type checkedViewReadHook func(*emitter, ir.Property, string) (string, bool)

var checkedViewReadAdapter checkedViewReadHook

func (e *emitter) viewField(property ir.Property) string {
	return e.viewFieldAt(property, "")
}

// viewFieldAt is viewField from object, the receiver already evaluated, or from property.Object when
// object is empty.
func (e *emitter) viewFieldAt(property ir.Property, object string) string {
	if checkedViewReadAdapter != nil {
		if value, handled := checkedViewReadAdapter(e, property, object); handled {
			return value
		}
	}
	if object == "" {
		object = e.value(property.Object)
	}
	slot := e.temporary()
	names := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.Object: "object", ir.Array: "array", ir.Map: "Map"}
	name, supported := names[property.Of]
	if !supported || property.Absent || property.Method {
		panic("compiler bug: incomplete checked field contract")
	}
	if property.ViewType != "" {
		name = property.ViewType
	}
	read := fmt.Sprintf("adamic_object_view(%s, %s, &%s, %d, %s, %s)", object, cString(property.Name), e.cache(), property.Of, cString(name), cString(property.View))
	absent := ""
	if property.Optional {
		// ?. reads undefined from an absent receiver before any slot is looked at. A present one is
		// checked as any other read is.
		receiver := e.temporary()
		e.line("%s %s = %s;", cType(property.Object.Type()), receiver, object)
		absent = receiver + " == NULL"
		read = fmt.Sprintf("%s ? (adamic_value){0} : adamic_object_view(%s, %s, &%s, %d, %s, %s)", absent, receiver, cString(property.Name), e.cache(), property.Of, cString(name), cString(property.View))
	}
	e.line("adamic_value %s = %s;", slot, read)
	value := unslotted(property.Of, fmt.Sprintf("%s.%s", slot, member(property.Of)))
	if len(property.ViewAllowed) != 0 {
		tests := []string{}
		for _, allowed := range property.ViewAllowed {
			tests = append(tests, e.binary(ir.Equal, property.Of, value, e.value(allowed)))
		}
		if absent != "" {
			e.line("if (!(%s) && !(%s)) adamic_view_literal_failure(%s, %s, %d, %s);", absent, strings.Join(tests, " || "), cString(property.View), cString(name), property.Of, slot)
		} else {
			e.line("if (!(%s)) adamic_view_literal_failure(%s, %s, %d, %s);", strings.Join(tests, " || "), cString(property.View), cString(name), property.Of, slot)
		}
	}
	if property.Of.IsReference() {
		return e.own(property.Of, fmt.Sprintf("adamic_retain((%s)%s)", cType(property.Of), value))
	}
	if property.Optional {
		// Through ?., a number or a boolean is the maybe pair an absent receiver leaves empty.
		return e.snapshot(property.Type(), fmt.Sprintf("(%s ? %s : %s)", absent, zero(property.Type()), maybe(property.Type(), value)))
	}
	return e.snapshot(property.Of, value)
}
