package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

type checkedViewReadHook func(*emitter, ir.Property) (string, bool)

var checkedViewReadAdapter checkedViewReadHook

func (e *emitter) checkedViewField(property ir.Property) string {
	if property.Of == ir.Closure && property.ViewContract > 0 && e.program.ViewContracts[property.ViewContract-1].Kind == ir.ViewCallable {
		return e.emitViewCallableProperty(property)
	}
	if checkedViewReadAdapter != nil {
		if value, handled := checkedViewReadAdapter(e, property); handled {
			return value
		}
	}
	expected := property.ViewType
	if expected == "" {
		expected = map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.Object: "object", ir.Array: "array", ir.Map: "Map"}[property.Of]
	}
	return fmt.Sprintf("adamicViewField(%s, %s, %s, %d, %s, [%s])", e.value(property.Object), quote(property.Name), quote(property.View), property.Of, quote(expected), e.values(property.ViewAllowed))
}
