package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// The runtime returns a borrowed slot snapshot only after validating its physical storage.
// Lowering must supply a complete runtime contract before using this for an admitted cast.
func (e *emitter) viewField(property ir.Property) string {
	object := e.value(property.Object)
	slot := e.temporary()
	names := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.Object: "object", ir.Array: "array", ir.Map: "Map"}
	name, supported := names[property.Of]
	if !supported || property.Optional || property.Absent || property.Method {
		panic("compiler bug: incomplete checked field contract")
	}
	e.line("adamic_value %s = adamic_object_view(%s, %s, &%s, %d, %s, %s);", slot, object, cString(property.Name), e.cache(), property.Of, cString(name), cString(property.View))
	value := unslotted(property.Of, fmt.Sprintf("%s.%s", slot, member(property.Of)))
	if property.Of.IsReference() {
		return e.own(property.Of, fmt.Sprintf("adamic_retain((%s)%s)", cType(property.Of), value))
	}
	return e.snapshot(property.Of, value)
}
