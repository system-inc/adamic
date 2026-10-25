package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

type checkedViewReadHook func(*emitter, ir.Property, string) (string, bool)

var checkedViewReadAdapter checkedViewReadHook

func (e *emitter) checkedViewField(property ir.Property) string {
	return e.checkedViewFieldAt(property, "")
}

// checkedViewFieldAt is checkedViewField from object, the receiver already evaluated, or from
// property.Object when object is empty.
func (e *emitter) checkedViewFieldAt(property ir.Property, object string) string {
	if checkedViewReadAdapter != nil {
		if value, handled := checkedViewReadAdapter(e, property, object); handled {
			return value
		}
	}
	if object == "" {
		object = e.value(property.Object)
	}
	expected := property.ViewType
	if expected == "" {
		expected = map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.Object: "object", ir.Array: "array", ir.Map: "Map"}[property.Of]
	}
	return fmt.Sprintf("adamicViewField(%s, %s, %s, %d, %s, [%s])", object, quote(property.Name), quote(property.View), property.Of, quote(expected), e.values(property.ViewAllowed))
}

// checkedViewWrite is object.name = value through a checked view, both already evaluated.
func (e *emitter) checkedViewWrite(statement ir.SetProperty, object, value string) string {
	return fmt.Sprintf("adamicViewWrite(%s, %s, %s, %d, %s)", object, quote(statement.Name), value, viewFieldRepresentation(statement.Value), quote(statement.DeclaredType))
}
